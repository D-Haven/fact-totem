FROM goboring/golang:1.15.6b5 AS builder

WORKDIR /go/src/github.com/D-Haven/spa-server/
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go get -d -v
RUN PROJECT=d-haven.org/spa-server && \
    RELEASE=$(git describe --tags | sed 's/release\/\([0-9.]\+\)/\1/g') && \
    COMMIT=$(git rev-parse --short HEAD) && \
    BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%S') && \
    go build -ldflags "-X ${PROJECT}/version.Release=${RELEASE} \
                -X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
             	-a -o /go/bin/spa-server

RUN go test ./... -cover

##########################################

FROM scratch

COPY --from=builder /go/bin/spa-server /

EXPOSE 8443/tcp
ENTRYPOINT ["/fact-totem"]