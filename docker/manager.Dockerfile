FROM golang:1.22.0-alpine3.19 AS builder
WORKDIR /build
COPY . .
RUN go get -d -v ./...
RUN go build -ldflags="-w -s" -o /go/bin/manager /build/cmd/gotama-manager/main.go
FROM scratch
COPY --from=builder /go/bin/manager /go/bin/manager
ENTRYPOINT ["/go/bin/manager"]
