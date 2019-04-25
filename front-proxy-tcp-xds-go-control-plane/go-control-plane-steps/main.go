package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"sync"
	"sync/atomic"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/envoyproxy/go-control-plane/pkg/util"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	accessLog_v2 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v2"
	filterAccessLog_v2 "github.com/envoyproxy/go-control-plane/envoy/config/filter/accesslog/v2"
	tcpProxy_v2 "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
)

var (
	debug bool

	port uint

	mode string

	version int32

	config cache.SnapshotCache
)

const (
	xdsCluster = "xds_cluster"
)

func init() {
	flag.BoolVar(&debug, "debug", true, "Use debug logging")
	flag.UintVar(&port, "port", 5678, "Management server port")
	flag.StringVar(&mode, "xds", "xds", "Management server type (ads, xds, rest)")
}

type logger struct{}

func (logger logger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}
func (logger logger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
func (cb *callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	log.WithFields(log.Fields{"fetches": cb.fetches, "requests": cb.requests}).Info("cb.Report()  callbacks")
}
func (cb *callbacks) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	log.Infof("OnStreamOpen %v:%v open for %s", ctx, id, typ)
	return nil
}
func (cb *callbacks) OnStreamClosed(id int64) {
	log.Infof("OnStreamClosed %d closed", id)
}
func (cb *callbacks) OnStreamRequest(int64, *v2.DiscoveryRequest) error {
	log.Infof("OnStreamRequest")
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.requests++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {
	log.Infof("OnStreamResponse...")
	cb.Report()
}
func (cb *callbacks) OnFetchRequest(ctx context.Context, req *v2.DiscoveryRequest) error {
	log.Infof("OnFetchRequest for ctx %v", ctx)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.fetches++
	if cb.signal != nil {
		close(cb.signal)
		cb.signal = nil
	}
	return nil
}
func (cb *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {}

type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
}

// Hasher returns node ID as an ID
type Hasher struct {
}

// ID function
func (h Hasher) ID(node *core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

const grpcMaxConcurrentStreams = 1000000

// RunManagementServer starts an xDS server at the given port.
func RunManagementServer(ctx context.Context, server xds.Server, port uint) {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	// register services
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, server)

	log.Infof("Will start the server")
	log.WithFields(log.Fields{"port": port}).Info("management server listening")
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			log.Error(err)
		}
	}()
	<-ctx.Done()
	log.WithFields(log.Fields{"port": port}).Info("management server listening started")

	grpcServer.GracefulStop()
}

func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	ctx := context.Background()

	log.Printf("Starting control plane")

	signal := make(chan struct{})
	cb := &callbacks{
		signal:   signal,
		fetches:  0,
		requests: 0,
	}
	config = cache.NewSnapshotCache(false, Hasher{}, logger{})

	srv := xds.NewServer(config, cb)

	// start the xDS server
	go RunManagementServer(ctx, srv, port)

	<-signal
	cb.Report()

	atomic.AddInt32(&version, 1)
	nodeID := config.GetStatusKeys()
	log.Infof("Getting the nodeId as %v", nodeID)

	log.Infof(">>>>>>>>>>>>>>>>>>> creating listeners ")

	filterStatPrefix := "ingress_tcp"
	lbCluster := "servicelb" // the cluster the filter will redirect the traffic to.
	listenerAddress := "0.0.0.0"
	var listenerPort uint32 = 8200

	accessLogConfigStruct := &accessLog_v2.FileAccessLog{
		Path: "/dev/stdout",
	}

	accessLogConfig, err := util.MessageToStruct(accessLogConfigStruct)
	if err != nil {
		log.Errorf("Failed to convert accessLog message %+v to struct", accessLogConfigStruct)
		panic(err)
	}

	filterConfigStruct := &tcpProxy_v2.TcpProxy{
		StatPrefix: filterStatPrefix,
		ClusterSpecifier: &tcpProxy_v2.TcpProxy_Cluster{
			Cluster: lbCluster,
		},
		AccessLog: []*filterAccessLog_v2.AccessLog{
			&filterAccessLog_v2.AccessLog{
				Name: util.FileAccessLog,
				ConfigType: &filterAccessLog_v2.AccessLog_Config{
					Config: accessLogConfig,
				},
			},
		},
	}

	filterConfig, err := util.MessageToStruct(filterConfigStruct)
	if err != nil {
		log.Errorf("Failed to convert filterConfig message %+v to struct", filterConfigStruct)
		panic(err)
	}

	l := []cache.Resource{
		&v2.Listener{
			Name: "listener1",
			Address: core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: core.TCP,
						Address:  listenerAddress,
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: listenerPort,
						},
					},
				},
			},
			FilterChains: []listener.FilterChain{
				listener.FilterChain{
					Filters: []listener.Filter{
						listener.Filter{
							Name: util.TCPProxy,
							ConfigType: &listener.Filter_Config{
								Config: filterConfig,
							},
						},
					},
				},
			},
		},
	}

	log.Infof(">>>>>>>>>>>>>>>>>>> creating clusters ")

	var serviceName = "serviceendpoints"
	log.Infof(">>>>>>>>>>>>>>>>>>> creating endpoints group " + serviceName)

	c := []cache.Resource{
		&v2.Cluster{
			Name:           lbCluster,
			ConnectTimeout: time.Millisecond * 250,
			LbPolicy:       v2.Cluster_ROUND_ROBIN,
			ClusterDiscoveryType: &v2.Cluster_Type{
				Type: v2.Cluster_EDS,
			},
			EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
				ServiceName: serviceName,
				EdsConfig: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
						ApiConfigSource: &core.ApiConfigSource{
							ApiType: core.ApiConfigSource_GRPC,
							GrpcServices: []*core.GrpcService{
								&core.GrpcService{
									TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
										EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
											ClusterName: xdsCluster,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	e := []cache.Resource{
		&v2.ClusterLoadAssignment{
			ClusterName: serviceName,
			Endpoints: []endpoint.LocalityLbEndpoints{
				endpoint.LocalityLbEndpoints{
					Locality: nil,
					LbEndpoints: []endpoint.LbEndpoint{
						endpoint.LbEndpoint{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &core.Address{
										Address: &core.Address_SocketAddress{
											SocketAddress: &core.SocketAddress{
												Protocol: core.TCP,
												Address:  "172.20.0.3",
												PortSpecifier: &core.SocketAddress_PortValue{
													PortValue: 8200,
												},
											},
										},
									},
								},
							},
						},
						endpoint.LbEndpoint{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &core.Address{
										Address: &core.Address_SocketAddress{
											SocketAddress: &core.SocketAddress{
												Protocol: core.TCP,
												Address:  "172.20.0.4",
												PortSpecifier: &core.SocketAddress_PortValue{
													PortValue: 8200,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// =================================================================================

	log.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version " + fmt.Sprint(version))
	snap := cache.NewSnapshot(fmt.Sprint(version), e, c, nil, l)

	config.SetSnapshot("id_1", snap)
	select {}

}
