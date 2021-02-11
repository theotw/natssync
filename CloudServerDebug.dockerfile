FROM bmason42/fullstackdev:latest as build
WORKDIR /build
COPY ./ ./
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/:/root/go/bin
ENV GOSUMDB=off
ARG IMAGE_TAG=latest
RUN make incontainergenerate
RUN go build -gcflags "all=-N -l"  -v -o out/bridgeserver_x64_linux apps/bridge_server.go
RUN go get github.com/go-delve/delve/cmd/dlv
ENTRYPOINT ["dlv","--listen=:2345","--headless=true","--api-version=2","--accept-multiclient","exec" ,"out/bridgeserver_x64_linux"]



