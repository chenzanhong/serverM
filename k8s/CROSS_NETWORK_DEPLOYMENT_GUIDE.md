# 跨网络环境下的serverM部署与反向SSH配置指南

## 场景说明

本文档适用于以下跨网络部署场景：

1. **网络环境**：
   - 服务器a位于内部网络A
   - 服务器b(宿主机)位于本地网络B
   - 服务器b上运行Kind创建的Kubernetes集群
   - serverM项目部署在Kubernetes集群中

2. **连接需求**：
   - 服务器a需要访问网络B中的serverM服务
   - 服务器a通过反向SSH连接到serverM（Pod）
   - 宿主机b上运行数据库服务，Kubernetes Pod需要访问这些服务

## 网络拓扑示意

```
网络A            网络B
[服务器a] <-----> [服务器b(宿主机)]
                     |
                     v
             [Kind Kubernetes集群]
                     |
       +-------------+-------------+
       |             |             |
 [serverM Pod]  [PostgreSQL]  [Redis/TDengine]
                     |             |
                     +-------------+
                      (宿主机上的服务)
```

## 前提条件

1. 网络A和网络B之间有可达的网络路由
2. 服务器a可以ping通服务器b的IP地址
3. Kind集群已正确安装在服务器b上
4. kubectl已正确配置连接到Kind集群
5. 宿主机b上的数据库服务（PostgreSQL、Redis、TDengine）已启动

## 部署配置步骤

### 1. 配置IP地址

确保以下配置文件中的IP地址已正确设置为服务器b的IP（示例：192.168.202.1）：

- `k8s/configmap.yaml`中的`public_server_ip`
- `k8s_internal/serverm-internal-config.yaml`中的`public_server_ip`
- `k8s/host-services.yaml`中的Endpoints IP

### 2. 配置宿主机防火墙

在服务器b（宿主机）上配置防火墙规则，允许以下端口的入站连接：

```bash
# 假设使用ufw防火墙
ufw allow 8080/tcp  # 用于API服务访问
ufw allow 30022/tcp  # 用于反向SSH连接

# 如果使用iptables
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
iptables -A INPUT -p tcp --dport 30022 -j ACCEPT
```

### 3. 配置服务暴露

#### 3.1 配置API服务

确保`k8s/service.yaml`正确配置：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: serverm-server
  namespace: serverm
spec:
  selector:
    app: serverm-server
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

然后配置端口转发，将服务暴露到宿主机：

```bash
# 在宿主机b上执行
kubectl port-forward -n serverm svc/serverm-server 8080:80 --address 0.0.0.0
```

#### 3.2 配置SSH服务

确保`k8s/ssh-service.yaml`正确配置：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: serverm-ssh
  namespace: serverm
spec:
  type: NodePort
  ports:
  - port: 22
    targetPort: 22
    nodePort: 30022
  selector:
    app: serverm-server
```

在Kind环境中，还需要额外的端口转发：

```bash
# 在宿主机b上执行
kubectl port-forward -n serverm svc/serverm-ssh 30022:22 --address 0.0.0.0
```

### 4. 部署数据库服务

根据您的选择，部署数据库服务：

#### 4.1 使用宿主机上的数据库服务

确保`k8s/host-services.yaml`正确配置了Endpoints：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: host-services
  namespace: serverm
spec:
  ports:
  - name: postgres
    port: 5432
    protocol: TCP
    targetPort: 5432
  - name: redis
    port: 6379
    protocol: TCP
    targetPort: 6379
  - name: tdengine
    port: 6030
    protocol: TCP
    targetPort: 6300
---
apiVersion: v1
kind: Endpoints
metadata:
  name: host-services
  namespace: serverm
subsets:
- addresses:
  - ip: 192.168.202.1  # 宿主机b的IP地址
  ports:
  - name: postgres
    port: 5432
    protocol: TCP
  - name: redis
    port: 6379
    protocol: TCP
  - name: tdengine
    port: 6030
    protocol: TCP
```

