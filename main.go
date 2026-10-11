package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armon/go-socks5"
	"golang.org/x/net/proxy"
)

type CascadeConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Port     string        `json:"port"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Cascade  CascadeConfig `json:"cascade"`
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("读取 config.json 失败: %v", err)
	}
	defer configFile.Close()

	var cfg Config
	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		log.Fatalf("解析 config.json 失败: %v", err)
	}

	if cfg.Port == "" || cfg.Username == "" || cfg.Password == "" {
		log.Fatal("port、username、password 不能为空")
	}

	// 默认直接连接目标地址，不使用上游代理。
	var dial func(context.Context, string, string) (net.Conn, error)

	if cfg.Cascade.Enabled {
		if cfg.Cascade.Host == "" || cfg.Cascade.Port == "" {
			log.Fatal("已启用级联，但上游 host 或 port 为空")
		}

		upstreamAddr := net.JoinHostPort(
			cfg.Cascade.Host,
			cfg.Cascade.Port,
		)

		var auth *proxy.Auth
		if cfg.Cascade.Username != "" || cfg.Cascade.Password != "" {
			if cfg.Cascade.Username == "" || cfg.Cascade.Password == "" {
				log.Fatal("上游认证的 username 和 password 必须同时填写")
			}
			auth = &proxy.Auth{
				User:     cfg.Cascade.Username,
				Password: cfg.Cascade.Password,
			}
		}

		upstream, err := proxy.SOCKS5(
			"tcp",
			upstreamAddr,
			auth,
			proxy.Direct,
		)
		if err != nil {
			log.Fatalf("创建上游 SOCKS5 拨号器失败: %v", err)
		}

		dial = func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 上游拨号器负责通过 SOCKS5 连接最终目标。
			// SOCKS5 拨号器的 Dial 方法本身不支持 context。
			return upstream.Dial(network, addr)
		}

		log.Printf("SOCKS5 级联已启用，上游: %s", upstreamAddr)
	} else {
		dialer := &net.Dialer{}
		dial = dialer.DialContext
		log.Println("SOCKS5 级联已关闭，使用直连出口")
	}

	credentials := socks5.StaticCredentials{
		cfg.Username: cfg.Password,
	}
	authenticator := socks5.UserPassAuthenticator{
		Credentials: credentials,
	}

	conf := &socks5.Config{
		AuthMethods: []socks5.Authenticator{authenticator},
		Dial:        dial,
	}

	server, err := socks5.New(conf)
	if err != nil {
		log.Fatalf("创建 SOCKS5 服务失败: %v", err)
	}

	addr := net.JoinHostPort("0.0.0.0", cfg.Port)

	go func() {
		log.Printf("SOCKS5 服务正在监听 %s", addr)
		if err := server.ListenAndServe("tcp", addr); err != nil {
			log.Fatalf("SOCKS5 服务启动失败: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("正在退出 SOCKS5 服务...")
	time.Sleep(3 * time.Second)
}
