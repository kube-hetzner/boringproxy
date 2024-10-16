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
	"log/slog"
	"net/http"
)

// Liveness probe handler (returns 200 OK)
func handleLiveness(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Liveness", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness probe handler (checks internet connectivity by making a HTTP GET request to the specified URL)
func handleReadiness(w http.ResponseWriter, r *http.Request, readinessURL string) {
	slog.Debug("Readiness", "remote_addr", r.RemoteAddr, "host", r.Host, "proto", r.Proto, "method", r.Method, "url", r.URL.String())
	if _, err := http.Get(readinessURL); err != nil {
		slog.Error("Readiness", "error", err, "url", readinessURL)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
