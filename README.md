# Boring Proxy - HTTP Proxy for Kubernetes Network Integration

This project provides an HTTP proxy designed to integrate host operating system (OS) network traffic into a Kubernetes cluster. By routing OS traffic through this proxy, users can inspect and control the traffic using their preferred Kubernetes CNI (e.g. Cilium). This allows for enhanced security, observability, and control over network traffic flowing through the cluster.

## Features

- **Proxy traffic**: Forward both HTTP and HTTPS traffic through Kubernetes.
- **Basic Authentication**: Secure the proxy with Basic Authentication using the `Proxy-Authorization` header.
- **Kubernetes Readiness and Liveness Probes**: Ensure the proxy is running and ready for traffic with dedicated health endpoints.

## Environment Variables

The proxy is highly configurable using the following environment variables:

| Environment Variable | Default Value                          | Description                                                          |
| -------------------- | -------------------------------------- | -------------------------------------------------------------------- |
| `PORT_PROXY`         | `31280`                                | The port on which the proxy server listens for traffic.              |
| `PORT_PROBES`        | `31281`                                | The port on which the probe server (health and readiness) listens.   |
| `SHUTDOWN_TIMEOUT`   | `5`                                    | The timeout duration in seconds for graceful shutdown of the server. |
| `READINESS_URL`      | `https://cloudflare.com/cdn-cgi/trace` | The URL to check for network readiness (internet connection check).  |
| `USERNAME`           | `proxy`                                | The username for Basic Authentication on the proxy.                  |
| `PASSWORD`           | `secret`                               | The password for Basic Authentication on the proxy.                  |

## Usage

This proxy is designed to run within a **Kubernetes cluster** to integrate OS-level network traffic into the Kubernetes network. By directing OS traffic through this proxy, users can leverage the capabilities of their chosen CNI (e.g., Cilium) to inspect, filter, or control the network traffic.

### Deployment in Kubernetes

1. **Create a Kubernetes Deployment**:
   You can deploy the proxy in a Kubernetes cluster with a configuration that routes OS traffic into the proxy service.

2. **Integrate with the Host's OS Traffic**:
   Configure your host or containers to send their network traffic through the proxy service. You may choose to redirect traffic using iptables or other routing mechanisms.

3. **Leverage the CNI**:
   Once OS traffic is routed through the proxy, your CNI plugin (e.g., Cilium) can inspect and control the traffic. This allows you to apply network policies, observe traffic flows, or implement security measures directly at the OS level.

### Example `curl` Command with Proxy Authentication

To test the proxy and ensure it's working as expected with Basic Authentication:

```bash
curl -x "http://localhost:31280" -U "proxy:secret" "http://example.com"
```

## Acknowledgements

This project is heavily inspired by [dumbproxy](https://github.com/SenseUnit/dumbproxy).

## License

This project is licensed under the Apache v2.0 License. See the [LICENSE](LICENSE) file for details.
