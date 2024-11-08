FROM docker.io/library/golang:1.23.3-bookworm@sha256:0e3377d7a71c1fcb31cdc3215292712e83baec44e4792aeaa75e503cfcae16ec AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o proxy .

FROM scratch
COPY --from=builder /app/proxy /proxy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 65532
ENTRYPOINT ["/proxy"]
EXPOSE 31280
EXPOSE 31281

LABEL org.opencontainers.image.source="https://github.com/kube-hetzner/boringproxy"
