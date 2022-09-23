package debuginfo

import (
	"context"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	parcadebuginfov1 "github.com/parca-dev/parca/gen/proto/go/parca/debuginfo/v1alpha1"
)

type fakeDebugInfo struct {
	parcadebuginfov1.UnimplementedDebugInfoServiceServer

	logger log.Logger
}

func New(logger log.Logger) parcadebuginfov1.DebugInfoServiceServer {
	return &fakeDebugInfo{
		logger: logger,
	}
}

// Exists returns true if the given build_id has debug info uploaded for it.
func (f *fakeDebugInfo) Exists(ctx context.Context, req *parcadebuginfov1.ExistsRequest) (*parcadebuginfov1.ExistsResponse, error) {
	level.Warn(f.logger).Log("msg", "received exists request", "buildid", req.GetBuildId(), "hash", req.GetHash())

	return &parcadebuginfov1.ExistsResponse{
		Exists: false,
	}, nil
}

// Upload ingests debug info for a given build_id
func (f *fakeDebugInfo) Upload(u parcadebuginfov1.DebugInfoService_UploadServer) error {
	req, err := u.Recv()
	if err != nil {
		return err
	}
	level.Warn(f.logger).Log("msg", "received upload", "buildid", req.GetInfo().GetBuildId(), "hash", req.GetInfo().GetHash())

	return nil
}

// Download returns the debug info for a given build_id.
func (_ *fakeDebugInfo) Download(*parcadebuginfov1.DownloadRequest, parcadebuginfov1.DebugInfoService_DownloadServer) error {
	return nil
}
