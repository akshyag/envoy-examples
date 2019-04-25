## How To Build And Use go-control-plane Binary

Steps for how to build the go-control-plane binary for the alpine container.

1. Start from the directory with the example
```sh
cd $GOPATH/src/github.com/envoy-examples/front-proxy-tcp-xds-go-control-plane
```

2. Checkout the go-control-plane repo
```sh
mkdir $GOPATH/src/github.com/envoyproxy
pushd $GOPATH/src/github.com/envoyproxy
git clone https://github.com/envoyproxy/go-control-plane.git
popd
```

3. We faced some issues in the way the proto files are processes. So we had to modify the build scripts. Replace the file in the go-control-plane repo with the one supplied
```sh
cp $GOPATH/src/github.com/envoy-examples/front-proxy-tcp-xds-go-control-plane/go-control-plane-steps/generate_protos.sh $GOPATH/src/github.com/envoyproxy/go-control-plane/build/
```

4. Copy the main.go in the examples dir to the go-control-plane repo. In case you want to play around with the go-control-plane configs, this file is the place to do so.
```sh
cp $GOPATH/src/github.com/envoy-examples/front-proxy-tcp-xds-go-control-plane/go-control-plane-steps/main.go $GOPATH/src/github.com/envoyproxy/go-control-plane/
```

5. Go to the go-control-plane repo, build and copy the built binary to the example folder
```sh
pushd $GOPATH/src/github.com/envoyproxy/go-control-plane
make tools
make depend.install
make generate
make format
make check
make build
make test
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo .
cp $GOPATH/src/github.com/envoyproxy/go-control-plane/go-control-plane $GOPATH/src/github.com/envoy-examples/front-proxy-tcp-xds-go-control-plane/
popd
```

6. Now that the binary has been built and copied to the right place, follow the steps in the README to run the example.
