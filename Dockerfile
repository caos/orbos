
FROM golang:1.14.0-alpine3.11 as build

RUN apk add -U --no-cache ca-certificates git && \
    go get github.com/go-delve/delve/cmd/dlv

# Runtime dependencies
RUN apk update && apk add git curl && \
    mkdir /dependencies && \
    curl -L "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.4.0/kustomize_v3.4.0_linux_amd64.tar.gz" |tar xvz && \
    mv ./kustomize /dependencies/kustomize && \
    curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /dependencies/kubectl && \
    curl -L "https://get.helm.sh/helm-v2.12.0-linux-amd64.tar.gz" |tar xvz && \
    mv linux-amd64/helm /dependencies/helm && \
    chmod +x /dependencies/helm

COPY artifacts/orbctl-Linux-x86_64 /orbctl
COPY artifacts/gen-charts /gen-charts

RUN cp /dependencies/helm /usr/local/bin/helm
RUN mkdir -p /boom
RUN /gen-charts -basepath "/boom"

COPY dashboards /boom/dashboards

ENTRYPOINT [ "dlv", "exec", "/orbctl", "--api-version", "2", "--headless", "--listen", "127.0.0.1:5000", "--accept-multiclient", "--" ]

FROM alpine:3.11 as prod

RUN adduser nobody nobody

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build --chown=nobody:nobody /orbctl /orbctl
COPY --from=build --chown=nobody:nobody /boom /boom
COPY --from=build --chown=nobody:nobody /dependencies /usr/local/bin/

USER nobody

ENTRYPOINT [ "/orbctl" ]
CMD [ "--help" ]
