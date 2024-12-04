# SPDX-License-Identifier: MIT OR Apache-2.0
# This project is dual-licensed under the MIT License and the Apache License, Version 2.0.
# You may choose either license to govern your use of this project.
# See the LICENSE-MIT and LICENSE-Apache files for details.

FROM docker.io/library/golang:1.23.4-bookworm@sha256:ef30001eeadd12890c7737c26f3be5b3a8479ccdcdc553b999c84879875a27ce AS builder
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

LABEL org.opencontainers.image.title="boringproxy"
LABEL org.opencontainers.image.description="A boring HTTP proxy written in pure Go stdlib."
LABEL org.opencontainers.image.licenses="MIT AND Apache-2.0"
LABEL org.opencontainers.image.url="https://github.com/kube-hetzner/boringproxy"
LABEL org.opencontainers.image.source="https://github.com/kube-hetzner/boringproxy"
LABEL org.opencontainers.image.documentation="https://github.com/kube-hetzner/boringproxy#readme"
LABEL org.opencontainers.image.authors="aleksasiriski <sir@tmina.org>"
LABEL org.opencontainers.image.vendor="kube-hetzner"
