package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
)

// SystemStatus 系统状态
type SystemStatus struct {
	Host    HostInfo
	Network NetworkSpeed
}

// HostInfo 系统信息 Struct
type HostInfo struct {
	OS       string `json:"os"`
	HostID   string `json"host_id"`
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
	Uptime   uint64 `json:"uptime"`
}

// NetworkSpeed struct
type NetworkSpeed struct {
	Download float64 `json:"download"`
	Upload   float64 `json:"upload"`
}

// GetNetworkInterface 获取网卡名字
func GetNetworkInterface() string {
	var networkInterface string
	switch runtime.GOOS {
	case "darwin":
		networkInterface = "en0"
	case "linux", "freebsd":
		networkInterface = "eth0"
	default:
		fmt.Println("Unsupport system yet.")
		os.Exit(1)
	}

	return networkInterface
}

// GetNetworkSpeed 监控网络流量
func GetNetworkSpeed() NetworkSpeed {
	networkInterface := GetNetworkInterface()
	agoIOCounters, _ := net.IOCounters(true)
	time.Sleep(time.Second)
	nowIOCounters, _ := net.IOCounters(true)

	networkSpeed := NetworkSpeed{}
	for idx, usage := range nowIOCounters {
		if usage.Name == networkInterface {
			agoUsage := agoIOCounters[idx]
			networkSpeed = NetworkSpeed{
				Upload:   (float64(usage.BytesRecv) - float64(agoUsage.BytesRecv)) / 1024,
				Download: (float64(usage.BytesSent) - float64(agoUsage.BytesSent)) / 1024,
			}
			break
		}
	}

	return networkSpeed
}

// GetHostInfo 获取系统信息
func GetHostInfo() HostInfo {
	info, _ := host.Info()
	hostInfo := HostInfo{
		OS:       info.OS,
		HostID:   info.HostID,
		Hostname: info.Hostname,
		Uptime:   info.Uptime,
		Version:  info.PlatformVersion,
	}

	return hostInfo
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("The cronus agent could not be started due to 1 parameter was expected but no parameter was given.")
		return
	}

	for {
		system := &SystemStatus{
			Network: GetNetworkSpeed(),
			Host:    GetHostInfo(),
		}

		addr := strings.TrimLeft(os.Args[1], "-addr=")
		json, _ := json.Marshal(system)
		req, err := http.NewRequest("POST", addr, bytes.NewBuffer(json))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		client.Timeout = time.Second * 15
		resp, err := client.Do(req)

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		log.Println(string(body))
	}
}
