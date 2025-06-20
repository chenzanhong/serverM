# 配置反向ssh服务
实现内网穿透，解决网络隔离问题，以确保文件传输功能的正常使用

## 公网服务器配置
1. 确保SSH服务已启用并允许远程登录
```bash
sudo systemctl status sshd
# 如果未启用，可以执行以下命令启用SSH服务：
sudo systemctl enable ssh
sudo systemctl start ssh
```

2. 创建一个专用用户用于反向SSH连接（可选，推荐）
```bash
sudo adduser reversessh
# 为新用户设置密码
sudo passwd reversessh
# 为新用户添加sudo权限（可选）
sudo usermod -aG sudo reversessh
```

3. 修改SSH配置（可选）：限制该用户的登录行为
编辑/etc/ssh/sshd_config 文件，添加如下内容
```bash
Match User reversessh
    PasswordAuthentication yes
    AllowTcpForwarding yes
    PermitTunnel yes
    X11Forwarding no
    PermitTTY no
    ForceCommand echo 'Only reverse SSH tunnel allowed'
```
保存后重启SSH服务：
```bash
sudo systemctl restart sshd
```

# 内网服务器配置（被监控服务器）———— 配置反向ssh服务的脚本，查看serverM/server/handle/agent/getscript/getscript.go
1. 安装 autossh
Debian/Ubuntu：
```bash
sudo apt update && sudo apt install autossh -y
```

CentOS/RHEL：
```bash
sudo yum install autossh -y
```

2. 创建SSH密钥对（非必须，但推荐）
```bash
ssh-keygen -t rsa -b 4096
```
一路回车即可。然后将公钥复制到公网服务器的 ~/.ssh/authorized_keys 中：
```bash
cat ~/.ssh/id_rsa.pub
```
把输出粘贴到公网服务器的：
```bash
~reversessh/.ssh/authorized_keys
```
这样就可以免密登录了。

3. 启动反向SSH隧道（测试用）
手动测试一下是否能成功建立隧道
```bash
autossh -M 0 -f -N -o "StrictHostKeyChecking=no" -R 2222:localhost:22 reversessh@113.44.170.52
```
说明：
-R 2222:localhost:22: 将公网服务器的2222端口转发到本机22端口（SSH）
-f: 后台运行
-N: 不执行命令，只做端口转发
-M 0: 关闭监控端口（新版autossh默认不依赖） 

4. 设置开机自启（systemd）
创建一个 systemd 服务文件：
```bash
sudo vim /etc/systemd/system/reversetunnel.service
```

写入以下内容（根据实际情况修改）：
```ini
[Unit]
Description=Reverse SSH Tunnel
After=network.target

[Service]
User=<your_local_user>  # 替换为实际用户名
ExecStart=/usr/bin/autossh -M 0 -N -o "StrictHostKeyChecking=no" -R 2222:localhost:22 reversessh@113.44.170.52
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

保存后启用服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable reversetunnel
sudo systemctl start reversetunnel
```