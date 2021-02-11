FROM bmason42/fullstackdev:latest as build
WORKDIR /build
COPY ./ ./
ARG IMAGE_TAG=latest
ENV GOSUMDB=off
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/
RUN make incontainergenerate
RUN rm -r -f out & mkdir -p out & mkdir -p webout & mkdir -p /certs
RUN make buildlinux

