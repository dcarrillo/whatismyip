package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/models"
	"github.com/dcarrillo/whatismyip/router"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

var (
	tcpServer *http.Server
	tlsServer *http.Server
	engine    *gin.Engine
)

func main() {
	o, err := setting.Setup(os.Args[1:])
	if err == flag.ErrHelp || err == setting.ErrVersion {
		fmt.Print(o)
		os.Exit(0)
	} else if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
	setupEngine()
	router.SetupTemplate(engine)
	router.Setup(engine)

	if setting.App.BindAddress != "" {
		runTCPServer()
	}

	if setting.App.TLSAddress != "" {
		runTLSServer()
	}

	runHandler()
}

func runHandler() {
	signalChan := make(chan os.Signal, 3)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()
	var s os.Signal

	for {
		s = <-signalChan

		if s == syscall.SIGHUP {
			models.CloseDBs()
			models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
			router.SetupTemplate(engine)

			if setting.App.BindAddress != "" {
				if err := tcpServer.Shutdown(ctx); err != nil {
					log.Printf("TCP server forced to shutdown: %s", err)
				}
				runTCPServer()
			}
			if setting.App.TLSAddress != "" {
				if err := tlsServer.Shutdown(ctx); err != nil {
					log.Printf("TLS server forced to shutdown: %s", err)
				}
				runTLSServer()
			}
		} else {
			log.Printf("Shutting down...")
			if setting.App.BindAddress != "" {
				if err := tcpServer.Shutdown(ctx); err != nil {
					log.Printf("TCP server forced to shutdown: %s", err)
				}
			}
			if setting.App.TLSAddress != "" {
				if err := tlsServer.Shutdown(ctx); err != nil {
					log.Printf("TLS server forced to shutdown: %s", err)
				}
			}
			models.CloseDBs()
			break
		}
	}
}

func runTCPServer() {
	tcpServer = &http.Server{
		Addr:         setting.App.BindAddress,
		Handler:      engine,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting TCP server listening on %s", setting.App.BindAddress)
		if err := tcpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		log.Printf("Stopping TCP server...")
	}()
}

func runTLSServer() {
	tlsServer = &http.Server{
		Addr:         setting.App.TLSAddress,
		Handler:      engine,
		ReadTimeout:  setting.App.Server.ReadTimeout,
		WriteTimeout: setting.App.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Starting TLS server listening on %s", setting.App.TLSAddress)
		if err := tlsServer.ListenAndServeTLS(setting.App.TLSCrtPath, setting.App.TLSKeyPath); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		log.Printf("Stopping TLS server...")
	}()
}

func setupEngine() {
	gin.DisableConsoleColor()
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine = gin.New()
	engine.Use(gin.LoggerWithFormatter(httputils.GetLogFormatter))
	engine.Use(gin.Recovery())
	if setting.App.EnableSecureHeaders {
		engine.Use(addSecureHeaders())
	}
	_ = engine.SetTrustedProxies(nil)
	engine.TrustedPlatform = setting.App.TrustedHeader
}

func addSecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := secure.New(secure.Options{
			BrowserXssFilter:   true,
			ContentTypeNosniff: true,
			FrameDeny:          true,
		}).Process(c.Writer, c.Request)
		if err != nil {
			c.Abort()
			return
		}

		// Avoid header rewrite if response is a redirection.
		if status := c.Writer.Status(); status > 300 && status < 399 {
			c.Abort()
		}
	}
}
