// START: start
package log

import (
	"context"
	"log"
	"sync"

	"google.golang.org/grpc"

	api "github.com/mferrell/proglog/api/v1"
)

type Replicator struct {
	DialOptions []grpc.DialOption
	LocalServer api.LogClient

	mu      sync.Mutex
	servers map[string]chan struct{}
	closed  bool
	close   chan struct{}
}

// END: start

// START: exported_add
func (r *Replicator) Join(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}

	if _, ok := r.servers[addr]; ok {
		// already replicating so skip
		return nil
	}
	r.servers[addr] = make(chan struct{})

	go r.replicate(addr, r.servers[addr])

	return nil
}

// END: exported_add

// START: start_add
func (r *Replicator) replicate(addr string, leave chan struct{}) {
	cc, err := grpc.Dial(addr, r.DialOptions...)
	if err != nil {
		r.err(err)
		return
	}
	defer cc.Close()

	client := api.NewLogClient(cc)

	ctx := context.Background()
	stream, err := client.ConsumeStream(ctx,
		&api.ConsumeRequest{
			Offset: 0,
		},
	)
	if err != nil {
		r.err(err)
		return
	}

	records := make(chan *api.Record)
	go func() {
		for {
			recv, err := stream.Recv()
			if err != nil {
				r.err(err)
				return
			}
			records <- recv.Record
		}
	}()

	// END: start_add
	// START: end_add
	for {
		select {
		case <-r.close:
			return
		case <-leave:
			return
		case record := <-records:
			_, err = r.LocalServer.Produce(ctx,
				&api.ProduceRequest{
					Record: record,
				},
			)
			if err != nil {
				r.err(err)
				return
			}
		}
	}
}

// END: end_add

// START: remove
func (r *Replicator) Leave(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if _, ok := r.servers[addr]; !ok {
		return nil
	}
	close(r.servers[addr])
	delete(r.servers, addr)
	return nil
}

// END: remove

// START: init
func (r *Replicator) init() {
	if r.servers == nil {
		r.servers = make(map[string]chan struct{})
	}
	if r.close == nil {
		r.close = make(chan struct{})
	}
}

// END: init

// START: close
func (r *Replicator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}
	r.closed = true
	close(r.close)
	return nil
}

// END: close

// START: err
func (r *Replicator) err(err error) {
	log.Printf("[ERROR] proglog: %v", err)
}

// END: err
