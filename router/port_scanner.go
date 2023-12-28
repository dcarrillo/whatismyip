package router

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/dcarrillo/whatismyip/service"
	"github.com/gin-gonic/gin"
)

type JSONScanResponse struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Reachable bool   `json:"reachable"`
	Reason    string `json:"reason"`
}

func scanTCPPort(ctx *gin.Context) {
	port, err := strconv.Atoi(ctx.Params.ByName("port"))
	if err == nil && (port < 1 || port > 65535) {
		err = fmt.Errorf("%d is not a valid port number", port)
	}
	if err != nil {
		ctx.JSON(http.StatusBadRequest, JSONScanResponse{
			Reason: err.Error(),
		})
		return
	}

	add := net.TCPAddr{
		IP:   net.ParseIP(ctx.ClientIP()),
		Port: port,
	}

	scan := service.PortScanner{
		Address: &add,
	}

	isOpen, err := scan.IsPortOpen()
	reason := ""
	if err != nil {
		reason = err.Error()
	}

	response := JSONScanResponse{
		IP:        ctx.ClientIP(),
		Port:      port,
		Reachable: isOpen,
		Reason:    reason,
	}
	ctx.JSON(http.StatusOK, response)
}
