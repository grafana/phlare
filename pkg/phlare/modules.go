package phlare

import (
	"context"
	"fmt"
	"net/http"
	"os"

	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/kv/codec"
	"github.com/grafana/dskit/kv/memberlist"
	"github.com/grafana/dskit/ring"
	"github.com/grafana/dskit/services"
	grpcgw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/thanos-io/thanos/pkg/discovery/dns"
	"github.com/weaveworks/common/middleware"
	"github.com/weaveworks/common/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/grafana/phlare/pkg/agent"
	"github.com/grafana/phlare/pkg/distributor"
	agentv1 "github.com/grafana/phlare/pkg/gen/agent/v1"
	"github.com/grafana/phlare/pkg/gen/agent/v1/agentv1connect"
	commonv1 "github.com/grafana/phlare/pkg/gen/common/v1"
	"github.com/grafana/phlare/pkg/gen/ingester/v1/ingesterv1connect"
	"github.com/grafana/phlare/pkg/gen/push/v1/pushv1connect"
	"github.com/grafana/phlare/pkg/gen/querier/v1/querierv1connect"
	"github.com/grafana/phlare/pkg/ingester"
	"github.com/grafana/phlare/pkg/objstore"
	objstoreclient "github.com/grafana/phlare/pkg/objstore/client"
	"github.com/grafana/phlare/pkg/objstore/providers/filesystem"
	"github.com/grafana/phlare/pkg/openapiv2"
	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
	"github.com/grafana/phlare/pkg/querier"
	"github.com/grafana/phlare/pkg/usagestats"
	"github.com/grafana/phlare/pkg/util"
	"github.com/grafana/phlare/pkg/util/build"
)

// The various modules that make up Phlare.
const (
	All          string = "all"
	Agent        string = "agent"
	Distributor  string = "distributor"
	Server       string = "server"
	Ring         string = "ring"
	Ingester     string = "ingester"
	MemberlistKV string = "memberlist-kv"
	Querier      string = "querier"
	GRPCGateway  string = "grpc-gateway"
	Storage      string = "storage"
	UsageReport  string = "usage-report"

	// RuntimeConfig            string = "runtime-config"
	// Overrides                string = "overrides"
	// OverridesExporter        string = "overrides-exporter"
	// TenantConfigs            string = "tenant-configs"
	// IngesterQuerier          string = "ingester-querier"
	// QueryFrontend            string = "query-frontend"
	// QueryFrontendTripperware string = "query-frontend-tripperware"
	// RulerStorage             string = "ruler-storage"
	// Ruler                    string = "ruler"
	// TableManager             string = "table-manager"
	// Compactor                string = "compactor"
	// IndexGateway             string = "index-gateway"
	// IndexGatewayRing         string = "index-gateway-ring"
	// QueryScheduler           string = "query-scheduler"
	// UsageReport              string = "usage-report"
)

var objectStoreTypeStats = usagestats.NewString("store_object_type")

func (f *Phlare) initQuerier() (services.Service, error) {
	q, err := querier.New(f.Cfg.Querier, f.ring, nil, f.logger, f.auth)
	if err != nil {
		return nil, err
	}
	// Those API are not meant to stay but allows us for testing through Grafana.
	f.Server.HTTP.Handle("/pyroscope/render", http.HandlerFunc(q.RenderHandler))
	f.Server.HTTP.Handle("/pyroscope/label-values", http.HandlerFunc(q.LabelValuesHandler))
	querierv1connect.RegisterQuerierServiceHandler(f.Server.HTTP, q, f.auth)

	return q, nil
}

func (f *Phlare) getPusherClient() pushv1connect.PusherServiceClient {
	return f.pusherClient
}

