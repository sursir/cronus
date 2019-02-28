package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
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
	HostID   string `json:"host_id"`
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
	Uptime   uint64 `json:"uptime"`
}

// NetworkSpeed struct
type NetworkSpeed struct {
	Download uint64 `json:"download"`
	Upload   uint64 `json:"upload"`
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
				Upload:   (uint64(usage.BytesRecv) - uint64(agoUsage.BytesRecv)),
				Download: (uint64(usage.BytesSent) - uint64(agoUsage.BytesSent)),
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

	addr := flag.String("addr", "http://127.0.0.1:8080", "Report Address")
	token := flag.String("token", "", "Authorization Token")
	flag.Parse()

	for {
		system := &SystemStatus{
			Network: GetNetworkSpeed(),
			Host:    GetHostInfo(),
		}

		json, _ := json.Marshal(system)
		req, err := http.NewRequest("POST", *addr, bytes.NewBuffer(json))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+*token)

		client := &http.Client{}
		client.Timeout = time.Second * 15
		resp, err := client.Do(req)

		if err != nil {
			log.Fatalf(err.Error())
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		log.Println(string(body))
	}
}
