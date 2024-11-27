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
		w.Header().Set("Proxy-Authenticate", "Basic realm=\"Proxy\"")
		http.Error(w, "ProxyAuthRequired", http.StatusProxyAuthRequired)
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
