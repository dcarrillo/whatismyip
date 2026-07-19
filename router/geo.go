package router

import (
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"

	"github.com/dcarrillo/whatismyip/models"
	"github.com/gin-gonic/gin"
)

type geoDataFormatter struct {
	title  string
	format func(*models.GeoRecord) string
}

type asnDataFormatter struct {
	title  string
	format func(*models.ASNRecord) string
}

var geoOutput = map[string]geoDataFormatter{
	"country": {
		title: "Country",
		format: func(record *models.GeoRecord) string {
			return record.Country.Names["en"]
		},
	},
	"country_code": {
		title: "Country Code",
		format: func(record *models.GeoRecord) string {
			return record.Country.ISOCode
		},
	},
	"city": {
		title: "City",
		format: func(record *models.GeoRecord) string {
			return record.City.Names["en"]
		},
	},
	"latitude": {
		title: "Latitude",
		format: func(record *models.GeoRecord) string {
			return fmt.Sprintf("%f", record.Location.Latitude)
		},
	},
	"longitude": {
		title: "Longitude",
		format: func(record *models.GeoRecord) string {
			return fmt.Sprintf("%f", record.Location.Longitude)
		},
	},
	"postal_code": {
		title: "Postal Code",
		format: func(record *models.GeoRecord) string {
			return record.Postal.Code
		},
	},
	"time_zone": {
		title: "Time Zone",
		format: func(record *models.GeoRecord) string {
			return record.Location.TimeZone
		},
	},
}

var asnOutput = map[string]asnDataFormatter{
	"number": {
		title: "ASN Number",
		format: func(record *models.ASNRecord) string {
			return fmt.Sprintf("%d", record.AutonomousSystemNumber)
		},
	},
	"organization": {
		title: "ASN Organization",
		format: func(record *models.ASNRecord) string {
			return record.AutonomousSystemOrganization
		},
	},
}

func getGeoAsString(ctx *gin.Context) {
	if geoSvc == nil {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	record := geoSvc.LookUpCity(net.ParseIP(ctx.ClientIP()))
	if record == nil {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	field := strings.ToLower(ctx.Params.ByName("field"))
	if field == "" {
		ctx.String(http.StatusOK, geoCityRecordToString(record))
		return
	}

	g, ok := geoOutput[field]
	if !ok {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	ctx.String(http.StatusOK, g.format(record))
}

func getASNAsString(ctx *gin.Context) {
	if geoSvc == nil {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	record := geoSvc.LookUpASN(net.ParseIP(ctx.ClientIP()))
	if record == nil {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}

	field := strings.ToLower(ctx.Params.ByName("field"))
	if field == "" {
		ctx.String(http.StatusOK, geoASNRecordToString(record))
		return
	}

	g, ok := asnOutput[field]
	if !ok {
		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	ctx.String(http.StatusOK, g.format(record))
}

func geoCityRecordToString(record *models.GeoRecord) string {
	var output string

	keys := make([]string, 0, len(geoOutput))
	for k := range geoOutput {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		output += fmt.Sprintf("%s: %v\n", geoOutput[k].title, geoOutput[k].format(record))
	}

	return output
}

func geoASNRecordToString(record *models.ASNRecord) string {
	var output string

	keys := make([]string, 0, len(asnOutput))
	for k := range asnOutput {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		output += fmt.Sprintf("%s: %v\n", asnOutput[k].title, asnOutput[k].format(record))
	}

	return output
}
