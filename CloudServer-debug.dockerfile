FROM bmason42/fullstackdev:latest as build
WORKDIR /build
COPY ./ ./

ENV GOSUMDB=off
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/
RUN make incontainergenerate
RUN rm -r -f out & mkdir -p out & mkdir -p webout & mkdir -p /certs
RUN make buildlinux

ENV GIN_MODE=debug
ENTRYPOINT ["/build/scripts/run_debug_server.sh"]