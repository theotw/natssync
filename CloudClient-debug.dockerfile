FROM bmason42/fullstackdev:latest as build
WORKDIR /build
COPY ./ ./

ENV GOSUMDB=off
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/
RUN make incontainergenerate
RUN make buildlinux

RUN mkdir -p webout & mkdir -p /certs
ENV GIN_MODE=debug
ENTRYPOINT ["/buid/scripts/run_debug_client.sh"]