package phlaredb

import (
	"context"
	"fmt"
	"sync"
	"time"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/segmentio/parquet-go"
)

// pool for profile rows
var profilePool = sync.Pool{
	New: func() interface{} {
		return parquet.Row{}
	},
}

type profileHead struct {
	buffer *parquet.GenericBuffer[*schemav1.Profile]
	ch     chan *schemav1.Profile

	persister schemav1.ProfilePersister
	helper    profilesHelper
}

func newProfileHead() *profileHead {
	h := &profileHead{
		ch: make(chan *schemav1.Profile, 32),
	}
	h.buffer = parquet.NewGenericBuffer[*schemav1.Profile](h.persister.Schema(), h.persister.SortingColumns())

	return h
}

func (p *profileHead) run(ctx context.Context) error {
	go func() {
		var (
			batch              = make([]*schemav1.Profile, 0, 32)
			oldestBatchElement time.Time
		)

		for {

			if (!oldestBatchElement.IsZero() && time.Since(oldestBatchElement) > time.Second) || len(batch) == cap(batch) {
				l, err := p.buffer.Write(batch)
				if err != nil {
					panic(err)
				}
				fmt.Printf("wrote %d profiles\n", l)

				// reset batch
				for pos := range batch {
					profilePool.Put(batch[pos])
				}

				batch = batch[:0]
				oldestBatchElement = time.Time{}

			}

			select {
			case <-ctx.Done():
				return
			case profile := <-p.ch:
				if oldestBatchElement.IsZero() {
					oldestBatchElement = time.Now()
				}
				batch = append(batch, profile)
			}
		}
	}()
	return nil
}

func (p *profileHead) ingest(ctx context.Context, elems []*schemav1.Profile, rewriter *rewriter) error {
	h := profilesHelper{}

	// rewrite elements
	for pos := range elems {
		if err := h.rewrite(rewriter, elems[pos]); err != nil {
			return err
		}

		row := profilePool.Get().(parquet.Row)
		p.ch <- p.persister.Deconstruct(row, 0, elems[pos])
	}
	return nil
}
