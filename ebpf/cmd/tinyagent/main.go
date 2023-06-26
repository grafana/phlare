package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	ebpfspy "github.com/grafana/phlare/ebpf"
	"github.com/grafana/phlare/ebpf/cmd/tinyagent/k8s"
	"github.com/grafana/phlare/ebpf/pprof"
	"github.com/grafana/phlare/ebpf/sd"
	"github.com/grafana/phlare/ebpf/symtab"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"

	"os"
	"sync"
	"time"
)

var endpointURL = flag.String("endpoint-url", "", "")
var endpointUsername = flag.String("endpoint-username", "", "")
var endpointPasswordFile = flag.String("endpoint-password-file", "", "")

var write *Write

func main() {
	flag.Parse()
	logger := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))

	if *endpointURL == "" {
		panic("endpoint-url")
	}

	write = NewWrite(*endpointURL, *endpointUsername, "", *endpointPasswordFile)
	k8sClient, err := k8s.New(os.Getenv("HOSTNAME"))
	if err != nil {
		level.Error(logger).Log("msg", "failed to create k8s client", "err", err)
	}

	arguments := getArguments()
	targetFinder, err := sd.NewTargetFinder(os.DirFS("/"), logger, targetsOptionFromArgs(arguments))
	if err != nil {
		panic(fmt.Errorf("ebpf target finder create: %w", err))
	}

	session, err := ebpfspy.NewSession(
		logger,
		targetFinder,
		sessionOptionsFromArgs(arguments),
	)
	if err != nil {
		panic(fmt.Errorf("ebpf session create: %w", err))
	}
	registry := prometheus.NewRegistry()
	component, err := New(logger, registry, arguments, session, targetFinder)
	if err != nil {
		panic(err)
	}

	var g run.Group
	g.Add(func() error {
		return component.Run(context.Background())
	}, func(err error) {

	})
	if k8sClient != nil {
		g.Add(func() error {
			return loopK8sSD(k8sClient, component)
		}, func(err error) {

		})
	}
	err = g.Run()
	if err != nil {
		println(err.Error())
		for {

		}
	}
}

func loopK8sSD(client *k8s.Client, component *Component) error {
	for {
		targets := client.GetTargetsLite()
		convertedTargets := make([]sd.DiscoveryTarget, len(targets))
		for i, target := range targets {
			convertedTargets[i] = sd.DiscoveryTarget(target)
		}
		arguments := getArguments()
		arguments.Targets = convertedTargets
		component.Update(arguments)
		time.Sleep(15 * time.Second)
	}
}

func New(logger log.Logger, register prometheus.Registerer, args Arguments, session ebpfspy.Session, targetFinder sd.TargetFinder) (*Component, error) {
	flowAppendable := NewFanout(args.ForwardTo, "", register)

	metrics := newMetrics(register)

	res := &Component{

		logger:       logger,
		metrics:      metrics,
		appendable:   flowAppendable,
		args:         args,
		targetFinder: targetFinder,
		session:      session,
		argsUpdate:   make(chan Arguments),
	}
	res.metrics.targetsActive.Set(float64(len(res.targetFinder.DebugInfo())))
	return res, nil
}

type Arguments struct {
	ForwardTo            []Appendable         `river:"forward_to,attr"`
	Targets              []sd.DiscoveryTarget `river:"targets,attr,optional"`
	DefaultTarget        sd.DiscoveryTarget   `river:"default_target,attr,optional"` // undocumented, keeping it until we have other sd
	TargetsOnly          bool                 `river:"targets_only,attr,optional"`   // undocumented, keeping it until we have other sd
	CollectInterval      time.Duration        `river:"collect_interval,attr,optional"`
	SampleRate           int                  `river:"sample_rate,attr,optional"`
	PidCacheSize         int                  `river:"pid_cache_size,attr,optional"`
	BuildIDCacheSize     int                  `river:"build_id_cache_size,attr,optional"`
	SameFileCacheSize    int                  `river:"same_file_cache_size,attr,optional"`
	ContainerIDCacheSize int                  `river:"container_id_cache_size,attr,optional"`
	CacheRounds          int                  `river:"cache_rounds,attr,optional"`
	CollectUserProfile   bool                 `river:"collect_user_profile,attr,optional"`
	CollectKernelProfile bool                 `river:"collect_kernel_profile,attr,optional"`
}

func (rc *Arguments) UnmarshalRiver(f func(interface{}) error) error {
	*rc = defaultArguments()
	type config Arguments
	return f((*config)(rc))
}

func getArguments() Arguments {
	res := defaultArguments()
	res.ForwardTo = []Appendable{write}
	host, _ := os.Hostname()
	if host == "korniltsev" || os.Getenv("HOSTNAME") == "korniltsev" {
		res.DefaultTarget = sd.DiscoveryTarget{"service_name": "tolyan-host1"}
		res.TargetsOnly = false
	}
	return res
}
func defaultArguments() Arguments {
	return Arguments{
		CollectInterval:      15 * time.Second,
		SampleRate:           97,
		PidCacheSize:         32,
		ContainerIDCacheSize: 1024,
		BuildIDCacheSize:     64,
		SameFileCacheSize:    8,
		CacheRounds:          3,
		CollectUserProfile:   true,
		CollectKernelProfile: true,
		TargetsOnly:          true,
	}
}

