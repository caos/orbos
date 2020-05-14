
FROM golang:1.14.0-alpine3.11 as build

RUN apk add -U --no-cache ca-certificates git && \
    echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd && \
    go get github.com/go-delve/delve/cmd/dlv

COPY artifacts/orbctl-Linux-x86_64 /orbctl

ENTRYPOINT [ "dlv", "exec", "/orbctl", "--api-version", "2", "--headless", "--listen", "127.0.0.1:5000", "--accept-multiclient", "--" ]

FROM scratch as prod

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc_passwd /etc/passwd
COPY --from=build --chown=65534:65534 /orbctl /orbctl

USER nobody

ENTRYPOINT [ "/orbctl" ]
CMD [ "--help" ]
