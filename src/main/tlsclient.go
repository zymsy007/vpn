package main

import (
	"crypto/tls"
	"io"
	"net"
	"time"
	"wat"
)

// Start tls client
func StartClient(config Config) {
	iface := CreateTun(config)
	go tunToTLS(config, iface)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
	}
	if config.TLSSni != "" {
		tlsconfig.ServerName = config.TLSSni
	}
	for {
		conn, err := tls.Dial("tcp", config.ServerAddr, tlsconfig)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		GetCache().Set("tlsconn", conn, 24*time.Hour)
		tlsToTun(config, conn, iface)
		GetCache().Delete("tlsconn")
	}
}

func tunToTLS(config Config, iface *wat.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || n == 0 {
			continue
		}
		if v, ok := GetCache().Get("tlsconn"); ok {
			b := packet[:n]
			tlsconn := v.(net.Conn)
			tlsconn.SetWriteDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
			_, err = tlsconn.Write(b)
			if err != nil {
				continue
			}
		}
	}
}

func tlsToTun(config Config, tlsconn net.Conn, iface *wat.Interface) {
	defer tlsconn.Close()
	packet := make([]byte, config.MTU)
	for {
		tlsconn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		n, err := tlsconn.Read(packet)
		if err != nil || err == io.EOF {
			break
		}
		b := packet[:n]
		_, err = iface.Write(b)
		if err != nil {
			break
		}
	}
}
