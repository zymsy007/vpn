package main

import (
	"log"
	"net"
	"runtime"
	"strconv"
	"wat"
)

func CreateTun(config Config) (iface *wat.Interface) {
	c := wat.Config{DeviceType: wat.TUN}
	if config.DeviceName != "" {
		c = wat.Config{DeviceType: wat.TUN, PlatformSpecificParams: wat.PlatformSpecificParams{Name: config.DeviceName}}
	}
	iface, err := wat.New(c)
	if err != nil {
		log.Fatalln("failed to create tun interface:", err)
	}
	log.Println("interface created:", iface.Name())
	configTun(config, iface)
	return iface
}

func configTun(config Config, iface *wat.Interface) {
	ExecCmd("/sbin/ip", "link", "set", "dev", iface.Name(), "mtu", strconv.Itoa(config.MTU))
		ExecCmd("/sbin/ip", "addr", "add", config.CIDR, "dev", iface.Name())
		ExecCmd("/sbin/ip", "-6", "addr", "add", config.CIDRv6, "dev", iface.Name())
		ExecCmd("/sbin/ip", "link", "set", "dev", iface.Name(), "up")
		if !config.ServerMode && config.GlobalMode {
			physicalIface := GetInterface()
			host, _, err := net.SplitHostPort(config.ServerAddr)
			if err != nil {
				log.Panic("error server address")
			}
			serverIP := LookupIP(host)
			if physicalIface != "" && serverIP != nil {
				ExecCmd("/sbin/ip", "route", "add", "0.0.0.0/1", "dev", iface.Name())
				ExecCmd("/sbin/ip", "-6", "route", "add", "::/1", "dev", iface.Name())
				ExecCmd("/sbin/ip", "route", "add", "128.0.0.0/1", "dev", iface.Name())
				ExecCmd("/sbin/ip", "route", "add", config.DNSServerIP+"/32", "via", config.LocalGateway, "dev", physicalIface)
				if serverIP.To4() != nil {
					ExecCmd("/sbin/ip", "route", "add", serverIP.To4().String()+"/32", "via", config.LocalGateway, "dev", physicalIface)
				} else {
					ExecCmd("/sbin/ip", "-6", "route", "add", serverIP.To16().String()+"/64", "via", config.LocalGateway, "dev", physicalIface)
				}
			}
		}
}

func Reset(config Config) {
	os := runtime.GOOS
	if os == "darwin" && !config.ServerMode && config.GlobalMode {
		ExecCmd("route", "add", "default", config.LocalGateway)
		ExecCmd("route", "change", "default", config.LocalGateway)
	}
}
