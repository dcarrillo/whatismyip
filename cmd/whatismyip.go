package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/dcarrillo/whatismyip/internal/httputils"
	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/server"
	"github.com/gin-contrib/secure"

	"github.com/dcarrillo/whatismyip/models"
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

	models.Setup(setting.App.GeodbPath.City, setting.App.GeodbPath.ASN)
	engine := setupEngine()
	router.SetupTemplate(engine)
	router.Setup(engine)

	whatismyip := server.Setup(context.Background(), engine.Handler())
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
