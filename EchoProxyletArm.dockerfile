FROM golang:1.14.0 as build
WORKDIR /build
COPY ./ ./

ENV GOSUMDB=off
ARG IMAGE_TAG=latest
RUN make buildarm

#FROM alpine:3.12.0
FROM scratch
COPY --from=build /build/out/echo_main_x86_linux_arm ./echo_main_x86_linux_arm
ENTRYPOINT ["./echo_main_x86_linux_arm"]

