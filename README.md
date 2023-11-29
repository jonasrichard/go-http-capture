# go-http-capture

HTTP capture tool with on the fly filtering

## Build for CentOS

The binary statically links the specified verion of pcap library, so on the host OS
you don't need to install it.

```
docker run -ti --rm -v $PWD:/src golang:bookworm

# in the container

cd /src
CGO_ENABLED=1 go build --ldflags "-L /usr/local/lib -linkmode external -extldflags '-static'" -o target/capture .
```
