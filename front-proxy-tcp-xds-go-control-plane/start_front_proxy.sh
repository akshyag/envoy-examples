#!/bin/sh
/usr/local/bin/go-control-plane &
/usr/local/bin/envoy -c /etc/front-envoy.yaml --service-cluster front-proxy
