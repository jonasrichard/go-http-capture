FROM golang:bookworm

RUN apt update && apt install -y flex yacc

COPY . /src

WORKDIR /

RUN wget http://www.tcpdump.org/release/libpcap-1.5.3.tar.gz && \
    tar xvf libpcap-1.5.3.tar.gz && \
    cd libpcap-1.5.3 && \
    ./configure --with-pcap=linux && \
    make install

WORKDIR /src

RUN CGO_ENABLED=1 go build --ldflags "-L /usr/local/lib -linkmode external -extldflags '-static'" -o target/capture .
#CGO_ENABLED=0 go build --ldflags "-L /libpcap-1.5.3 -linkmode external -extldflags \"static\"" -o target/capture main.go

