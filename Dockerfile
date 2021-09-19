FROM goboring/golang:1.16.7b7 AS builder

WORKDIR /go/src/github.com/D-Haven/fact-totem/
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go get -d -v
RUN PROJECT=github.com/D-Haven/fact-totem && \
    RELEASE=$(git describe --tags | sed 's/release\/\([0-9.]\+\)/\1/g') && \
    COMMIT=$(git rev-parse --short HEAD) && \
    BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%S') && \
    go build -ldflags "-X ${PROJECT}/version.Release=${RELEASE} \
                -X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
             	-a -o /go/bin/fact-totem

RUN go test ./... -cover

##########################################

FROM scratch

COPY --from=builder /go/bin/fact-totem /

USER 1001

EXPOSE 8443/tcp
ENTRYPOINT ["/fact-totem"]