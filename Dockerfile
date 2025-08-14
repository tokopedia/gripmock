FROM golang:1.24-alpine3.22 AS builder

ARG version

COPY . /gripmock-src

WORKDIR /gripmock-src

RUN apk add --no-cache binutils \
    && go build -o /usr/local/bin/gripmock . \
    && strip /usr/local/bin/gripmock \
    && apk del binutils \
    && rm -rf /root/.cache /go/pkg /tmp/* /var/cache/*

RUN chmod +x /gripmock-src/entrypoint.sh && chmod +x /usr/local/bin/gripmock

FROM alpine:3.22

COPY --from=builder /usr/local/bin/gripmock /usr/local/bin/gripmock
COPY --from=builder /gripmock-src/entrypoint.sh /entrypoint.sh

EXPOSE 4770 4771

HEALTHCHECK --start-interval=1s --start-period=30s \
    CMD gripmock check --silent

ENTRYPOINT ["/entrypoint.sh"]
