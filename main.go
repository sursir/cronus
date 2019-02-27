package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/net"
)

// NetworkUsage struct
type NetworkUsage struct {
	Download float64
	Upload   float64
}

// GetNetworkInterface 获取网卡名字
func GetNetworkInterface() string {
	var networkInterface string
	switch runtime.GOOS {
	case "darwin":
		networkInterface = "en0"
	case "linux", "FreeBSD":
		networkInterface = "eth0"
	default:
		fmt.Println("Unsupport system yet.")
		os.Exit(1)
	}

	return networkInterface
}

// NetworkUsageMonitor 监控网络流量
func NetworkUsageMonitor() NetworkUsage {
	networkInterface := GetNetworkInterface()
	agoIOCounters, _ := net.IOCounters(true)
	time.Sleep(time.Second)
	nowIOCounters, _ := net.IOCounters(true)

	networkUsage := NetworkUsage{}
	for idx, usage := range nowIOCounters {
		if usage.Name == networkInterface {
			agoUsage := agoIOCounters[idx]
			networkUsage = NetworkUsage{
				Upload:   (float64(usage.BytesRecv) - float64(agoUsage.BytesRecv)) / 1024,
				Download: (float64(usage.BytesSent) - float64(agoUsage.BytesSent)) / 1024,
			}
			break
		}
	}

	return networkUsage
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("The cronus agent could not be started due to 1 parameter was expected but no parameter was given.")
		return
	}

	for {
		// addr := os.Args[1]
		networkUsage := NetworkUsageMonitor()

		fmt.Println(networkUsage)
	}
}
