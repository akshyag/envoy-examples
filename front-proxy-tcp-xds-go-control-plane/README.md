## How To Run The Example

In this example we run a tcp front proxy that load-balances between 2 endpoints. endpoints, listeners and clusters will be dynamic resources and will be given by the the go-control-plane
The difference from front-proxy-tcp is that instead of statically giving the values, the endpoints, clusters and listeners are provided through the go-control-plane xds server. The built binary has been added to the repo. (Note: Files and instructions to build the binary have also been added).

The only static config was the address for the xds-server.

The dynamic config structures can be seen in the file go-control-plane-steps/main.go


1. Build the containers and start the services and front-proxy

```sh
$ docker-compose pull
$ docker-compose up --build -d
```

2. Verify that the containers started and ports were exposed

```sh
$docker-compose ps

                  Name                                Command               State                    Ports
---------------------------------------------------------------------------------------------------------------------------
front-proxy-tcp-xds-go-control-            /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:5678->5678/tcp,
plane_front-envoy_1                                                                 0.0.0.0:8001->8001/tcp,
                                                                                    0.0.0.0:8000->8200/tcp
front-proxy-tcp-xds-go-control-            /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8101->8200/tcp,
plane_service1_1                                                                    8201/tcp
front-proxy-tcp-xds-go-control-            /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8102->8200/tcp,
plane_service2_1                                                                    8202/tcp
```

3. Now first test the individual services

```sh
$ echo -n "test out the server" | nc 0.0.0.0 8102
Hello from behind Envoy (service 2)! hostname: ab0a93cc48a3 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8101
Hello from behind Envoy (service 1)! hostname: 5cd66dd26b03 data got: test out the server
```

4. Finally check out the load-balancing.

```sh
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: ab0a93cc48a3 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5cd66dd26b03 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5cd66dd26b03 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: ab0a93cc48a3 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: ab0a93cc48a3 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5cd66dd26b03 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5cd66dd26b03 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: ab0a93cc48a3 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: ab0a93cc48a3 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5cd66dd26b03 data got: test out the server
```

5. You can also see the stats at the access_logs endpoint: http://localhost:8001