type Component struct {
	//options      component.Options
	args         Arguments
	argsUpdate   chan Arguments
	appendable   *Fanout
	targetFinder sd.TargetFinder
	session      ebpfspy.Session

	debugInfo     DebugInfo
	debugInfoLock sync.Mutex
	metrics       *metrics
	logger        log.Logger
}

func (c *Component) Run(ctx context.Context) error {
	err := c.session.Start()
	if err != nil {
		return fmt.Errorf("ebpf profiling session start: %w", err)
	}
	defer c.session.Stop()

	var g run.Group
	g.Add(func() error {
		collectInterval := c.args.CollectInterval
		t := time.NewTicker(collectInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case newArgs := <-c.argsUpdate:
				c.args = newArgs
				c.targetFinder.Update(targetsOptionFromArgs(c.args))
				c.metrics.targetsActive.Set(float64(len(c.targetFinder.DebugInfo())))
				err := c.session.Update(sessionOptionsFromArgs(c.args))
				if err != nil {
					return nil
				}
				c.appendable.UpdateChildren(newArgs.ForwardTo)
				if c.args.CollectInterval != collectInterval {
					t.Reset(c.args.CollectInterval)
					collectInterval = c.args.CollectInterval
				}
			case <-t.C:
				err := c.collectProfiles()
				if err != nil {
					c.metrics.profilingSessionsFailingTotal.Inc()
					return err
				}
				c.updateDebugInfo()
			}
		}
	}, func(error) {

	})
	return g.Run()
}

func (c *Component) Update(newArgs Arguments) error {
	c.argsUpdate <- newArgs
	return nil
}

func (c *Component) DebugInfo() interface{} {
	c.debugInfoLock.Lock()
	defer c.debugInfoLock.Unlock()
	return c.debugInfo
}

func (c *Component) collectProfiles() error {
	c.metrics.profilingSessionsTotal.Inc()
	level.Debug(c.logger).Log("msg", "ebpf  collectProfiles")
	args := c.args
	builders := pprof.NewProfileBuilders(args.SampleRate)
	err := c.session.CollectProfiles(func(target *sd.Target, stack []string, value uint64, pid uint32) {
		labelsHash, labels := target.Labels()
		builder := builders.BuilderForTarget(labelsHash, labels)
		builder.AddSample(stack, value)
	})

	if err != nil {
		return fmt.Errorf("ebpf session collectProfiles %w", err)
	}
	level.Debug(c.logger).Log("msg", "ebpf collectProfiles done", "profiles", len(builders.Builders))
	bytesSent := 0
	for _, builder := range builders.Builders {
		c.metrics.pprofsTotal.Inc()

		buf := bytes.NewBuffer(nil)
		_, err := builder.Write(buf)
		if err != nil {
			return fmt.Errorf("ebpf profile encode %w", err)
		}
		rawProfile := buf.Bytes()

		appender := c.appendable.Appender()
		bytesSent += len(rawProfile)
		samples := []*RawSample{{RawProfile: rawProfile}}
		err = appender.Append(context.Background(), builder.Labels, samples)
		if err != nil {
			level.Error(c.logger).Log("msg", "ebpf pprof write", "err", err)
			continue
		}
	}
	level.Debug(c.logger).Log("msg", "ebpf append done", "bytes_sent", bytesSent)
	return nil
}

type DebugInfo struct {
	Targets interface{} `river:"targets,attr,optional"`
	Session interface{} `river:"session,attr,optional"`
}

func (c *Component) updateDebugInfo() {
	c.debugInfoLock.Lock()
	defer c.debugInfoLock.Unlock()

	c.debugInfo = DebugInfo{
		Targets: c.targetFinder.DebugInfo(),
		Session: c.session.DebugInfo(),
	}
}

func targetsOptionFromArgs(args Arguments) sd.TargetsOptions {
	targets := make([]sd.DiscoveryTarget, 0, len(args.Targets))
	for _, t := range args.Targets {
		targets = append(targets, sd.DiscoveryTarget(t))
	}
	return sd.TargetsOptions{
		Targets:            targets,
		DefaultTarget:      sd.DiscoveryTarget(args.DefaultTarget),
		TargetsOnly:        args.TargetsOnly,
		ContainerCacheSize: args.ContainerIDCacheSize,
	}
}

func cacheOptionsFromArgs(args Arguments) symtab.CacheOptions {
	return symtab.CacheOptions{
		PidCacheOptions: symtab.GCacheOptions{
			Size:       args.PidCacheSize,
			KeepRounds: args.CacheRounds,
		},
		BuildIDCacheOptions: symtab.GCacheOptions{
			Size:       args.BuildIDCacheSize,
			KeepRounds: args.CacheRounds,
		},
		SameFileCacheOptions: symtab.GCacheOptions{
			Size:       args.SameFileCacheSize,
			KeepRounds: args.CacheRounds,
		},
	}
}

func sessionOptionsFromArgs(args Arguments) ebpfspy.SessionOptions {
	return ebpfspy.SessionOptions{
		CollectUser:   args.CollectUserProfile,
		CollectKernel: args.CollectKernelProfile,
		SampleRate:    args.SampleRate,
		CacheOptions:  cacheOptionsFromArgs(args),
	}
}
