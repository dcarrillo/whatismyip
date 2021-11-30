package router

import (
	"html/template"
	"log"

	"github.com/dcarrillo/whatismyip/internal/setting"
	"github.com/gin-gonic/gin"
)

// SetupTemplate reads and parses a template from file
func SetupTemplate(r *gin.Engine) {
	if setting.App.TemplatePath == "" {
		t, _ := template.New("home").Parse(home)
		r.SetHTMLTemplate(t)
	} else {
		log.Printf("Template %s has been loaded", setting.App.TemplatePath)
		r.LoadHTMLFiles(setting.App.TemplatePath)
	}
}

// Setup defines the endpoints
func Setup(r *gin.Engine) {
	r.GET("/", getRoot)
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
