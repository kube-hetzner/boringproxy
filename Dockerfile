FROM docker.io/library/golang:1.23.2-bookworm@sha256:345d5e81c88be2c500edf00ed1dca6be656e4485cd79e4e0bcc73a90361910e0 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o proxy .

FROM scratch
COPY --from=builder /app/proxy /proxy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 31280
# EXPOSE 31281 # Probes port, not exposed by default
ENTRYPOINT ["/proxy"]
