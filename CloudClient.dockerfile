FROM bmason42/fullstackdev:latest as build
LABEL stage=build
WORKDIR /build
COPY ./ ./

ENV GOSUMDB=off
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/
ARG IMAGE_TAG=latest
RUN make incontainergenerate
RUN make buildlinux

FROM alpine:3.12.0
#FROM scratch
RUN mkdir -p webout
COPY --from=build /build/out/bridgeclient_x64_linux ./bridgeclient_x64_linux
COPY --from=build /build/third_party/swaggerui/ ./third_party/swaggerui/
COPY --from=build /build/openapi/bridge_client_v1.yaml ./openapi/
RUN date -uIseconds > ./BUILD_DATE

ENV GIN_MODE=release
ENTRYPOINT ["./bridgeclient_x64_linux"]

