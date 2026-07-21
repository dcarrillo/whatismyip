package router

import (
	"html/template"
	"log"

	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-gonic/gin"
)

type Router struct {
	geo               *service.Geo
	trustedHeader     string
	trustedPortHeader string
	templatePath      string
	disableScan       bool
}

func NewRouter(geo *service.Geo, trustedHeader, trustedPortHeader, templatePath string, disableScan bool) *Router {
	return &Router{
		geo:               geo,
		trustedHeader:     trustedHeader,
		trustedPortHeader: trustedPortHeader,
		templatePath:      templatePath,
		disableScan:       disableScan,
	}
}

func SetupTemplate(r *gin.Engine, templatePath string) {
	if templatePath == "" {
		r.SetHTMLTemplate(template.Must(template.New("home").Parse(home)))
	} else {
		log.Printf("Template %s has been loaded", templatePath)
		r.LoadHTMLFiles(templatePath)
	}
}

func Setup(r *gin.Engine, rt *Router) {
	r.GET("/", rt.getRoot)
	if !rt.disableScan {
		r.GET("/scan/tcp/:port", rt.scanTCPPort)
	}
	r.GET("/client-port", rt.getClientPortAsString)
	r.GET("/geo", rt.getGeoAsString)
	r.GET("/geo/:field", rt.getGeoAsString)
	r.GET("/asn", rt.getASNAsString)
	r.GET("/asn/:field", rt.getASNAsString)
	r.GET("/headers", rt.getHeadersAsSortedString)
	r.GET("/all", rt.getAllAsString)
	r.GET("/json", rt.getJSON)
	r.GET("/:header", rt.getHeaderAsString)
}