func (f *Phlare) initGRPCGateway() (services.Service, error) {
	f.grpcGatewayMux = grpcgw.NewServeMux(
		grpcgw.WithMarshalerOption("application/json+pretty", &grpcgw.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				Indent:    "  ",
				Multiline: true, // Optional, implied by presence of "Indent".
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	return nil, nil
}

func (f *Phlare) initDistributor() (services.Service, error) {
	d, err := distributor.New(f.Cfg.Distributor, f.ring, nil, f.reg, f.logger, f.auth)
	if err != nil {
		return nil, err
	}

	// initialise direct pusher, this overwrites the default HTTP client
	f.pusherClient = d

	pushv1connect.RegisterPusherServiceHandler(f.Server.HTTP, d, f.auth)
	return d, nil
}

func (f *Phlare) initAgent() (services.Service, error) {
	a, err := agent.New(&f.Cfg.AgentConfig, f.logger, f.getPusherClient)
	if err != nil {
		return nil, err
	}
	f.agent = a

	// register endpoint at grpc gateway
	if err := agentv1.RegisterAgentServiceHandlerServer(context.Background(), f.grpcGatewayMux, a); err != nil {
		return nil, err
	}

	agentv1connect.RegisterAgentServiceHandler(f.Server.HTTP, a.ConnectHandler())
	return a, nil
}

func (f *Phlare) initMemberlistKV() (services.Service, error) {
	f.Cfg.MemberlistKV.MetricsRegisterer = f.reg
	f.Cfg.MemberlistKV.Codecs = []codec.Codec{
		ring.GetCodec(),
		usagestats.JSONCodec,
	}

	dnsProviderReg := prometheus.WrapRegistererWithPrefix(
		"phlare_",
		prometheus.WrapRegistererWith(
			prometheus.Labels{"name": "memberlist"},
			f.reg,
		),
	)
	dnsProvider := dns.NewProvider(f.logger, dnsProviderReg, dns.GolangResolverType)

	f.MemberlistKV = memberlist.NewKVInitService(&f.Cfg.MemberlistKV, f.logger, dnsProvider, f.reg)

	f.Cfg.Ingester.LifecyclerConfig.RingConfig.KVStore.MemberlistKV = f.MemberlistKV.GetMemberlistKV

	return f.MemberlistKV, nil
}

func (f *Phlare) initRing() (_ services.Service, err error) {
	f.ring, err = ring.New(f.Cfg.Ingester.LifecyclerConfig.RingConfig, "ingester", "ring", f.logger, prometheus.WrapRegistererWithPrefix("phlare_", f.reg))
	if err != nil {
		return nil, err
	}
	f.Server.HTTP.Path("/ring").Methods("GET", "POST").Handler(f.ring)
	return f.ring, nil
}

func (f *Phlare) initStorage() (_ services.Service, err error) {
	objectStoreTypeStats.Set(f.Cfg.Storage.Bucket.Backend)
	if cfg := f.Cfg.Storage.Bucket; cfg.Backend != "filesystem" {
		b, err := objstoreclient.NewBucket(
			f.context(),
			cfg,
			"storage",
		)
		if err != nil {
			return nil, errors.Wrap(err, "unable to initialise bucket")
		}
		f.storageBucket = b
	}

	if f.Cfg.Target.String() != All && f.storageBucket == nil {
		return nil, errors.New("storage bucket configuration is required when running in microservices mode")
	}

	return nil, nil
}

// TODO: This should be passed to all other services and could also be used to signal shutdown
func (f *Phlare) context() context.Context {
	phlarectx := phlarecontext.WithLogger(context.Background(), f.logger)
	return phlarecontext.WithRegistry(phlarectx, f.reg)
}

func (f *Phlare) initIngester() (_ services.Service, err error) {
	f.Cfg.Ingester.LifecyclerConfig.ListenPort = f.Cfg.Server.HTTPListenPort

	ingester, err := ingester.New(f.context(), f.Cfg.Ingester, f.Cfg.PhlareDB, f.storageBucket)
	if err != nil {
		return nil, err
	}
	prefix, handler := grpchealth.NewHandler(grpchealth.NewStaticChecker(ingesterv1connect.IngesterServiceName))
	f.Server.HTTP.NewRoute().PathPrefix(prefix).Handler(handler)
	ingesterv1connect.RegisterIngesterServiceHandler(f.Server.HTTP, ingester, f.auth)
	return ingester, nil
}

func (f *Phlare) initServer() (services.Service, error) {
	prometheus.MustRegister(version.NewCollector("phlare"))
	DisableSignalHandling(&f.Cfg.Server)
	f.Cfg.Server.Registerer = prometheus.WrapRegistererWithPrefix("phlare_", f.reg)
	// TODO(cyril) figure why this is locking the bidi stream see https://github.com/grafana/phlare/issues/231
	f.Cfg.Server.DoNotAddDefaultHTTPMiddleware = true

	serv, err := server.New(f.Cfg.Server)
	if err != nil {
		return nil, err
	}

	f.Server = serv

	servicesToWaitFor := func() []services.Service {
		svs := []services.Service(nil)
		for m, s := range f.serviceMap {
			// Server should not wait for itself.
			if m != Server {
				svs = append(svs, s)
			}
		}
		return svs
	}

	// sounds like logging is the problem. see https://github.com/grafana/phlare/issues/231
	defaultHTTPMiddleware := []middleware.Interface{
		middleware.Tracer{
			RouteMatcher: f.Server.HTTP,
		},
		// middleware.Log{
		// 	Log:                   f.Server.Log,
		// 	LogRequestAtInfoLevel: f.Cfg.Server.LogRequestAtInfoLevel,
		// },
		// middleware.Instrument{
		// 	RouteMatcher:     router,
		// 	Duration:         requestDuration,
		// 	RequestBodySize:  receivedMessageSize,
		// 	ResponseBodySize: sentMessageSize,
		// 	InflightRequests: inflightRequests,
		// },
	}
	f.Server.HTTPServer.Handler = middleware.Merge(defaultHTTPMiddleware...).Wrap(f.Server.HTTP)

	s := NewServerService(f.Server, servicesToWaitFor, f.logger)
	// todo configure http2
	f.Server.HTTPServer.Handler = h2c.NewHandler(f.Server.HTTPServer.Handler, &http2.Server{})
	f.Server.HTTPServer.Handler = util.RecoveryHTTPMiddleware.Wrap(f.Server.HTTPServer.Handler)

	// expose openapiv2 definition
	openapiv2Handler, err := openapiv2.Handler()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize openapiv2 handler: %w", err)
	}
	f.Server.HTTP.Handle("/api/swagger.json", openapiv2Handler)

	// register grpc-gateway api
	f.Server.HTTP.NewRoute().PathPrefix("/api").Handler(f.grpcGatewayMux)

	// register status service providing config and buildinfo at grpc gateway
	if err := commonv1.RegisterStatusServiceHandlerServer(context.Background(), f.grpcGatewayMux, f.statusService()); err != nil {
		return nil, err
	}

	return s, nil
}

func (f *Phlare) initUsageReport() (services.Service, error) {
	if !f.Cfg.UsageReport.Enabled {
		return nil, nil
	}
	f.Cfg.UsageReport.Leader = false
	// ingester is the only component that can be a leader
	if f.isModuleActive(Ingester) {
		f.Cfg.UsageReport.Leader = true
	}

	usagestats.Target(f.Cfg.Target.String())

	var b objstore.Bucket
	if f.storageBucket == nil {
		if err := os.MkdirAll(f.Cfg.PhlareDB.DataPath, 0o777); err != nil {
			return nil, fmt.Errorf("mkdir %s: %w", f.Cfg.PhlareDB.DataPath, err)
		}
		fs, err := filesystem.NewBucket(f.Cfg.PhlareDB.DataPath)
		if err != nil {
			return nil, err
		}
		b = fs
	}

	ur, err := usagestats.NewReporter(f.Cfg.UsageReport, f.Cfg.Ingester.LifecyclerConfig.RingConfig.KVStore, b, f.logger, f.reg)
	if err != nil {
		level.Info(f.logger).Log("msg", "failed to initialize usage report", "err", err)
		return nil, nil
	}
	f.usageReport = ur
	return ur, nil
}

type statusService struct {
	commonv1.UnimplementedStatusServiceServer
	configYaml    string
	defaultConfig *Config
	actualConfig  *Config
}

func (s *statusService) GetBuildInfo(ctx context.Context, req *commonv1.GetBuildInfoRequest) (*commonv1.GetBuildInfoResponse, error) {
	version := build.GetVersion()
	return &commonv1.GetBuildInfoResponse{
		Status: "success",
		Data: &commonv1.GetBuildInfoData{
			Version:   version.Version,
			Revision:  build.Revision,
			Branch:    version.Branch,
			GoVersion: version.GoVersion,
		},
	}, nil
}

const (
	// There is not standardised and generally used content-type for YAML,
	// text/plain ensures the YAML is displayed in the browser instead of
	// offered as a download
	yamlContentType = "text/plain; charset=utf-8"
)

func (s *statusService) GetConfig(ctx context.Context, req *commonv1.GetConfigRequest) (*httpbody.HttpBody, error) {
	body, err := yaml.Marshal(s.actualConfig)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: yamlContentType,
		Data:        body,
	}, nil
}

