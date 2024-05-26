FROM alpine

ENV GID 1000
ENV UID 1000

# Curl is used for healthcheck
RUN apk add curl

COPY dnsr /usr/bin

ENTRYPOINT ["/usr/bin/dnsr"]