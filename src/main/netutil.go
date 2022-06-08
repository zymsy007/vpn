package main

import (
	"log"
	"net"
	"os/exec"
	"strings"
)


func GetInterface() (name string) {
	ifaces := getAllInterfaces()
	if len(ifaces) == 0 {
		return ""
	}
	netAddrs, _ := ifaces[0].Addrs()
	for _, addr := range netAddrs {
		ip, ok := addr.(*net.IPNet)
		if ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
			name = ifaces[0].Name
			break
		}
	}
	return name
}

func getAllInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return nil
	}

	var outInterfaces []net.Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp == 1 && isPhysicalInterface(iface.Name) {
			netAddrs, _ := iface.Addrs()
			if len(netAddrs) > 0 {
				outInterfaces = append(outInterfaces, iface)
			}
		}
	}
	return outInterfaces
}

func isPhysicalInterface(addr string) bool {
	prefixArray := []string{"ens", "enp", "enx", "eno", "eth", "en0", "wlan", "wlp", "wlo", "wlx", "wifi0", "lan0"}
	for _, pref := range prefixArray {
		if strings.HasPrefix(strings.ToLower(addr), pref) {
			return true
		}
	}
	return false
}

func LookupIP(domain string) net.IP {
	ips, err := net.LookupIP(domain)
	if err != nil || len(ips) == 0 {
		log.Println(err)
		return nil
	}
	return ips[0]
}

func IsIPv4(packet []byte) bool {
	flag := packet[0] >> 4
	return flag == 4
}

func IsIPv6(packet []byte) bool {
	flag := packet[0] >> 4
	return flag == 6
}

func GetIPv4Src(packet []byte) net.IP {
	return net.IPv4(packet[12], packet[13], packet[14], packet[15])
}

func GetIPv4Dst(packet []byte) net.IP {
	return net.IPv4(packet[16], packet[17], packet[18], packet[19])
}

func GetIPv6Src(packet []byte) net.IP {
	return net.IP(packet[8:24])
}

func GetIPv6Dst(packet []byte) net.IP {
	return net.IP(packet[24:40])
}

func GetSrcKey(packet []byte) string {
	key := ""
	if IsIPv4(packet) && len(packet) >= 20 {
		key = GetIPv4Src(packet).To4().String()
	} else if IsIPv6(packet) && len(packet) >= 40 {
		key = GetIPv6Src(packet).To16().String()
	}
	return key
}

func GetDstKey(packet []byte) string {
	key := ""
	if IsIPv4(packet) && len(packet) >= 20 {
		key = GetIPv4Dst(packet).To4().String()
	} else if IsIPv6(packet) && len(packet) >= 40 {
		key = GetIPv6Dst(packet).To16().String()
	}
	return key
}

func ExecCmd(c string, args ...string) string {
	log.Printf("exec cmd: %v %v:", c, args)
	cmd := exec.Command(c, args...)
	out, err := cmd.Output()
	if err != nil {
		log.Println("failed to exec cmd:", err)
	}
	if len(out) == 0 {
		return ""
	}
	s := string(out)
	return strings.ReplaceAll(s, "\n", "")
}

func GetLocalGatewayOnLinux(ipv4 bool) string {
	if ipv4 {
		return ExecCmd("sh", "-c", "route -n | grep 'UG[ \t]' | awk 'NR==1{print $2}'")
	}
	return ExecCmd("sh", "-c", "route -6 -n | grep 'UG[ \t]' | awk 'NR==1{print $2}'")
}

func GetLocalGatewayOnMac(ipv4 bool) string {
	if ipv4 {
		return ExecCmd("sh", "-c", "route -n get default | grep 'gateway' | awk 'NR==1{print $2}'")
	}
	return ExecCmd("sh", "-c", "route -6 -n get default | grep 'gateway' | awk 'NR==1{print $2}'")
}
