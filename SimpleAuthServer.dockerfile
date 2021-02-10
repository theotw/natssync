FROM golang:1.14.0 as build
WORKDIR /build
COPY ./ ./
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
RUN make buildlinux

#FROM alpine:3.12.0
FROM scratch
COPY --from=build /build/out/simple_auth_x64_linux ./simple_auth_x64_linux
ENTRYPOINT ["./simple_auth_x64_linux"]

