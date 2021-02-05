#FROM openjdk:8-jre-alpine as openapigenerate
FROM bmason42/fullstackdev:latest as build
LABEL stage=build
WORKDIR /build
COPY ./ ./
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/
ENV GOSUMDB=off
ARG IMAGE_REPO=theotw
ARG IMAGE_TAG=latest
RUN make incontainergenerate
RUN make buildlinux

FROM alpine:3.12.0
#FROM scratch
RUN mkdir -p web
COPY --from=build /build/out/bridgeserver_x64_linux ./bridgeserver_x64_linux
COPY --from=build /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=build /build/openapi/bridge_server_v1.yaml ./openapi/
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeserver_x64_linux"]

