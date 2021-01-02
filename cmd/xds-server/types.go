package main

import (
	"context"
	"sync"

	log "github.com/iancoffey/moraceae/pkg/log"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func (cb *callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.logger.Infof("at=report fetches=%d requests=%d", cb.fetches, cb.requests)
}
func (cb *callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	cb.logger.Infof("OnStreamOpen %d open for Type [%s]", id, typ)
	return nil
}
func (cb *callbacks) OnStreamClosed(id int64) {
	cb.logger.Infof("OnStreamClosed %d closed", id)
}
func (cb *callbacks) OnStreamRequest(id int64, r *v2.DiscoveryRequest) error {
	cb.logger.Infof("OnStreamRequest %d  Request[%v]", id, r.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.requests++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *callbacks) OnStreamResponse(id int64, req *v2.DiscoveryRequest, resp *v2.DiscoveryResponse) {
	cb.logger.Infof("OnStreamResponse... %d   Request [%v],  Response[%v]", id, req.TypeUrl, resp.TypeUrl)
	cb.Report()
}
func (cb *callbacks) OnFetchRequest(ctx context.Context, req *v2.DiscoveryRequest) error {
	cb.logger.Infof("OnFetchRequest... Request [%v]", req.TypeUrl)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.fetches++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *callbacks) OnFetchResponse(req *v2.DiscoveryRequest, resp *v2.DiscoveryResponse) {
	cb.logger.Infof("OnFetchResponse... Resquest[%v],  Response[%v]", req.TypeUrl, resp.TypeUrl)
}

type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
	logger   *log.Logger
}
