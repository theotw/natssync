#FROM openjdk:8-jre-alpine as openapigenerate
FROM raghu4026/fullstackdev:arm as build
LABEL stage=build
WORKDIR /build
COPY ./ ./
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin
ENV GOSUMDB=off
ARG IMAGE_REPO=raghu4026
ARG IMAGE_TAG=latest
RUN make incontainergenerate
RUN make buildarm

FROM alpine:3.12.0
#FROM scratch
RUN mkdir -p web
COPY --from=build /build/out/bridgeserver_x86_linux_arm ./bridgeserver_x86_linux_arm
COPY --from=build /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=build /build/openapi/bridge_server_v1.yaml ./openapi/
RUN date -uIseconds > ./BUILD_DATE
ENV GIN_MODE=release
ENTRYPOINT ["./bridgeserver_x86_linux_arm"]

