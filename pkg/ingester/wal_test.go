package ingester

import (
	"testing"

	pushv1 "github.com/grafana/phlare/api/gen/proto/go/push/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func Test_WalWrite(t *testing.T) {
	wallDir := t.TempDir()

	wal, err := newWAL(wallDir, prometheus.NewRegistry())
	require.NoError(t, err)

	err = wal.Log(&pushv1.PushRequest{})
	require.NoError(t, err)
}
