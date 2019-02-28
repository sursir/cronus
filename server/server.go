package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     uint64 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// RequestData 系统状态
type RequestData struct {
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

// Config 启动配置
type Config struct {
	ListenAddr string         `json:"listen_addr"`
	Token      string         `json:"token"`
	Database   DatabaseConfig `json:"database"`
}

var config Config
var db *sql.DB

// Middleware struct
type Middleware func(http.HandlerFunc) http.HandlerFunc

// RemoteIP 获取服务器 IP
func RemoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("XRealIP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("XForwardedFor"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

// Logging logs all request
func Logging() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestStartedAt := time.Now()
			defer func() {
				log.Println(r.URL.Path, RemoteIP(r), time.Since(requestStartedAt))
			}()

			f(w, r)
		}
	}
}

// Authenticate verify request
func Authenticate(token string) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestToken := strings.TrimSpace(strings.TrimLeft(r.Header.Get("Authorization"), "Bearer"))
			if requestToken != token {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			f(w, r)
		}
	}
}

// Chain 执行 Middleware 和 HandleFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}

	return f
}

// ConnectDatabase 连接数据库
func ConnectDatabase() (err error) {

	dbConfig := config.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbConfig.Username, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	if db, err = sql.Open("mysql", dsn); err != nil {
		return
	}

	err = db.Ping()
	return
}

// ReceiveSystemUsage 处理上报数据
func ReceiveSystemUsage(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	requestData := RequestData{}
	json.Unmarshal(b, &requestData)

	stmt, err := db.Prepare("INSERT INTO usages(host_id, os, hostname, version, uptime, download, upload) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("%v", err.Error())
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		requestData.Host.HostID,
		requestData.Host.OS,
		requestData.Host.Hostname,
		requestData.Host.Version,
		requestData.Host.Uptime,
		requestData.Network.Download,
		requestData.Network.Upload)

	if err != nil {
		log.Fatalf("DB Error %v", err.Error())
	}

	fmt.Fprint(w, http.StatusText(http.StatusCreated), http.StatusCreated)
}

func main() {
	// load configurations
	configPath := flag.String("c", "", "configurations path")
	flag.Parse()

	content, err := os.Open(*configPath)
	defer content.Close()
	if err != nil {
		panic(err)
	}

	b, _ := ioutil.ReadAll(content)
	json.Unmarshal(b, &config)

	// Connect DB

	if err = ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect database %v", err.Error())
	}

	// routes
	http.HandleFunc("/", Chain(ReceiveSystemUsage, Authenticate(config.Token), Logging()))

	// start server
	log.Printf("server started %v", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}
