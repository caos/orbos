FROM golang:1.14.4-alpine3.11 as build

RUN apk add -U --no-cache ca-certificates git openssh && \
    go get github.com/go-delve/delve/cmd/dlv

COPY artifacts/orbctl-Linux-x86_64 /orbctl

ENTRYPOINT [ "dlv", "exec", "/orbctl", "--api-version", "2", "--headless", "--listen", "127.0.0.1:2345", "--" ]

FROM python:3.8.3-alpine3.11 as prod

RUN apk add openssh && mkdir -p /home/nobody && chown -R 65534:65534 /home/nobody

ENV GODEBUG="madvdontneed=1" \
    HOME="/home/nobody"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build --chown=65534:65534 /orbctl /orbctl

USER nobody

ENTRYPOINT [ "/orbctl" ]
CMD [ "--help" ]
