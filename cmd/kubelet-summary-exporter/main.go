/*
 * Copyright (c) 2022, salesforce.com, inc.
 * All rights reserved.
 * SPDX-License-Identifier: BSD-3-Clause
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
 */
package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/salesforce/kubelet-summary-exporter/pkg/scraper"
	"github.com/salesforce/kubelet-summary-exporter/pkg/utils"
)

type CLI struct {
	PromListen     string        `help:"Address to listen for for Prometheus metrics" default:":9091"`
	NodeHost       string        `help:"Address to request kubelet's stats/summary from" env:"NODE_HOST"`
	Insecure       bool          `help:"Don't validate certificates" env:"INSECURE" default:"false"`
	CA             string        `help:"Certificate location" env:"CA_CRT"`
	TokenPath      string        `help:"Token location" env:"TOKEN"`
	Timeout        time.Duration `help:"Timeout for requests" env:"TIMEOUT" default:"5s"`
	LookUpHostname bool          `help:"Use api-server to deterimine hostname (assumes in cluster config)" env:"LOOK_UP_HOSTNAME" default:"true"`
}

func main() {
	cli := &CLI{}
	_ = kong.Parse(cli)
	ctx := context.Background()

	logger := slog.Default().With(slog.String("app", "kubelet-stats-exporter"))

	if err := utils.ConfigureTLS(cli.CA, cli.Insecure, cli.NodeHost); err != nil {
		logger.Error("unable to configure tls", slog.Any("error", err))
		os.Exit(1)
	}

	if _, err := os.Stat(cli.TokenPath); os.IsNotExist(err) {
		logger.Error("token not found", slog.String("file", cli.TokenPath), slog.Any("error", err))
	}
	serverAddr := cli.NodeHost
	if cli.LookUpHostname {
		//Handle downward API using a node name that isn't identical to the node's Hostname
		name, err := utils.ServerAddrFromCluster(cli.NodeHost)
		if err != nil {
			logger.Error("failed to retrieve in node hostname", slog.Any("error", err))
			os.Exit(1)
		} else {
			serverAddr = name
			logger.Info("using updated serverAddr for certificate validation", slog.String("hostname", serverAddr), slog.String("original", cli.NodeHost))
		}
	}
	scraper := scraper.NewScraper(logger, serverAddr, cli.TokenPath, cli.Timeout)

	promRegistry := prometheus.NewRegistry()
	err := promRegistry.Register(scraper)
	if err != nil {
		logger.Error("failed to register storage metric")
		os.Exit(1)
	}

	promLis, err := net.Listen("tcp", cli.PromListen)
	if err != nil {
		logger.Error("failed to open prometheus listener",
			slog.String("prometheus-listen", cli.PromListen),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	promMux := http.NewServeMux()
	promMux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))
	promServer := http.Server{Handler: promMux}

	var g run.Group

	g.Add(run.SignalHandler(ctx, os.Interrupt, syscall.SIGTERM))

	g.Add(func() error {
		return promServer.Serve(promLis)
	}, func(error) {
		// Give the prom server its own timeout to cleanly shutdown
		sctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = promServer.Shutdown(sctx)
	})

	if err := g.Run(); err != nil {
		if serr, ok := err.(run.SignalError); ok {
			logger.Info("caught signal",
				slog.String("signal", serr.Signal.String()),
			)
		} else {
			logger.Error("actor failed",
				slog.Any("error", err),
			)

			os.Exit(1)
		}
	}
}
