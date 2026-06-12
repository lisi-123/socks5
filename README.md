```bash
# 赋予执行权限
chmod +x /root/socks5/socks5-proxy

# 创建 systemd 服务文件
cat > /etc/systemd/system/socks5-proxy.service <<EOF
[Unit]
Description=SOCKS5 Proxy
After=network.target

[Service]
ExecStart=/root/socks5/socks5-proxy
WorkingDirectory=/root/socks5
Restart=always
RestartSec=5
User=root
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# 重载 systemd，启动并设置开机自启
systemctl daemon-reload
systemctl start socks5-proxy
systemctl enable socks5-proxy

# 查看状态（可选）
systemctl status socks5-proxy

```

<br>

```bash

{
      "tag": "socks5",
      "protocol": "socks",
      "settings": {
      "servers": [
        {
        "address": "127.0.0.1",
        "port": 1080,
        "users": [
          {
            "user": "myuser",
            "pass": "Myp@ssw0rd#2026"
          }
        ]
        }
      ]
    }
  }

```
