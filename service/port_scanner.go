package service

import (
	"net"
	"time"
)

type PortScanner struct {
	Address net.Addr
}

func (p *PortScanner) IsPortOpen() (bool, error) {
	conn, err := net.DialTimeout(p.Address.Network(), p.Address.String(), 3*time.Second)
	if err != nil {
		return false, err
	}
	if conn != nil {
		defer conn.Close()
	}
	return true, nil
}
