package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func GetWslIp() (string, string) {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Name != "eth0" {
			continue
		}
		addrs, errAddrs := iface.Addrs()
		if errAddrs != nil {
			continue
		}
		for _, addr := range addrs {
			ip, ipnet, _ := net.ParseCIDR(addr.String())
			if ip.String() != ip.To4().String() {
				continue
			}
			localIp := ip.String()
			ipInt := binary.BigEndian.Uint32(ipnet.IP) + 1
			winIp := fmt.Sprintf(
				"%d.%d.%d.%d",
				(ipInt>>24)&0xff, (ipInt>>16)&0xff, (ipInt>>8)&0xff, (ipInt)&0xff,
			)
			return localIp, winIp
		}
	}
	return "", ""
}

type Config struct {
	ChromedriverBin string `json:"chromedriver_bin"`
	WindowsHostIp   string `json:"windows_host_ip"`
	WindowsHostPort string `json:"windows_host_port"`
}

func GetConfig() Config {
	var cfg Config
	binPath, err := os.Executable()
	if err != nil {
		return cfg
	}
	jsonPath := fmt.Sprintf("%s/chromedriver_wsl_config.json", filepath.Dir(binPath))
	byteArray, _ := os.ReadFile(jsonPath)
	if err := json.Unmarshal(byteArray, &cfg); err != nil {
		return cfg
	}
	return cfg
}

func main() {
	// setup configuration
	cfg := GetConfig()
	if cfg.ChromedriverBin == "" {
		fmt.Println("Please setup chromedriver_bin to chromedriver_wsl_config.json.")
		os.Exit(1)
	}
	localHostIp, winHostIp := GetWslIp()
	if cfg.WindowsHostIp == "" {
		cfg.WindowsHostIp = winHostIp
	}
	if cfg.WindowsHostPort == "" {
		cfg.WindowsHostPort = "9515"
	}

	// start chromedriver
	var port = "9515"
	var chromedriverArgs = os.Args[1:]
	for i, v := range chromedriverArgs {
		if strings.HasPrefix(v, "--port=") {
			port = v[7:]
			chromedriverArgs[i] = "--port=" + cfg.WindowsHostPort
		}
	}
	chromedriverArgs = append(chromedriverArgs, "--allowed-ips", localHostIp)
	cmd := exec.Command(cfg.ChromedriverBin, chromedriverArgs...)
	if err := cmd.Start(); err != nil {
		fmt.Println("chromedriver failed to start.")
		fmt.Println(err)
		os.Exit(1)
	}

	// start proxy server
	quit := make(chan os.Signal)
	director := func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = fmt.Sprintf("%s:%s", cfg.WindowsHostIp, cfg.WindowsHostPort)
	}
	modifier := func(res *http.Response) error {
		if res.Request.RequestURI == "/shutdown" {
			close(quit)
		}
		return nil
	}
	rp := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifier,
	}
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: rp,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
		}
	}()

	// wait for shutdown
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-quit
	if err := cmd.Process.Kill(); err != nil {
		fmt.Println(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
}
