# Base image
FROM natssync-base:latest as base

# Test image
FROM natssync-base:latest as natssync-tests
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE
RUN rm -r -f out & mkdir -p out & mkdir -p webout & mkdir -p /certs
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE

# Bridge server
FROM scratch as natssync-server
WORKDIR /build
COPY --from=base /build/web /build/web
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE
COPY --from=base /build/out/bridgeserver_x64_linux ./bridgeserver_x64_linux
COPY --from=base /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=base /build/openapi/bridge_server_v1.yaml ./openapi/
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeserver_x64_linux"]

# Bridge client
FROM scratch as natssync-client
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
WORKDIR /build
COPY --from=base /build/webout /build/webout
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE
COPY --from=base /build/out/bridgeclient_x64_linux ./bridgeclient_x64_linux
COPY --from=base /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=base /build/openapi/bridge_client_v1.yaml ./openapi/
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeclient_x64_linux"]

# Cloudserver debug
FROM natssync-base:latest as debugnatssync-server
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE
RUN go build -gcflags "all=-N -l"  -v -o out/bridgeserver_x64_linux apps/bridge_server.go
RUN go get github.com/go-delve/delve/cmd/dlv
ENTRYPOINT ["dlv","--listen=:2345","--headless=true","--api-version=2","--accept-multiclient","exec" ,"out/bridgeserver_x64_linux"]

# Echo proxylet
FROM scratch as echo-proxylet
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
COPY --from=base /build/out/echo_main_x64_linux ./echo_main_x64_linux
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE
ENTRYPOINT ["./echo_main_x64_linux"]

# Simple auth
FROM scratch as simple-reg-auth
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
COPY --from=base /build/BUILD_DATE /build/BUILD_DATE
COPY --from=base /build/out/simple_auth_x64_linux ./simple_auth_x64_linux
ENTRYPOINT ["./simple_auth_x64_linux"]