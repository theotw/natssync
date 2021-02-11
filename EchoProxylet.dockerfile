FROM golang:1.14.0 as build
WORKDIR /build
COPY ./ ./

ENV GOSUMDB=off
ARG IMAGE_TAG=latest
RUN make buildlinux

#FROM alpine:3.12.0
FROM scratch
COPY --from=build /build/out/echo_main_x64_linux ./echo_main_x64_linux
ENTRYPOINT ["./echo_main_x64_linux"]

