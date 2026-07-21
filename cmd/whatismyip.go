package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/metrics"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/resolver"
	"github.com/dcarrillo/whatismyip/router"
	"github.com/dcarrillo/whatismyip/server"
	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-contrib/secure"
	"github.com/patrickmn/go-cache"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, o, err := setting.Setup(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) || errors.Is(err, setting.ErrVersion) {
			fmt.Print(o)
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var geoSvc *service.Geo
	if cfg.GeodbPath.City != "" || cfg.GeodbPath.ASN != "" {
		if geoSvc, err = service.NewGeo(context.Background(), cfg.GeodbPath.City, cfg.GeodbPath.ASN); err != nil {
			log.Fatalf("Failed to load geo databases: %s", err)
		}
	}

	servers := []server.Server{}
	engine := setupEngine(cfg)

	if cfg.Resolver.Domain != "" {
		store := cache.New(1*time.Minute, 10*time.Minute)
		var dnsEngine *resolver.Resolver
		if dnsEngine, err = resolver.Setup(store, resolver.Settings{
			Domain:          cfg.Resolver.Domain,
			ResourceRecords: cfg.Resolver.ResourceRecords,
			RedirectPort:    cfg.Resolver.RedirectPort,
			IPv4:            cfg.Resolver.Ipv4,
			IPv6:            cfg.Resolver.Ipv6,
		}); err != nil {
			log.Fatalf("Invalid resolver configuration: %s", err)
		}
		nameServer := server.NewDNSServer(context.Background(), dnsEngine.Handler())
		servers = append(servers, nameServer)
		engine.Use(router.GetDNSDiscoveryHandler(store, geoSvc, cfg.Resolver.Domain, cfg.Resolver.RedirectPort))
	}

	rt := router.NewRouter(geoSvc, cfg.TrustedHeader, cfg.TrustedPortHeader, cfg.TemplatePath, cfg.DisableTCPScan)
	router.SetupTemplate(engine, cfg.TemplatePath)
	router.Setup(engine, rt)
	servers = slices.Concat(servers, setupHTTPServers(context.Background(), engine.Handler(), cfg))

	if cfg.PrometheusAddress != "" {
		prometheusServer := server.NewPrometheusServer(context.Background(), cfg.PrometheusAddress,
			server.ServerTimeouts{
				ReadTimeout:  cfg.Server.ReadTimeout,
				WriteTimeout: cfg.Server.WriteTimeout,
			})
		servers = append(servers, prometheusServer)
	}

	whatismyip := server.Setup(servers, geoSvc)
	whatismyip.Run()
}

func setupEngine(cfg setting.Settings) *gin.Engine {
	gin.DisableConsoleColor()
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.LoggerWithFormatter(httputils.GetLogFormatter), gin.Recovery())
	if cfg.PrometheusAddress != "" {
		metrics.Enable()
		engine.Use(metrics.GinMiddleware())
	}
	if cfg.EnableSecureHeaders {
		engine.Use(secure.New(secure.Config{
			BrowserXssFilter:   true,
			ContentTypeNosniff: true,
			FrameDeny:          true,
		}))
	}
	_ = engine.SetTrustedProxies(nil)
	engine.TrustedPlatform = cfg.TrustedHeader

	return engine
}

func setupHTTPServers(ctx context.Context, handler http.Handler, cfg setting.Settings) []server.Server {
	var servers []server.Server
	timeouts := server.ServerTimeouts{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	if cfg.BindAddress != "" {
		tcpServer := server.NewTCPServer(ctx, handler, cfg.BindAddress, timeouts)
		servers = append(servers, tcpServer)
	}

	if cfg.TLSAddress != "" {
		tlsServer := server.NewTLSServer(ctx, handler, cfg.TLSAddress, cfg.TLSCrtPath, cfg.TLSKeyPath, timeouts)
		servers = append(servers, tlsServer)
		if cfg.EnableHTTP3 {
			quicServer := server.NewQuicServer(ctx, tlsServer, cfg.TLSAddress, cfg.TLSCrtPath, cfg.TLSKeyPath)
			servers = append(servers, quicServer)
		}
	}

	return servers
}
