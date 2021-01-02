package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"

	log "github.com/iancoffey/moraceae/pkg/log"
	snapshot "github.com/iancoffey/moraceae/pkg/snapshot"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	xDSPort                  = 18000
	grpcMaxConcurrentStreams = 10000
)

func RunManagementServer(ctx context.Context, server xds.Server, logger *log.Logger) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", xDSPort))
	if err != nil {
		logger.Errorf("at=binding-xdp-port port=%s err=%q", xDSPort, err)
	}

	// register services
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	v2.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, server)

	logger.Infof("at=xds-server-listening port=%d", xDSPort)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Errorf("at=grpc-serve port=%d err=%q", xDSPort, err)
		}
	}()
	<-ctx.Done()

	grpcServer.GracefulStop()
}

func main() {
	ctx := context.Background()

	viper.SetEnvPrefix("PICO")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	pflag.StringSlice("cluster-id", []string{}, "The unique ID of the cluster")
	pflag.StringSlice("debug", []string{}, "debug")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	services := viper.GetStringSlice("target-services") // flags
	clusterID := viper.GetString("cluster-id")

	nodeID := viper.GetString("HOSTNAME") //env from downward api

	logger := log.NewLogger(nodeID, clusterID, viper.GetBool("debug"))
	logger.Infof("at=booting port=%d target-services=%d cluster-id=%s node-id=%s", xDSPort, services, clusterID, nodeID)

	signal := make(chan struct{})
	cb := &callbacks{
		signal:   signal,
		fetches:  0,
		requests: 0,
		logger:   logger,
	}
	snapshotCache := cache.NewSnapshotCache(true, cache.IDHash{}, nil)

	srv := xds.NewServer(ctx, snapshotCache, cb)

	// We need to boot the mgmt server
	go RunManagementServer(ctx, srv, logger)
	//go RunManagementGateway(ctx, srv, gatewayPort)

	<-signal

	nodeId := snapshotCache.GetStatusKeys()[0]
	logger.Infof("at=adding-node NodeID=%s", nodeId)
	snapshotter := snapshot.New(logger, clusterID)

	for {
		ss, err := snapshotter.Generate(services)
		if err != nil {
			logger.Errorf("Error in Generating the SnapShot %q", err)
		} else {
			snapshotCache.SetSnapshot(nodeID, *ss)
			time.Sleep(60 * time.Second)
		}
	}

}
