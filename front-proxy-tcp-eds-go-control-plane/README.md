## How To Run The Example

In this example we run a tcp front proxy that load-balances between 2 endpoints.
The difference from front-proxy-tcp is that instead of statically giving the endpoints, the endpoints are provided through the go-control-plane eds server. The built binary has been added to the repo. (Note: Files and instructions to build the binary have also been added) 
The structure to provide the endpoints in the go-control-plane EDS was as follows:
```sh
    e := []cache.Resource{
        &v2.ClusterLoadAssignment{
            ClusterName: service_name,
            Endpoints: []endpoint.LocalityLbEndpoints{
                {
                    Locality: nil,
                    LbEndpoints: []endpoint.LbEndpoint{
                        {
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
                        {
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

```

Note that the address needs to be given as static ip addresses. STRICT_DNS won't work as with the front-proxy-tcp example.


1. Build the containers and start the services and front-proxy

```sh
$ docker-compose pull
$ docker-compose up --build -d
```

2. Verify that the containers started and ports were exposed

```sh
$docker-compose ps
                 Name                                Command               State                    Ports
--------------------------------------------------------------------------------------------------------------------------
front-proxy-tcp-eds-go-control-           /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:5678->5678/tcp,
plane_front-envoy_1                                                                0.0.0.0:8001->8001/tcp,
                                                                                   0.0.0.0:8000->8200/tcp
front-proxy-tcp-eds-go-control-           /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8101->8200/tcp,
plane_service1_1                                                                   8201/tcp
front-proxy-tcp-eds-go-control-           /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8102->8200/tcp,
plane_service2_1                                                                   8202/tcp
```

3. Now first test the individual services

```sh
$ echo -n "test out the server" | nc 0.0.0.0 8102
Hello from behind Envoy (service 2)! hostname: 2f8a1783e501 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8101
Hello from behind Envoy (service 1)! hostname: 5df424475410 data got: test out the server
```

4. Finally check out the load-balancing.

```sh
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 1)! hostname: c44ea05f5e4f data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 2)! hostname: 3dbd550937ba data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 1)! hostname: c44ea05f5e4f data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 1)! hostname: c44ea05f5e4f data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 2)! hostname: 3dbd550937ba data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 2)! hostname: 3dbd550937ba data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 1)! hostname: c44ea05f5e4f data got: test out the server
$ echo -n "test out the server" | nc localhost 8000
Hello from behind Envoy (service 2)! hostname: 3dbd550937ba data got: test out the server
```

5. You can also see the stats at the access_logs endpoint: http://localhost:8001

