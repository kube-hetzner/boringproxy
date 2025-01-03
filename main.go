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
)

func main() {
	// Check for the "--exit" argument
	for _, arg := range os.Args[1:] {
		if arg == "--exit" {
			slog.Info("Exiting as requested by --exit argument")
			os.Exit(0)
		}
	}

	// Load env vars
	debugS := getEnv("DEBUG", strconv.FormatBool(debugDefault))
	portProxy := getEnv("PORT_PROXY", portProxyDefault)
	portProbes := getEnv("PORT_PROBES", portProbesDefault)
	shutdownTimeoutS := getEnv("SHUTDOWN_TIMEOUT", shutdownTimeoutDefault.String())
	readinessURL := getEnv("READINESS_URL", readinessURLDefault)
	username := getEnv("USERNAME")
	password := getEnv("PASSWORD")

	// Check if the username and password are set
	if username == "" || password == "" {
		slog.Error("Username and password must be set")
		os.Exit(1)
	}

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
			panic(err)
		}
	}()
	go func() {
		slog.Info("Starting probes server", "port", portProbes)
		if err := serverProbes.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Probes: ListenAndServe", "error", err)
			panic(err)
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
