package ingester

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/services"
	"github.com/oklog/ulid"

	"github.com/grafana/phlare/pkg/phlaredb/block"
	diskutil "github.com/grafana/phlare/pkg/util/disk"
)

const (
	defaultMinFreeDisk                        = 10 * 1024 * 1024 * 1024 // 10Gi
	defaultMinDiskAvailablePercentage         = 0.05
	defaultRetentionPolicyEnforcementInterval = 5 * time.Minute

	phlareDBLocalPath = "local"
)

type retentionPolicy struct {
	MinFreeDisk                uint64
	MinDiskAvailablePercentage float64
	EnforcementInterval        time.Duration
}

func defaultRetentionPolicy() retentionPolicy {
	return retentionPolicy{
		MinFreeDisk:                defaultMinFreeDisk,
		MinDiskAvailablePercentage: defaultMinDiskAvailablePercentage,
		EnforcementInterval:        defaultRetentionPolicyEnforcementInterval,
	}
}

type retentionPolicyEnforcer struct {
	services.Service

	ingester *Ingester
	policy   retentionPolicy
	fs       fileSystem
	vc       diskutil.VolumeChecker

	stopCh chan struct{}
	wg     sync.WaitGroup
}

type tenantBlock struct {
	ulid     ulid.ULID
	tenantID string
	path     string
}

type fileSystem interface {
	fs.ReadDirFS
	RemoveAll(string) error
}

type realFileSystem struct{}

func (*realFileSystem) Open(name string) (fs.File, error)          { return os.Open(name) }
func (*realFileSystem) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }
func (*realFileSystem) RemoveAll(path string) error                { return os.RemoveAll(path) }

func newRetentionPolicyEnforcer(ingester *Ingester, policy retentionPolicy) *retentionPolicyEnforcer {
	e := retentionPolicyEnforcer{
		ingester: ingester,
		policy:   policy,
		stopCh:   make(chan struct{}),
		fs:       new(realFileSystem),
		vc:       diskutil.NewVolumeChecker(policy.MinFreeDisk, policy.MinDiskAvailablePercentage),
	}
	e.Service = services.NewBasicService(nil, e.running, e.stopping)
	return &e
}

func (e *retentionPolicyEnforcer) running(ctx context.Context) error {
	e.wg.Add(1)
	retentionPolicyEnforcerTicker := time.NewTicker(e.policy.EnforcementInterval)
	defer func() {
		retentionPolicyEnforcerTicker.Stop()
		e.wg.Done()
	}()
	for {
		// Enforce retention policy immediately at start.
		if err := e.cleanupBlocksWhenHighDiskUtilization(ctx); err != nil {
			level.Error(e.ingester.logger).Log("msg", "failed to enforce retention policy", "err", err)
		}
		select {
		case <-retentionPolicyEnforcerTicker.C:
		case <-ctx.Done():
			return nil
		case <-e.stopCh:
			return nil
		}
	}
}

func (e *retentionPolicyEnforcer) stopping(_ error) error {
	close(e.stopCh)
	e.wg.Wait()
	return nil
}

func (e *retentionPolicyEnforcer) localBlocks(root string) ([]*tenantBlock, error) {
	blocks := make([]*tenantBlock, 0, 32)
	tenants, err := fs.ReadDir(e.fs, root)
	if err != nil {
		return nil, err
	}
	var blockDirs []fs.DirEntry
	for _, tenantDir := range tenants {
		if !tenantDir.IsDir() {
			continue
		}
		tenantID := tenantDir.Name()
		tenantDirPath := filepath.Join(root, tenantID, phlareDBLocalPath)
		if blockDirs, err = fs.ReadDir(e.fs, tenantDirPath); err != nil {
			if os.IsNotExist(err) {
				// Must be created by external means, skipping.
				continue
			}
			return nil, err
		}
		for _, blockDir := range blockDirs {
			if !blockDir.IsDir() {
				continue
			}
			blockPath := filepath.Join(tenantDirPath, blockDir.Name())
			if blockID, ok := block.IsBlockDir(blockPath); ok {
				blocks = append(blocks, &tenantBlock{
					ulid:     blockID,
					path:     blockPath,
					tenantID: tenantID,
				})
			}
			// A malformed/invalid ULID likely means that the
			// directory is not a valid block, ignoring.
		}
	}

	// Sort the blocks by their id, which will be the time they've been created.
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].ulid.Compare(blocks[j].ulid) < 0
	})

	return blocks, nil
}

func (e *retentionPolicyEnforcer) cleanupBlocksWhenHighDiskUtilization(ctx context.Context) error {
	var volumeStatsPrev *diskutil.VolumeStats
	volumeStatsCurrent, err := e.vc.HasHighDiskUtilization(e.ingester.dbConfig.DataPath)
	if err != nil {
		return err
	}
	// Not in high disk utilization, nothing to do.
	if !volumeStatsCurrent.HighDiskUtilization {
		return nil
	}
	// Get all block across all the tenants. Any block
	// produced or imported during the procedure is ignored.
	blocks, err := e.localBlocks(e.ingester.dbConfig.DataPath)
	if err != nil {
		return err
	}

	for volumeStatsCurrent.HighDiskUtilization && len(blocks) > 0 && ctx.Err() == nil {
		// When disk utilization is not lower since the last loop, we end the
		// cleanup there to avoid deleting all blocks when disk usage reporting
		// is delayed.
		if volumeStatsPrev != nil && volumeStatsPrev.BytesAvailable >= volumeStatsCurrent.BytesAvailable {
			level.Warn(e.ingester.logger).Log("msg", "disk utilization is not lowered by deletion of a block, pausing until next cycle")
			break
		}
		// Delete the oldest block.
		var b *tenantBlock
		b, blocks = blocks[0], blocks[1:]
		level.Warn(e.ingester.logger).Log("msg", "disk utilization is high, deleted oldest block", "path", b.path)
		if err = e.deleteBlock(b); err != nil {
			return err
		}
		volumeStatsPrev = volumeStatsCurrent
		if volumeStatsCurrent, err = e.vc.HasHighDiskUtilization(e.ingester.dbConfig.DataPath); err != nil {
			return err
		}
	}

	return ctx.Err()
}

func (e *retentionPolicyEnforcer) deleteBlock(b *tenantBlock) error {
	// We lock instances map for writes to ensure that no new instances are created
	// during the procedure. Otherwise, during initialization, the new PhlareDB
	// instance may load a block that has already been deleted.
	e.ingester.instancesMtx.RLock()
	defer e.ingester.instancesMtx.RUnlock()
	// The map only contains PhlareDB instances that has been initialized since
	// the process start, therefore there is no guarantee that we will find the
	// discovered candidate block there. If it is the case, we have to ensure that
	// the block won't be accessed, before and during deleting it from the disk.
	if pdb, ok := e.ingester.instances[b.tenantID]; ok {
		if _, err := pdb.Evict(b.ulid); err != nil {
			return fmt.Errorf("failed to evict block %q: %w", b.path, err)
		}
	}
	if err := e.fs.RemoveAll(b.path); err != nil {
		if os.IsNotExist(err) {
			level.Warn(e.ingester.logger).Log("msg", "block not found on disk", "path", b.path)
			return nil
		}
		return fmt.Errorf("failed to delete oldest block %q: %w", b.path, err)
	}
	return nil
}
