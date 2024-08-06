package examples

import (
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func RegisterServer(grpcServer *grpc.Server, server server.Server) {
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, server)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)

}

// func RunServer(svr server.Server, port uint) {
// 	var grpcOptions []grpc.ServerOption
// 	grpcOptions = append(grpcOptions,
// 		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
// 		grpc.KeepaliveParams(keepalive.ServerParameters{
// 			Time:    grpcKeepaliveTime,
// 			Timeout: grpcKeepaliveTimeout,
// 		}),
// 		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
// 			MinTime:             grpcKeepaliveMinTime,
// 			PermitWithoutStream: true,
// 		}),
// 	)
// 	grpcServer := grpc.NewServer(grpcOptions...)

// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	registerServer(grpcServer, svr)

// 	log.Printf("management server listening on %d\n", port)
// 	if err = grpcServer.Serve(lis); err != nil {
// 		log.Println(err)
// 	}
// }
