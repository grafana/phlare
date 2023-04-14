package ingester

import (
	"io/fs"

	diskutil "github.com/grafana/phlare/pkg/util/disk"
	"github.com/stretchr/testify/mock"
)

type fakeVolumeFS struct {
	mock.Mock
}

func (f *fakeVolumeFS) HasHighDiskUtilization(path string) (*diskutil.VolumeStats, error) {
	args := f.Called(path)
	return args[0].(*diskutil.VolumeStats), args.Error(1)
}

func (f *fakeVolumeFS) Open(path string) (fs.File, error) {
	args := f.Called(path)
	return args[0].(fs.File), args.Error(1)
}

func (f *fakeVolumeFS) RemoveAll(path string) error {
	args := f.Called(path)
	return args.Error(0)
}

func (f *fakeVolumeFS) ReadDir(path string) ([]fs.DirEntry, error) {
	args := f.Called(path)
	return args[0].([]fs.DirEntry), args.Error(1)
}

type fakeFile struct {
	name string
	dir  bool
}

func (f *fakeFile) Name() string               { return f.name }
func (f *fakeFile) IsDir() bool                { return f.dir }
func (f *fakeFile) Info() (fs.FileInfo, error) { panic("not implemented") }
func (f *fakeFile) Type() fs.FileMode          { panic("not implemented") }

/*
func TestPhlareDB_cleanupBlocksWhenHighDiskUtilization(t *testing.T) {
	const suffix = "0000000000000000000000"

	for _, tc := range []struct {
		name     string
		mock     func(fs *fakeVolumeFS)
		logLines []string
		err      string
	}{
		{
			name: "no-high-disk-utilization",
			mock: func(f *fakeVolumeFS) {
				f.On("HasHighDiskUtilization", mock.Anything).Return(&diskutil.VolumeStats{HighDiskUtilization: false}, nil).Once()
			},
		},
		{
			name: "high-disk-utilization-no-blocks",
			mock: func(f *fakeVolumeFS) {
				f.On("HasHighDiskUtilization", mock.Anything).Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
				f.On("ReadDir", mock.Anything).Return([]fs.DirEntry{&fakeFile{"just-a-file", false}}, nil).Once()
			},
		},
		{
			name: "high-disk-utilization-delete-single-block",
			mock: func(f *fakeVolumeFS) {
				f.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
				f.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
					&fakeFile{"01AA" + suffix, true},
				}, nil).Once()
				f.On("RemoveAll", "local/01AA"+suffix).Return(nil).Once()
				f.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: false, BytesAvailable: 11}, nil).Once()
			},
			logLines: []string{`{"level":"warn", "msg":"disk utilization is high, deleted oldest block", "path":"local/01AA0000000000000000000000"}`},
		},
		{
			name: "high-disk-utilization-delete-two-blocks",
			mock: func(f *fakeVolumeFS) {
				f.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
				f.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
					&fakeFile{"01AA" + suffix, true},
				}, nil).Once()
				f.On("RemoveAll", "local/01AA"+suffix).Return(nil).Once()
				f.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 11}, nil).Once()
				f.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
				}, nil).Once()
				f.On("RemoveAll", "local/01AB"+suffix).Return(nil).Once()
				f.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: false, BytesAvailable: 12}, nil).Once()
			},
			logLines: []string{
				`{"level":"warn", "msg":"disk utilization is high, deleted oldest block", "path":"local/01AA0000000000000000000000"}`,
				`{"level":"warn", "msg":"disk utilization is high, deleted oldest block", "path":"local/01AB0000000000000000000000"}`,
			},
		},
		{
			name: "high-disk-utilization-delete-blocks-no-reduction-in-usage",
			mock: func(fakeFS *fakeVolumeFS) {
				fakeFS.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
				fakeFS.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
					&fakeFile{"01AA" + suffix, true},
				}, nil).Once()
				fakeFS.On("RemoveAll", "local/01AA"+suffix).Return(nil).Once()
				fakeFS.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
				fakeFS.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
				}, nil).Once()
				fakeFS.On("RemoveAll", "local/01AB"+suffix).Return(nil).Once()
				fakeFS.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
			},
			logLines: []string{
				`{"level":"warn", "msg":"disk utilization is high, deleted oldest block", "path":"local/01AA0000000000000000000000"}`,
				`{"level":"warn", "msg":"disk utilization is not lowered by deletion of block, pausing until next cycle", "path":"local"}`,
			},
		},
		{
			name: "high-disk-utilization-delete-blocks-block-not-removed",
			mock: func(fakeFS *fakeVolumeFS) {
				fakeFS.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 10}, nil).Once()
				fakeFS.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
					&fakeFile{"01AA" + suffix, true},
				}, nil).Once()
				fakeFS.On("RemoveAll", "local/01AA"+suffix).Return(nil).Once()
				fakeFS.On("HasHighDiskUtilization", "local").Return(&diskutil.VolumeStats{HighDiskUtilization: true, BytesAvailable: 11}, nil).Once()
				fakeFS.On("ReadDir", mock.Anything).Return([]fs.DirEntry{
					&fakeFile{"01AC" + suffix, true},
					&fakeFile{"01AB" + suffix, true},
					&fakeFile{"01AA" + suffix, true},
				}, nil).Once()
			},
			err: "making no progress in deletion: trying to delete block '01AA0000000000000000000000' again",
			logLines: []string{
				`{"level":"warn", "msg":"disk utilization is high, deleted oldest block", "path":"local/01AA0000000000000000000000"}`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				logBuf = bytes.NewBuffer(nil)
				logger = log.NewJSONLogger(log.NewSyncWriter(logBuf))
				ctx    = context.Background()
				fakeFS = &fakeVolumeFS{}
			)

			db := &PhlareDB{
				logger:        logger,
				volumeChecker: fakeFS,
				fs:            fakeFS,
			}

			tc.mock(fakeFS)

			if tc.err == "" {
				require.NoError(t, db.cleanupBlocksWhenHighDiskUtilization(ctx))
			} else {
				require.Equal(t, tc.err, db.cleanupBlocksWhenHighDiskUtilization(ctx).Error())
			}

			// check for log lines
			if len(tc.logLines) > 0 {
				lines := strings.Split(strings.TrimSpace(logBuf.String()), "\n")
				require.Len(t, lines, len(tc.logLines))
				for idx := range tc.logLines {
					require.JSONEq(t, tc.logLines[idx], lines[idx])
				}
			}
		})
	}
}
*/