#### 4.2 使用Kubernetes集群内部的数据库服务

如果选择使用内部数据库，请部署：

```bash
kubectl apply -f k8s_internal/postgresql.yaml
kubectl apply -f k8s_internal/redis.yaml
kubectl apply -f k8s_internal/tdengine.yaml
```

### 5. 部署应用

根据数据库选择，应用相应的配置：

#### 使用宿主机数据库

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/host-services.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ssh-service.yaml
```

#### 使用内部数据库

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s_internal/serverm-internal-config.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ssh-service.yaml
```

## 服务器a上的配置与连接

### 1. 验证网络连接

在服务器a上验证是否可以访问服务器b的必要端口：

```bash
# 测试API服务连接
curl http://192.168.202.1:8080

# 测试SSH服务连接
nc -zv 192.168.202.1 30022
```

### 2. 访问前端界面

在服务器b上运行前端项目
在服务器a上通过浏览器访问serverM前端：

```
http://192.168.202.1:5173
```

### 3. 获取并安装代理

1. 登录系统并获取代理安装脚本
2. 在服务器a上执行脚本：

```bash
chmod +x install_script.sh
./install_script.sh
```

脚本将：
- 下载agent程序
- 安装为系统服务
- 建立反向SSH隧道到`192.168.202.1:30022`

## 反向SSH连接工作原理

反向SSH隧道工作流程：

1. 服务器a上的代理程序通过SSH连接到serverM服务（使用配置的`public_server_ip`）
2. 在serverM Pod上打开一个随机端口（例如10001）
3. 代理程序将该端口转发到服务器a的SSH服务（localhost:22）
4. serverM系统可以通过这个隧道访问服务器a的SSH服务

## 常见问题与排查

### 1. 服务器a无法访问服务器b

- 检查网络路由是否正确
- 确认防火墙规则允许相关端口
- 验证服务器b的IP地址配置是否正确

### 2. 反向SSH连接失败

- 确认服务器a上的SSH服务是否正常运行
- 验证`public_server_ip`配置是否正确
- 检查服务器b的30022端口是否可以访问
- 查看代理日志获取详细错误信息

### 3. 数据库连接问题

- 如果使用宿主机数据库：检查数据库服务是否允许来自Pod IP的连接
- 如果使用内部数据库：验证数据库Pod是否正常运行，服务是否可解析

## 高级配置

### 持久化端口转发

为了确保端口转发在服务器b重启后仍然有效，可以创建systemd服务：

```bash
cat <<EOF | sudo tee /etc/systemd/system/serverm-api-forward.service
[Unit]
Description=ServerM API Port Forwarding
After=network.target

[Service]
User=your-user
WorkingDirectory=/path/to/project
ExecStart=/usr/bin/kubectl port-forward -n serverm svc/serverm-server 8080:80 --address 0.0.0.0
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

cat <<EOF | sudo tee /etc/systemd/system/serverm-ssh-forward.service
[Unit]
Description=ServerM SSH Port Forwarding
After=network.target

[Service]
User=your-user
WorkingDirectory=/path/to/project
ExecStart=/usr/bin/kubectl port-forward -n serverm svc/serverm-ssh 30022:22 --address 0.0.0.0
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now serverm-api-forward.service
sudo systemctl enable --now serverm-ssh-forward.service
```

### 网络安全增强

对于跨网络环境，建议加强安全措施：

1. 使用SSH密钥认证替代密码认证
2. 配置VPN连接两个网络，替代直接暴露端口
3. 限制SSH连接仅接受特定IP范围
4. 考虑使用防火墙白名单限制访问

## 性能优化

在跨网络环境中，考虑以下优化措施：

1. 增加SSH连接的心跳间隔
2. 优化数据库连接池配置
3. 对监控数据进行压缩传输
4. 根据网络状况调整数据采集频率