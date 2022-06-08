package main

type Config struct {
	DeviceName                string `json:"deviceName"`
	LocalAddr                 string `json:"localAddr"`
	ServerAddr                string `json:"serverAddr"`
	DNSServerIP               string `json:"dNSServerIP"`
	CIDR                      string `json:"cIDR"`
	CIDRv6                    string `json:"cIDRv6"`
	ServerMode                bool `json:"ServerMode"`
	GlobalMode                bool `json:"GlobalMode"`
	MTU                       int `json:"mTU"`
	Timeout                   int `json:"timeout"`
	LocalGateway              string `json:"localGateway"`
	TLSCertificateFilePath    string `json:"tLSCertificateFilePath"`
	TLSCertificateKeyFilePath string `json:"tLSCertificateKeyFilePath"`
	TLSSni                    string `json:"tLSSni"`
}
