## How To Run The Example

In this example we run a tcp front proxy that load-balances between 2 endpoints.
The difference from front-proxy-tcp is that instead of statically giving the endpoints, the endpoints are provided through a file. The file can be changed on the fly to add/remove endpoints.
The format of eds.conf is as follows:
```sh
{
  "version_info": "0",
  "resources": [{
    "@type": "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment",
    "cluster_name": "serviceendpoints",
    "endpoints": [
      {
        "locality": {},
        "lb_endpoints": [
          {
            "endpoint": {
              "address": {
                "socket_address": {
                  "address": "172.20.0.3",
                  "port_value": 8200
                }
              }
            }
          },
          {
            "endpoint": {
              "address": {
                "socket_address": {
                  "address": "172.20.0.4",
                  "port_value": 8200
                }
              }
            }
          }
        ]
      }
    ]
  }]
}
```
Note that the address needs to be given as static ip addresses. STRICT_DNS won't work as with the front-proxy-tcp example. Also due to the way the structures are defined, an empty value for locality had to be provided otherwise all the traffic was being sent to all endpoints.

Note: In order for the envoy process to pick the change, just editing the file may not be enough, so we may have to rename the file and rename it back. For ex, let the name of file be eds.conf

```sh
$ mv /etc/envoy/eds.conf tmp; mv tmp /etc/envoy/eds.conf
```

1. Build the containers and start the services and front-proxy

```sh
$ docker-compose pull
$ docker-compose up --build -d
```

2. Verify that the containers started and ports were exposed

```sh
$ docker-compose ps
                    Name                                  Command               State                             Ports
-------------------------------------------------------------------------------------------------------------------------------------------------
front-proxy-tcp-eds-file-based_front-envoy_1   /docker-entrypoint.sh /bin ...   Up      10000/tcp, 0.0.0.0:8001->8001/tcp, 0.0.0.0:8000->8200/tcp
front-proxy-tcp-eds-file-based_service1_1      /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8101->8200/tcp, 8201/tcp
front-proxy-tcp-eds-file-based_service2_1      /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8102->8200/tcp, 8202/tcp
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
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 9a140863c4cb data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: d9e0cecd6401 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 9a140863c4cb data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: d9e0cecd6401 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: d9e0cecd6401 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 9a140863c4cb data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: d9e0cecd6401 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 9a140863c4cb data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 9a140863c4cb data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: d9e0cecd6401 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 9a140863c4cb data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: d9e0cecd6401 data got: test out the server
```

