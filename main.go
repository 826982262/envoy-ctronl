package main

import (
	"context"
	"net"
	"test/examples"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
)

func main() {
	var log = examples.Logger{}
	log.Debug = true

	ctx := context.Background()
	lis, err := net.Listen("tcp", ":18000")
	if err != nil {
		log.Errorf("Failed to listen: %v", err)
	}
	conn, err := lis.Accept()
	log.Debugf("envoy地址：%v", conn.RemoteAddr())

	sc := cache.NewSnapshotCache(true, cache.IDHash{}, log)
	srv := server.NewServer(ctx, sc, nil)
	// new grpc server
	gs := grpc.NewServer()

	examples.RegisterServer(gs, srv)
	// clusterservice.RegisterClusterDiscoveryServiceServer(gs, srv)
	// endpointservice.RegisterEndpointDiscoveryServiceServer(gs, srv)
	// listenerservice.RegisterListenerDiscoveryServiceServer(gs, srv)
	// routeservice.RegisterRouteDiscoveryServiceServer(gs, srv)
	// discoverygrpc.RegisterAggregatedDiscoveryServiceServer(gs, srv)

	err = examples.SetSnapshot(ctx, "xds-node-id", sc)
	if err != nil {
		log.Errorf("set snapshot error: %v", err)
	} else {
		log.Infof("set snapshot success")
	}

	log.Infof("Starting control plane server...")
	if err := gs.Serve(lis); err != nil {
		log.Errorf("Failed to serve: %v", err)
	}

}
