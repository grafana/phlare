package querier

import (
	"context"

	"github.com/grafana/dskit/ring"
)

type ResponseFromReplica[T interface{}] struct {
	addr     string
	response T
}

type QueryReplicaFn[T interface{}, Querier any] func(context.Context, Querier) (T, error)

type Closer interface {
	CloseRequest() error
	CloseResponse() error
}

type ClientFactory[T any] func(addr string) (T, error)

// forGivenReplicationSet runs f, in parallel, for given ingesters
func forGivenReplicationSet[Result any, Querier any](ctx context.Context, clientFactory func(string) (Querier, error), replicationSet ring.ReplicationSet, f QueryReplicaFn[Result, Querier]) ([]ResponseFromReplica[Result], error) {
	results, err := ring.DoUntilQuorum(ctx, replicationSet, func(ctx context.Context, ingester *ring.InstanceDesc) (ResponseFromReplica[Result], error) {
		var res ResponseFromReplica[Result]
		client, err := clientFactory(ingester.Addr)
		if err != nil {
			return res, err
		}

		resp, err := f(ctx, client)
		if err != nil {
			return res, err
		}

		return ResponseFromReplica[Result]{ingester.Addr, resp}, nil
	}, func(result ResponseFromReplica[Result]) {
		// If the result was streamed, we need to close the request and response
		if stream, ok := any(result.response).(interface {
			CloseRequest() error
		}); ok {
			stream.CloseRequest()
		}
		if stream, ok := any(result.response).(interface {
			CloseResponse() error
		}); ok {
			stream.CloseResponse()
		}
	})
	if err != nil {
		return nil, err
	}

	responses := make([]ResponseFromReplica[Result], 0, len(results))
	for _, result := range results {
		responses = append(responses, result)
	}

	return responses, err
}
