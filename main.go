package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armon/go-socks5"
)

type Config struct {
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	// 读取配置文件
	configFile, err := os.Open("config.json")
	if err != nil {
		os.Exit(1)
	}
	defer configFile.Close()

	var cfg Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&cfg); err != nil {
		os.Exit(1)
	}

	// 验证必要字段
	if cfg.Port == "" || cfg.Username == "" || cfg.Password == "" {
		os.Exit(1)
	}

	// 构建认证信息
	credentials := socks5.StaticCredentials{
		cfg.Username: cfg.Password,
	}
	authenticator := socks5.UserPassAuthenticator{Credentials: credentials}

	conf := &socks5.Config{
		AuthMethods: []socks5.Authenticator{authenticator},
	}

	server, err := socks5.New(conf)
	if err != nil {
		os.Exit(1)
	}

	// 监听地址：0.0.0.0:配置的端口
	addr := "0.0.0.0:" + cfg.Port

	go func() {
		if err := server.ListenAndServe("tcp", addr); err != nil {
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	time.Sleep(3 * time.Second)
}
