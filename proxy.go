/*
Copyright 2024 kube-hetzner

This file is dual-licensed under the MIT License and the Apache License, Version 2.0.
You may choose either license to govern your use of this file.

Apache License, Version 2.0 (the "Apache License"):
    You may not use this file except in compliance with the Apache License.
    You may obtain a copy of the Apache License at:
        http://www.apache.org/licenses/LICENSE-2.0

MIT License (the "MIT License"):
    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:
        The above copyright notice and this permission notice shall be included in
        all copies or substantial portions of the Software.

Unless required by applicable law or agreed to in writing, software distributed
under either license is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the respective licenses
for the specific language governing permissions and limitations under those licenses.
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