func (s *statusService) GetDefaultConfig(ctx context.Context, req *commonv1.GetConfigRequest) (*httpbody.HttpBody, error) {
	body, err := yaml.Marshal(s.defaultConfig)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: yamlContentType,
		Data:        body,
	}, nil
}

func (s *statusService) GetDiffConfig(ctx context.Context, req *commonv1.GetConfigRequest) (*httpbody.HttpBody, error) {
	aBody, err := yaml.Marshal(s.actualConfig)
	if err != nil {
		return nil, err
	}
	aCfg := map[interface{}]interface{}{}
	if err := yaml.Unmarshal(aBody, &aCfg); err != nil {
		return nil, err
	}

	dBody, err := yaml.Marshal(s.defaultConfig)
	if err != nil {
		return nil, err
	}
	dCfg := map[interface{}]interface{}{}
	if err := yaml.Unmarshal(dBody, &dCfg); err != nil {
		return nil, err
	}

	diff, err := util.DiffConfig(dCfg, aCfg)
	if err != nil {
		return nil, err
	}

	body, err := yaml.Marshal(diff)
	if err != nil {
		return nil, err
	}

	return &httpbody.HttpBody{
		ContentType: yamlContentType,
		Data:        body,
	}, nil
}

func (f *Phlare) statusService() commonv1.StatusServiceServer {
	return &statusService{
		actualConfig:  &f.Cfg,
		defaultConfig: newDefaultConfig(),
	}
}

