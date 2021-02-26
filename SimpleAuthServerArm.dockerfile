FROM golang:1.14.0 as build
WORKDIR /build
COPY ./ ./
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
RUN make buildarm

#FROM alpine:3.12.0
FROM scratch
COPY --from=build /build/out/simple_auth_x86_linux_arm ./simple_auth_x86_linux_arm
ENTRYPOINT ["./simple_auth_x86_linux_arm"]

