/*
Copyright 2024 kube-hetzner

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"io"
	"log/slog"
	"net"
	"net/http"
)

// Proxy handler
func handleProxy(w http.ResponseWriter, r *http.Request, creds credentials) {
	// Check for basic authentication
	if !checkBasicAuth(w, r, creds) {
		slog.Warn("Proxy (unauthorized)", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
		return // Unauthorized response already sent
	}

	// Handle the request (HTTP or HTTPS)
	slog.Info("Proxy", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
	if r.Method == http.MethodConnect {
		handleConnect(w, r)
	} else {
		handleHTTP(w, r)
	}
}

// Handle HTTPS tunneling via CONNECT method
func handleConnect(w http.ResponseWriter, r *http.Request) {
	// Dial the destination server
	destConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	// Send a 200 OK status to the client
	w.WriteHeader(http.StatusOK)

	// Flush the response headers
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush() // Ensure the 200 OK is sent before hijacking
	}

	// Hijack the connection to establish a tunnel
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Transfer data between client and destination
	dcChan := make(chan struct{})
	go func() {
		defer close(dcChan)
		io.Copy(destConn, clientConn)
	}()
	cdChan := make(chan struct{})
	go func() {
		defer close(cdChan)
		io.Copy(clientConn, destConn)
	}()

	// Wait until either end closes the connection
	select {
	case <-dcChan:
	case <-cdChan:
	}
}

// Handle HTTP requests (GET, POST, etc.)
func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Modify the request URL to point to the destination server
	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = r.Host

	// Forward the request to the destination server
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers and body to the client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
