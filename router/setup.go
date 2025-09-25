package router

import (
	"html/template"
	"log"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-gonic/gin"
)

var geoSvc *service.Geo

func SetupTemplate(r *gin.Engine) {
	if setting.App.TemplatePath == "" {
		t, _ := template.New("home").Parse(home)
		r.SetHTMLTemplate(t)
	} else {
		log.Printf("Template %s has been loaded", setting.App.TemplatePath)
		r.LoadHTMLFiles(setting.App.TemplatePath)
	}
}

func Setup(r *gin.Engine, geo *service.Geo) {
	geoSvc = geo
	r.GET("/", getRoot)
	if !setting.App.DisableTCPScan {
		r.GET("/scan/tcp/:port", scanTCPPort)
	}
	r.GET("/client-port", getClientPortAsString)
	r.GET("/geo", getGeoAsString)
	r.GET("/geo/:field", getGeoAsString)
	r.GET("/asn", getASNAsString)
	r.GET("/asn/:field", getASNAsString)
	r.GET("/headers", getHeadersAsSortedString)
	r.GET("/all", getAllAsString)
	r.GET("/json", getJSON)
	r.GET("/:header", getHeaderAsString)
}
