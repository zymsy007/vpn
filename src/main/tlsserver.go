package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
	"wat"
	"patrickmn/go-cache"
)

var _cache = cache.New(30*time.Minute, 10*time.Minute)

func GetCache() *cache.Cache {
	return _cache
}
var _key = []byte("vpn")

func SetKey(key string) {
	_key = []byte(key)
}

func XOR(src []byte) []byte {
	_klen := len(_key)
	for i := 0; i < len(src); i++ {
		src[i] ^= _key[i%_klen]
	}
	return src
}

//Start tls server
func StartServer(config Config) {
	iface := CreateTun(config)
	cert, err := tls.LoadX509KeyPair(config.TLSCertificateFilePath, config.TLSCertificateKeyFilePath)
	if err != nil {
		log.Panic(err)
	}
	tlsconfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	ln, err := tls.Listen("tcp", config.LocalAddr, tlsconfig)
	if err != nil {
		log.Panic(err)
	}
	// server -> client
	go toClient(config, iface)
	// client -> server
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go toServer(config, conn, iface)
	}
}

func toClient(config Config, iface *wat.Interface) {
	packet := make([]byte, config.MTU)
	for {
		n, err := iface.Read(packet)
		if err != nil || err == io.EOF || n == 0 {
			continue
		}
		b := packet[:n]
		if key := GetDstKey(b); key != "" {
			if v, ok := GetCache().Get(key); ok {
				v.(net.Conn).Write(b)
			}
		}
	}
}

// todo fallback to http
func toServer(config Config, tlsconn net.Conn, iface *wat.Interface) {
	defer tlsconn.Close()
	packet := make([]byte, config.MTU)
	for {
		tlsconn.SetReadDeadline(time.Now().Add(time.Duration(config.Timeout) * time.Second))
		n, err := tlsconn.Read(packet)
		if err != nil || err == io.EOF {
			break
		}
		b := packet[:n]
		if key := GetSrcKey(b); key != "" {
			GetCache().Set(key, tlsconn, 10*time.Minute)
			iface.Write(b)
		}
	}
}
