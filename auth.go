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
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"
)

type credentials struct {
	username string
	password string
}

// Checks if the request has valid Basic Authentication
func checkBasicAuth(w http.ResponseWriter, r *http.Request, creds credentials) bool {
	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		slog.Debug("No Proxy-Authorization header found", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Expected format: "Basic base64(username:password)"
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		slog.Debug("Invalid Proxy-Authorization header", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Decode the base64 username:password
	payload, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		slog.Debug("Failed to decode Proxy-Authorization header", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Check if the credentials match
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 || parts[0] != creds.username || parts[1] != creds.password {
		slog.Debug("Invalid credentials", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Credentials are valid
	return true
}