func (f *Phlare) isModuleActive(m string) bool {
	for _, target := range f.Cfg.Target {
		if target == m {
			return true
		}
		if f.recursiveIsModuleActive(target, m) {
			return true
		}
	}
	return false
}

func (f *Phlare) recursiveIsModuleActive(target, m string) bool {
	if targetDeps, ok := f.deps[target]; ok {
		for _, dep := range targetDeps {
			if dep == m {
				return true
			}
			if f.recursiveIsModuleActive(dep, m) {
				return true
			}
		}
	}
	return false
}

// NewServerService constructs service from Server component.
// servicesToWaitFor is called when server is stopping, and should return all
// services that need to terminate before server actually stops.
// N.B.: this function is NOT Cortex specific, please let's keep it that way.
// Passed server should not react on signals. Early return from Run function is considered to be an error.
func NewServerService(serv *server.Server, servicesToWaitFor func() []services.Service, log log.Logger) services.Service {
	serverDone := make(chan error, 1)

	runFn := func(ctx context.Context) error {
		go func() {
			defer close(serverDone)
			serverDone <- serv.Run()
		}()

		select {
		case <-ctx.Done():
			return nil
		case err := <-serverDone:
			if err != nil {
				return err
			}
			return fmt.Errorf("server stopped unexpectedly")
		}
	}

	stoppingFn := func(_ error) error {
		// wait until all modules are done, and then shutdown server.
		for _, s := range servicesToWaitFor() {
			_ = s.AwaitTerminated(context.Background())
		}

		// shutdown HTTP and gRPC servers (this also unblocks Run)
		serv.Shutdown()

		// if not closed yet, wait until server stops.
		<-serverDone
		level.Info(log).Log("msg", "server stopped")
		return nil
	}

	return services.NewBasicService(nil, runFn, stoppingFn)
}

// DisableSignalHandling puts a dummy signal handler
func DisableSignalHandling(config *server.Config) {
	config.SignalHandler = make(ignoreSignalHandler)
}

type ignoreSignalHandler chan struct{}

func (dh ignoreSignalHandler) Loop() {
	<-dh
}

func (dh ignoreSignalHandler) Stop() {
	close(dh)
}
