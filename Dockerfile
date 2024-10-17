FROM docker.io/library/golang:1.23.2@sha256:cc637ce72c1db9586bd461cc5882df5a1c06232fd5dfe211d3b32f79c5a999fc AS builder
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
# EXPOSE 31281 # Probes port, not exposed by default

LABEL org.opencontainers.image.source="https://github.com/kube-hetzner/boringproxy"
