package ingester

import (
	"github.com/colega/zeropool"
	pushv1 "github.com/grafana/phlare/api/gen/proto/go/push/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/wal"
)

var pool zeropool.Pool[[]byte]

func init() {
	pool = zeropool.New(func() []byte {
		return make([]byte, 0, 0)
	})
}

type Wal struct {
	log *wal.Log
}

func newWAL(path string, reg prometheus.Registerer) (*Wal, error) {
	log, err := wal.Open(path+"/wal", nil)
	if err != nil {
		return nil, err
	}
	return &Wal{
		log: log,
	}, nil
}

func (w *Wal) Log(record *pushv1.PushRequest) error {
	data := pool.Get()
	defer pool.Put(data)

	_, err := record.MarshalToVT(data)
	if err != nil {
		return err
	}
	w.log.Write(0, data)
	return nil
}
