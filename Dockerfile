FROM golang AS builder
ENV GOPROXY https://goproxy.cn,direct
ENV CGO_ENABLED 0
ENV GOOS linux
WORKDIR /build
COPY go.mod go.sum ./
RUN go generate ./...
RUN go mod download && go mod verify
COPY . .
RUN go build -o /plutus ./cmd/plutus/main.go

FROM scratch
ENV TZ Asia/Shanghai
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /plutus /plutus
ENTRYPOINT ["/plutus"]
