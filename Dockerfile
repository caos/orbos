FROM golang:1.12 AS build

RUN apt-get update && \
    apt-get install -y \
        ansible \
        apt-transport-https \
        curl \
        jq && \
    mkdir /artifacts

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

WORKDIR /src

ENV GO111MODULE=on \
    INFROP_ROOT=/src

COPY go.mod go.sum ./
RUN go mod download

ARG GIT_TAG
ARG GIT_COMMIT
RUN test -n "$GIT_TAG" && test -n "$GIT_COMMIT"
COPY . .

# go test --timeout 40m --race --cover --bench . ./... && \
RUN go run cmd/gen-executables/*.go -tag $GIT_TAG -commit $GIT_COMMIT && \
    CGO_ENABLED=0 go build -ldflags "-s -w -X main.gitCommit=$GIT_COMMIT -X main.gitTag=$GIT_TAG" -o /artifacts/orbiter ./cmd/orbiter

FROM scratch AS prod

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc_passwd /etc/passwd
COPY --from=build --chown=65534:65534 /artifacts/ /artifacts/

USER nobody

ENTRYPOINT [ "/artifacts/orbiter" ]
