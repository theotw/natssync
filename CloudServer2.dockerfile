#FROM openjdk:8-jre-alpine as openapigenerate
ARG BASE_IMAGE_TAG=latest
FROM theotw/natssync-tests:$BASE_IMAGE_TAG as build

#FROM alpine:3.12.0
FROM scratch
COPY --from=build /build/out/bridgeserver_x64_linux ./bridgeserver_x64_linux
COPY --from=build /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=build /build/openapi/bridge_server_v1.yaml ./openapi/
COPY --from=build /build/BUILD_DATE ./
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeserver_x64_linux"]

