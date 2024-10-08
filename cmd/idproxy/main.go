// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dexidp/dex/server"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client/config"

	greenhousesapv1alpha1 "github.com/cloudoperators/greenhouse/pkg/apis/greenhouse/v1alpha1"
	"github.com/cloudoperators/greenhouse/pkg/idproxy"
	"github.com/cloudoperators/greenhouse/pkg/idproxy/web"
)

func main() {
	var kubeconfig, kubecontext, kubenamespace string
	var issuer string
	var idTokenValidity time.Duration
	var listenAddr, metricsAddr string
	var allowedOrigins []string
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "Use kubeconfig for authentication")
	flag.StringVar(&kubecontext, "kubecontext", os.Getenv("KUBECONTEXT"), "Use context from kubeconfig")
	flag.StringVar(&kubenamespace, "kubenamespace", os.Getenv("KUBENAMESPACE"), "Use namespace")
	flag.StringVar(&issuer, "issuer", "", "Issuer URL")
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "oidc listen address")
	flag.StringVar(&metricsAddr, "metrics-addr", ":6543", "bind address for metrics")
	flag.StringSliceVar(&allowedOrigins, "allowed-origins", []string{"*"}, "list of allowed origins for CORS requests on discovery, token and keys endpoint")
	flag.DurationVar(&idTokenValidity, "id-token-validity", 1*time.Hour, "Maximum validity of issued id tokens")
	flag.Parse()

	if issuer == "" {
		log.Fatal("No --issuer given")
	}

	dexStorage, err := idproxy.NewKubernetesStorage(kubeconfig, kubecontext, kubenamespace, logger.With("component", "storage"))
	if err != nil {
		log.Fatalf("Failed to initialize kubernetes storage: %s", err)
	}

	refreshPolicy, err := server.NewRefreshTokenPolicy(logger.With("component", "refreshtokenpolicy"), true, "24h", "24h", "5s")
	if err != nil {
		log.Fatalf("Failed to setup refresh token policy: %s", err)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	config := server.Config{
		Issuer:             issuer,
		SkipApprovalScreen: true,
		Logger:             logger.With("component", "server"),
		Storage:            dexStorage,
		AllowedOrigins:     allowedOrigins,
		IDTokensValidFor:   idTokenValidity,
		RefreshTokenPolicy: refreshPolicy,
		PrometheusRegistry: registry,
		Web: server.WebConfig{
			WebFS: web.FS(),
		},
	}

	server.ConnectorsConfig["greenhouse-oidc"] = func() server.ConnectorConfig {
		k8sConfig, err := ctrl.GetConfigWithContext(kubecontext)
		if err != nil {
			log.Fatalf(`Failed to create k8s config: %s`, err)
		}

		scheme := runtime.NewScheme()
		err = greenhousesapv1alpha1.AddToScheme(scheme)
		if err != nil {
			log.Fatalf(`Failed to create scheme: %s`, err)
		}
		k8sClient, err := client.New(k8sConfig, client.Options{Scheme: scheme})
		if err != nil {
			log.Fatalf(`Failed to create k8s client: %s`, err)
		}

		oidcConfig := new(idproxy.OIDCConfig)
		oidcConfig.AddClient(k8sClient)
		oidcConfig.AddRedirectURI(issuer + "/callback")

		return oidcConfig
	}

	var g run.Group

	// Add signal handler
	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	// oidc server
	ctx, cancel := context.WithCancel(context.Background())
	serv, err := server.NewServer(ctx, config)
	if err != nil {
		log.Fatalf("OIDC server setup failed: %s", err)
	}
	s := &http.Server{Handler: serv}
	g.Add(func() error {
		ln, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
		}
		logger.Info("OIDC server listening ", "address", listenAddr)
		return s.Serve(ln)
	}, func(_ error) {
		cancel()
		timeoutCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		s.Shutdown(timeoutCtx) //nolint: errcheck
	})

	// metrics server
	ms := &http.Server{Handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry})}
	g.Add(func() error {
		ln, err := net.Listen("tcp", metricsAddr)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", metricsAddr, err)
		}
		logger.Info("Metrics server listing", "address", metricsAddr)
		return ms.Serve(ln)
	}, func(_ error) {
		ms.Close()
	})

	err = g.Run()
	var signalErr run.SignalError
	if ok := errors.As(err, &signalErr); ok {
		return
	}
	log.Fatal(err)
}
