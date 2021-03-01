#FROM openjdk:8-jre-alpine as openapigenerate
FROM bmason42/fullstackdev:latest as build
LABEL stage=build
WORKDIR /build
RUN mkdir -p web
RUN mkdir -p webout
COPY ./ ./
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/
ENV GOSUMDB=off
ARG IMAGE_REPO=theotw
ARG IMAGE_TAG=latest
RUN make incontainergenerate
RUN make buildlinux
RUN date -uIseconds > ./BUILD_DATE

# Bridge server
FROM scratch as natssync-server
WORKDIR /build
COPY --from=build /build/web /build/web
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE
COPY --from=build /build/out/bridgeserver_x64_linux ./bridgeserver_x64_linux
COPY --from=build /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=build /build/openapi/bridge_server_v1.yaml ./openapi/
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeserver_x64_linux"]

# Bridge client
FROM scratch as natssync-client
WORKDIR /build
COPY --from=build /build/webout /build/webout
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE
COPY --from=build /build/out/bridgeclient_x64_linux ./bridgeclient_x64_linux
COPY --from=build /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=build /build/openapi/bridge_client_v1.yaml ./openapi/
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeclient_x64_linux"]

# Cloudserver debug
FROM build as debugnatssync-server
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE
RUN go build -gcflags "all=-N -l"  -v -o out/bridgeserver_x64_linux apps/bridge_server.go
RUN go get github.com/go-delve/delve/cmd/dlv
ENTRYPOINT ["dlv","--listen=:2345","--headless=true","--api-version=2","--accept-multiclient","exec" ,"out/bridgeserver_x64_linux"]

# Echo proxylet
FROM scratch as echo-proxylet
COPY --from=build /build/out/echo_main_x64_linux ./echo_main_x64_linux
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE
ENTRYPOINT ["./echo_main_x64_linux"]

# Test image
FROM build as natssync-tests
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE
RUN rm -r -f out & mkdir -p out & mkdir -p webout & mkdir -p /certs
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE

# Simple auth
FROM scratch as simple-reg-auth
COPY --from=build /build/BUILD_DATE /build/BUILD_DATE
COPY --from=build /build/out/simple_auth_x64_linux ./simple_auth_x64_linux
ENTRYPOINT ["./simple_auth_x64_linux"]