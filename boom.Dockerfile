####################################################################################################
# Download dependencies and build
####################################################################################################
FROM golang:1.14.2-alpine3.11 AS dependencies

WORKDIR $GOPATH/src/github.com/caos/orbos

# Runtime dependencies
RUN apk update && apk add git curl && \
    mkdir /artifacts && \
    curl -L "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.4.0/kustomize_v3.4.0_linux_amd64.tar.gz" |tar xvz && \
    mv ./kustomize /artifacts/kustomize && \
    curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /artifacts/kubectl && \
    curl -L "https://get.helm.sh/helm-v2.12.0-linux-amd64.tar.gz" |tar xvz && \
    mv linux-amd64/helm /artifacts/helm && \
    chmod +x /artifacts/helm && \
    go get -u github.com/go-delve/delve/cmd/dlv

# copy all sourcecode from the current repository
COPY ./go.mod ./go.sum ./
RUN go mod download

# Copy the go source
COPY cmd cmd
COPY mntr mntr
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /gen cmd/gen-charts/*.go

# ####################################################################################################
# Create base runtime
# ####################################################################################################
FROM alpine:3.11 AS runtime

RUN apk update && apk add bash ca-certificates
COPY --from=dependencies /artifacts /usr/local/bin/
COPY --from=dependencies /gen /

COPY dashboards /dashboards

RUN /gen

# ####################################################################################################
# Build debug binary
# ####################################################################################################
FROM dependencies AS debug-build

# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o boom main.go
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags 'all=-N -l' -o /boom cmd/boom/*.go

# ####################################################################################################
# Create debug runtime
# ####################################################################################################
FROM runtime AS debug
COPY --from=dependencies /go/bin/dlv /usr/local/bin/
COPY --from=debug-build /boom /

ENTRYPOINT [ "dlv", "exec", "--api-version", "2", "--headless", "--accept-multiclient", "--listen", ":2345", "/boom", "--"]

# ####################################################################################################
# Run tests
# ####################################################################################################
FROM dependencies AS test

# RUN CGO_ENABLED=0 GOOS=linux go test -short $(go list ./... | grep -v /vendor/)
# RUN go test -race -short $(go list ./... | grep -v /vendor/)
# RUN go test -msan -short $(go list ./... | grep -v /vendor/)

# ####################################################################################################
# Run build
# ####################################################################################################
FROM dependencies AS build

# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o boom main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /boom cmd/boom/*.go

# ####################################################################################################
# Create production runtime
# ####################################################################################################
FROM runtime

COPY --from=build /boom /

ENTRYPOINT ["/boom"]