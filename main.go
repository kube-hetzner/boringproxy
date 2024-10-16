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
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	_ "time/tzdata" // Required for Docker Scratch image (timezone data)
)

const (
	debugDefault           = false
	portProxyDefault       = "31280"
	portProbesDefault      = "31281"
	shutdownTimeoutDefault = 5 * time.Second
	readinessURLDefault    = "https://cloudflare.com/cdn-cgi/trace"
	usernameDefault        = "proxy"
	passwordDefault        = "secret"
)

func main() {
	// Load env vars
	debugS := getEnv("DEBUG", strconv.FormatBool(debugDefault))
	portProxy := getEnv("PORT_PROXY", portProxyDefault)
	portProbes := getEnv("PORT_PROBES", portProbesDefault)
	shutdownTimeoutS := getEnv("SHUTDOWN_TIMEOUT", shutdownTimeoutDefault.String())
	readinessURL := getEnv("READINESS_URL", readinessURLDefault)
	username := getEnv("USERNAME", usernameDefault)
	password := getEnv("PASSWORD", passwordDefault)

	// Parse the debug value
	debug, err := strconv.ParseBool(debugS)
	if err != nil {
		slog.Error("Invalid debug value", "error", err)
	}

	// Create the default logger
	if debug {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}

	// Parse the shutdown timeout
	shutdownTimeout, err := time.ParseDuration(shutdownTimeoutS)
	if err != nil {
		slog.Error("Invalid shutdown timeout", "error", err)
		os.Exit(1)
	}

	// Create the credetials struct
	creds := credentials{
		username,
		password,
	}

	// Create a mux for the probes server
	muxProbes := http.NewServeMux()

	// Register routes for probes
	muxProbes.HandleFunc("/healthz", handleLiveness)
	muxProbes.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		handleReadiness(w, r, readinessURL)
	})

	// Define the HTTP servers
	server := http.Server{
		Addr: ":" + portProxy,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleProxy(w, r, creds)
		}), // Required instead of mux to pass CONNECT requests properly
	}
	serverProbes := http.Server{
		Addr:    ":" + portProbes,
		Handler: muxProbes,
	}

	// Channel to listen for system signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	// Run the servers in goroutines
	go func() {
		slog.Info("Starting proxy server", "port", portProxy)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Proxy: ListenAndServe", "error", err)
			os.Exit(1)
		}
	}()
	go func() {
		slog.Info("Starting probes server", "port", portProbes)
		if err := serverProbes.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Probes: ListenAndServe", "error", err)
			os.Exit(1)
		}
	}()

	// Block until we receive a signal
	<-stop

	// Gracefully shutdown the server, waiting for ongoing requests to finish
	slog.Info("Server shutting down gracefully", "timeout", shutdownTimeout.String())
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Sync the shutdown of both servers
	wg, ok := sync.WaitGroup{}, atomic.Bool{}
	wg.Add(2)
	ok.Store(true)

	// Shutdown both servers
	go func() {
		defer wg.Done()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Proxy server forced to shutdown", "error", err)
			ok.Store(false)
		}
	}()
	go func() {
		defer wg.Done()
		if err := serverProbes.Shutdown(ctx); err != nil {
			slog.Error("Probes server forced to shutdown", "error", err)
			ok.Store(false)
		}
	}()

	// Wait for both servers to shutdown
	wg.Wait()

	// Check if the shutdown was successful for both servers
	if !ok.Load() {
		os.Exit(1)
	}

	slog.Info("Server shutdown successfully")
}
