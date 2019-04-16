## How To Run The Example

In this example we run a tcp front proxy that load-balances between 2 endpoints

1. Build the containers and start the services and front-proxy

```sh
$ docker-compose pull
$ docker-compose up --build -d
```

2. Verify that the containers started and ports were exposed

```sh
$ docker-compose ps
            Name                           Command               State                             Ports
----------------------------------------------------------------------------------------------------------------------------------
front-proxy-tcp_front-envoy_1   /docker-entrypoint.sh /bin ...   Up      10000/tcp, 0.0.0.0:8001->8001/tcp, 0.0.0.0:8000->8200/tcp
front-proxy-tcp_service1_1      /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8101->8200/tcp
front-proxy-tcp_service2_1      /bin/sh -c /usr/local/bin/ ...   Up      10000/tcp, 0.0.0.0:8102->8200/tcp
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
Hello from behind Envoy (service 2)! hostname: 2f8a1783e501 data got: test out the server
$echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5df424475410 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5df424475410 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: 2f8a1783e501 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: 2f8a1783e501 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5df424475410 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 1)! hostname: 5df424475410 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: 2f8a1783e501 data got: test out the server
$ echo -n "test out the server" | nc 0.0.0.0 8000
Hello from behind Envoy (service 2)! hostname: 2f8a1783e501 data got: test out the server
```

