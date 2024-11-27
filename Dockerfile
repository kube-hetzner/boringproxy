FROM docker.io/library/golang:1.23.3-bookworm@sha256:3f3b9daa3de608f3e869cd2ff8baf21555cf0fca9fd34251b8f340f9b7c30ec5 AS builder
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
