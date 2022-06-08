package main

import (
	"io/ioutil"
	"net"
	"log"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"util"
)

func main() {
	config := Config{}
	loadCfg(&config)
	go startApp(config)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	stopApp(config)
}

func loadCfg(cfg *Config) error {
	filePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Print(err)
		return err
	}
	fullPathFile := filePath + "/vpn.json"

	bytes, err := ioutil.ReadFile(fullPathFile)
	if err != nil {
		util.LOG.Errorf("loadPayCfg ReadFile: %s", err.Error())
		return err
	}

	if err := json.Unmarshal(bytes, cfg); err != nil {
		util.LOG.Errorf("loadPayCfg Unmarshal error: %s", err.Error())
		return err
	}
	return err
}

func startApp(config Config) {
		if config.ServerMode {
			StartServer(config)
		} else {
			StartClient(config)
		}
}

func stopApp(config Config) {
	Reset(config)
	log.Printf("stopped!!!")
}

func initConfig(config *Config) {
	if !config.ServerMode && config.GlobalMode {
		host, _, err := net.SplitHostPort(config.ServerAddr)
		if err != nil {
			log.Panic("error server address")
		}
		serverIP := LookupIP(host)
		config.LocalGateway = GetLocalGatewayOnLinux(serverIP.To4() != nil)

	}
	//cipher.SetKey(config.Key)
	json, _ := json.Marshal(config)
	log.Printf("init config:%s", string(json))
}