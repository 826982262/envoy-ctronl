package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	http_connection_manager "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

type EnvoyCluster struct {
	name      string
	port      uint32
	endpoints []string
}

var (
	endpoints     []types.Resource
	version       int
	snapshotCache cache.SnapshotCache
)

var envoyCluster EnvoyCluster

func SetSnapshot(ctx context.Context, nodeId string, sc cache.SnapshotCache) error {
	envoyCluster.name = "xds_demo_cluster"
	// clusterName := "xds_demo_cluster"

	manager := buildHttpManager(envoyCluster.name, "www.envoyproxy.io")
	fcs := buildFilterChain(manager)
	serviceListener := buildListener("0.0.0.0", 14000, fcs)

	// serviceEndpoint := buildEndpoint("www.envoyproxy.io", 443)
	serviceEndpoint := buildEndpoints()

	serviceCluster := buildClusters(envoyCluster.name, serviceEndpoint)

	// fmt.Printf("serviceCluster:%v", serviceCluster)
	// fmt.Printf("serviceEndpoint:%v", serviceEndpoint)

	rs := map[resource.Type][]types.Resource{
		resource.ClusterType: {serviceCluster},
		// resource.EndpointType: {serviceEndpoint},
		// resource.EndpointType: edsEndpoints,

		resource.ListenerType: {serviceListener},
		resource.RouteType:    {manager},
	}
	// fmt.Printf("%v", rs)

	snapshot, err := cache.NewSnapshot("1", rs)

	if err != nil {
		log.Fatalf("new snapshot error: %v", err)
	}
	return sc.SetSnapshot(ctx, nodeId, snapshot)
}

func buildFilterChain(manager *http_connection_manager.HttpConnectionManager) []*listener.FilterChain {
	managerPB, err := anypb.New(manager)
	if err != nil {
		log.Fatalf("Failed to marshal HttpConnectionManager: %v", err)
	}

	fcs := make([]*listener.FilterChain, 0, 0)
	fcs = append(fcs, &listener.FilterChain{
		Filters: []*listener.Filter{
			{
				Name: "envoy.filters.network.http_connection_manager",
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: managerPB,
				},
			},
		},
	})
	return fcs
}

func buildHttpManager(clusterName string, upstreamHost string) *http_connection_manager.HttpConnectionManager {
	xdsRoute := &route.RouteConfiguration{
		Name: "xds_demo_route",
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "xds_demo_service",
				Domains: []string{"*"},
				Routes: []*route.Route{
					{
						Match: &route.RouteMatch{
							PathSpecifier: &route.RouteMatch_Prefix{
								Prefix: "/",
							},
						},
						Action: &route.Route_Route{
							Route: &route.RouteAction{
								HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
									HostRewriteLiteral: upstreamHost,
								},
								// 集群要去下文一致
								ClusterSpecifier: &route.RouteAction_Cluster{
									Cluster: clusterName,
								},
							},
						},
					},
				},
			},
		},
	}
	routerConfig, _ := anypb.New(&routerv3.Router{})
	// http 链接管理器
	manager := &http_connection_manager.HttpConnectionManager{
		StatPrefix: "ingress_http",
		RouteSpecifier: &http_connection_manager.HttpConnectionManager_RouteConfig{
			RouteConfig: xdsRoute,
		},
		HttpFilters: []*http_connection_manager.HttpFilter{
			{
				Name: "envoy.filters.http.router",
				ConfigType: &http_connection_manager.HttpFilter_TypedConfig{
					TypedConfig: routerConfig,
				},
			},
		},
		SchemeHeaderTransformation: &corev3.SchemeHeaderTransformation{
			Transformation: &corev3.SchemeHeaderTransformation_SchemeToOverwrite{
				SchemeToOverwrite: "https",
			},
		},
	}
	return manager
}

func buildEndpoint(addr string, port uint32) *endpoint.LbEndpoint {

	epTarget := &endpoint.LbEndpoint{
		HostIdentifier: &endpoint.LbEndpoint_Endpoint{
			Endpoint: &endpoint.Endpoint{
				Address: &corev3.Address{
					Address: &corev3.Address_SocketAddress{
						SocketAddress: &corev3.SocketAddress{
							Address: addr,
							PortSpecifier: &corev3.SocketAddress_PortValue{
								PortValue: port,
							},
						},
					},
				},
			},
		},
	}
	return epTarget
}

type lbaddrs struct {
	Name   string
	Ipaddr string
	Port   uint32
}

func buildEndpoints() []*endpoint.LbEndpoint {
	lbs := []lbaddrs{
		{Name: "test", Port: 443, Ipaddr: "cn.bing.com"}, {Name: "test2", Port: 443, Ipaddr: "www.envoyproxy.io"},
	}

	lbEndpoints := []*endpoint.LbEndpoint{}
	for _, lb := range lbs {
		epTarget := &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &corev3.Address{
						Address: &corev3.Address_SocketAddress{
							SocketAddress: &corev3.SocketAddress{
								Address: lb.Ipaddr,
								PortSpecifier: &corev3.SocketAddress_PortValue{
									PortValue: lb.Port,
								},
							},
						},
					},
				},
			},
		}
		lbEndpoints = append(lbEndpoints, epTarget)

	}

	fmt.Printf("%v", lbEndpoints)
	return lbEndpoints
}

func buildClusters(clusterName string, ep []*endpoint.LbEndpoint) *cluster.Cluster {
	serviceCluster := &cluster.Cluster{
		Name:           clusterName,
		ConnectTimeout: durationpb.New(time.Second * 3),
		ClusterDiscoveryType: &cluster.Cluster_Type{
			Type: cluster.Cluster_STRICT_DNS,
		},
		DnsLookupFamily: cluster.Cluster_V4_ONLY,
		LbPolicy:        cluster.Cluster_ROUND_ROBIN,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: ep,
				},
			},
		},
		TransportSocket: &corev3.TransportSocket{
			Name:       "envoy.transport_sockets.tls",
			ConfigType: nil,
		},
	}
	us := &tlsv3.UpstreamTlsContext{
		Sni: "www.envoyproxy.io",
	}
	tlsConfig, _ := anypb.New(us)
	serviceCluster.TransportSocket.ConfigType = &corev3.TransportSocket_TypedConfig{
		TypedConfig: tlsConfig,
	}
	return serviceCluster
}

func buildListener(ip string, port uint32, fcs []*listener.FilterChain) *listener.Listener {
	return &listener.Listener{
		Name: "listener_xds_demo",
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: ip,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		// 过滤器链
		FilterChains: fcs,
	}
}
