package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/resolver"
	"github.com/dcarrillo/whatismyip/server"
	"github.com/gin-contrib/secure"
	"github.com/patrickmn/go-cache"

	"github.com/dcarrillo/whatismyip/router"
	"github.com/gin-gonic/gin"
)

func main() {
	o, err := setting.Setup(os.Args[1:])
	if err == flag.ErrHelp || err == setting.ErrVersion {
		fmt.Print(o)
		os.Exit(0)
	} else if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	servers := []server.Server{}
	engine := setupEngine()

	if setting.App.Resolver.Domain != "" {
		store := cache.New(1*time.Minute, 10*time.Minute)
		dnsEngine := resolver.Setup(store)
		nameServer := server.NewDNSServer(context.Background(), dnsEngine.Handler())
		servers = append(servers, nameServer)
		engine.Use(router.GetDNSDiscoveryHandler(store, setting.App.Resolver.Domain, setting.App.Resolver.RedirectPort))
	}

	router.SetupTemplate(engine)
	router.Setup(engine)
	servers = slices.Concat(servers, setupHTTPServers(context.Background(), engine.Handler()))

	whatismyip := server.Setup(servers)
	whatismyip.Run()
}

func setupEngine() *gin.Engine {
	gin.DisableConsoleColor()
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.LoggerWithFormatter(httputils.GetLogFormatter))
	engine.Use(gin.Recovery())
	if setting.App.EnableSecureHeaders {
		engine.Use(secure.New(secure.Config{
			BrowserXssFilter:   true,
			ContentTypeNosniff: true,
			FrameDeny:          true,
		}))
	}
	_ = engine.SetTrustedProxies(nil)
	engine.TrustedPlatform = setting.App.TrustedHeader

	return engine
}

func setupHTTPServers(ctx context.Context, handler http.Handler) []server.Server {
	var servers []server.Server

	if setting.App.BindAddress != "" {
		tcpServer := server.NewTCPServer(ctx, &handler)
		servers = append(servers, tcpServer)
	}

	if setting.App.TLSAddress != "" {
		tlsServer := server.NewTLSServer(ctx, &handler)
		servers = append(servers, tlsServer)
		if setting.App.EnableHTTP3 {
			quicServer := server.NewQuicServer(ctx, tlsServer)
			servers = append(servers, quicServer)
		}
	}

	return servers
}
