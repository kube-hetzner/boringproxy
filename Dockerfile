FROM docker.io/library/golang:1.23.3-bookworm@sha256:1f001ad8c8d90281cd9d6e0ae4a40363039c148c503bcd483ff38c946b3d4f6d AS builder
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
