
FROM alpine:3.6 as minimal

RUN apk add -U --no-cache ca-certificates && \
    echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

FROM scratch

COPY --from=minimal /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=minimal /etc_passwd /etc/passwd
COPY --chown=65534:65534 artifacts/orbctl-Linux-x86_64 /orbctl

USER nobody

ENTRYPOINT [ "/orbctl" ]
CMD [ "--help" ]
