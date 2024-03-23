package service

import (
	"net"
	"time"
)

const scannerTimeOut = 3 * time.Second

type PortScanner struct {
	Address net.Addr
}

func (p *PortScanner) IsPortOpen() (bool, error) {
	conn, err := net.DialTimeout(p.Address.Network(), p.Address.String(), scannerTimeOut)
	if err != nil {
		return false, err
	}
	if conn != nil {
		defer conn.Close()
	}

	return true, nil
}
