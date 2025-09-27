# serverM Kubernetes部署方案

## 整体架构

在Kubernetes中部署serverM项目需要考虑以下组件：

1. **后端服务（server）** - 使用Deployment部署
2. **配置管理** - 使用ConfigMap和Secret
3. **数据存储** - 连接到宿主机上的PostgreSQL、Redis和TDengine服务
4. **网络访问** - 使用Service和Ingress
5. **SSH隧道** - 通过NodePort暴露SSH服务，支持反向SSH连接

## 步骤1：Docker镜像已创建

Dockerfile已在项目根目录创建，包含了正确的构建和运行步骤。

## 步骤2：修改配置加载代码

为了更好地适应Kubernetes环境，我们需要修改配置加载代码，让它能够从环境变量中读取配置：

```go
// 在config/config.go中添加环境变量读取支持
func LoadConfig() (*Config, error) {
    // 首先尝试从环境变量获取配置路径
    configPath := os.Getenv("CONFIG_PATH")
    if configPath == "" {
        configPath = GetDBConfigPath()
    }
    
    // 读取配置文件
    content, err := ioutil.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(content, &config); err != nil {
        return nil, err
    }
    
    // 从环境变量覆盖敏感配置
    if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
        config.DB.Password = dbPassword
    }
    if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
        config.Redis.Password = redisPassword
    }
    // 其他环境变量覆盖...
    
    return &config, nil
}
```

## 访问宿主机服务的方法

当数据库（PostgreSQL、TDengine）和Redis服务运行在宿主机上时，有以下几种访问方式：

### 方法一：使用host.minikube.internal（Minikube环境）

```yaml
db:
  host: "host.minikube.internal"
  port: 5432
```

### 方法二：使用节点IP

```yaml
db:
  host: "192.168.x.x"  # 替换为实际节点IP
  port: 5432
```

### 方法三：使用ClusterIP Service和Endpoints（推荐用于生产环境）

1. 创建不选择Pod的Service和指向宿主机IP的Endpoints
2. 在配置中使用Service名称访问

```yaml
db:
  host: "host-services"
  port: 5432
```

#### host-services.yaml 示例

```yaml
apiVersion: v1
kind: Service
metadata:
  name: host-services
  namespace: serverm
spec:
  ports:
  - port: 5432
    targetPort: 5432
    name: postgres
  - port: 6379
    targetPort: 6379
    name: redis
  - port: 6030
    targetPort: 6030
    name: tdengine
---
apiVersion: v1
kind: Endpoints
metadata:
  name: host-services
  namespace: serverm
subsets:
- addresses:
  - ip: 192.168.x.x  # 替换为实际宿主机IP
  ports:
  - port: 5432
    name: postgres
  - port: 6379
    name: redis
  - port: 6030
    name: tdengine
```

## 步骤3：创建Kubernetes配置文件

### 3.1 创建命名空间

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: serverm
  labels:
    app: serverm
```

### 3.2 创建ConfigMap（存储非敏感配置）

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: serverm-config
  namespace: serverm
data:
  config.yaml: |
    db:
      host: postgres
      port: "5432"
      name: serverm
      user: serverm_user
    redis:
      host: redis
      port: "6379"
      db: "0"
    smtp_server:
      SMTPServer_host: smtp.163.com
      SMTPServer_port: "465"
    script:
      github_repo_url: https://gitee.com/chenzanhong/agent.git
      start_port: "10000"
      end_port: "45000"
      public_server_ip: "your-public-server-ip"
    tdengine:
      host: localhost
      port: "6030"
      name: servermonitor
      user: root
      password: taosdata
```

### 3.3 创建Secret（存储敏感配置）

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: serverm-secret
  namespace: serverm
type: Opaque
data:
  db_password: BASE64_ENCODED_PASSWORD  # 需要替换为实际的base64编码密码
  redis_password: BASE64_ENCODED_PASSWORD  # 如果Redis有密码
  email_name: BASE64_ENCODED_EMAIL  # 163邮箱地址base64编码
  email_password: BASE64_ENCODED_EMAIL_PASSWORD  # 163邮箱应用密码base64编码
  ssh_tunnel_username: BASE64_ENCODED_USERNAME  # SSH用户名base64编码
  ssh_tunnel_password: BASE64_ENCODED_PASSWORD  # SSH密码base64编码
```

### 3.4 创建PostgreSQL和Redis（使用Helm或直接配置）

这里我们假设使用外部的PostgreSQL和Redis服务。如果需要在K8s中部署，可以使用Helm或直接创建StatefulSet。

### 3.5 创建Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: serverm-server
  namespace: serverm
spec:
  replicas: 2
  selector:
    matchLabels:
      app: serverm-server
  template:
    metadata:
      labels:
        app: serverm-server
    spec:
      containers:
      - name: serverm-server
        image: your-registry/serverm:latest  # 替换为你的镜像地址
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: "100m"
            memory: "256Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        volumeMounts:
        - name: config-volume
          mountPath: /app/config/configs/config.yaml
          subPath: config.yaml
        - name: ssh-keys
          mountPath: /root/.ssh
          readOnly: true
        env:
        - name: CONFIG_PATH
          value: /app/config/configs/config.yaml
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: serverm-secret
              key: db_password
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: serverm-secret
              key: redis_password
        - name: EMAIL_NAME
          valueFrom:
            secretKeyRef:
              name: serverm-secret
              key: email_name
        - name: EMAIL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: serverm-secret
              key: email_password
        - name: SSH_TUNNEL_USERNAME
          valueFrom:
            secretKeyRef:
              name: serverm-secret
              key: ssh_tunnel_username
        - name: SSH_TUNNEL_PASSWORD
          valueFrom:
            secretKeyRef:
              name: serverm-secret
              key: ssh_tunnel_password
      volumes:
      - name: config-volume
        configMap:
          name: serverm-config
      - name: ssh-keys
        secret:
          secretName: serverm-ssh-keys
          defaultMode: 0600
```

### 3.6 创建Service

```yaml
# k8s/service.yaml
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

### 3.7 创建Ingress（可选，如果需要外部访问）

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: serverm-ingress
  namespace: serverm
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  rules:
  - host: serverm.example.com  # 替换为你的域名
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: serverm-server
            port:
              number: 80
```

## 步骤4：部署SSH隧道服务

由于项目使用反向SSH隧道，我们需要在K8s中特别配置SSH服务：

```yaml
# k8s/ssh-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: serverm-ssh
  namespace: serverm
spec:
  selector:
    app: serverm-server
  ports:
  - port: 22
    targetPort: 22
    nodePort: 30022  # 使用NodePort以便外部访问
  type: NodePort
```

### 反向SSH服务配置

#### 公网IP配置

对于反向SSH功能，`public_server_ip`应该设置为：
- 如果使用NodePort暴露SSH服务：运行Pod的Kubernetes节点的公网IP
- 如果使用LoadBalancer或Ingress：LoadBalancer分配的IP或Ingress的IP

#### 反向SSH连接示例

在被监控服务器上，使用以下命令建立反向SSH隧道：

```bash
ssh -fN -R 10001:localhost:22 ssh_tunnel_username@[公网IP] -p 30022
```

其中：
- `10001` 是远程主机上分配的端口
- `localhost:22` 是本地SSH服务
- `ssh_tunnel_username` 是SSH隧道用户名
- `[公网IP]` 是Kubernetes节点的公网IP
- `30022` 是NodePort暴露的SSH端口

## 步骤5：部署步骤

1. **准备环境**
   - 确保Kubernetes集群正常运行
   - 确保宿主机上的PostgreSQL、Redis和TDengine服务正常运行
   - 确保防火墙允许Kubernetes节点访问这些服务

2. **修改配置**
   - 更新`configmap.yaml`中的`public_server_ip`为实际公网IP
   - 更新`host-services.yaml`中的宿主机IP

3. **构建并推送Docker镜像**
   ```bash
   docker build -t your-registry/serverm:latest .
   docker push your-registry/serverm:latest
   ```

4. **准备敏感信息**
   ```bash
   # 生成base64编码的敏感信息
   echo -n "your-db-password" | base64
   echo -n "your-email@163.com" | base64
   # 替换secret.yaml中的对应值
   ```

5. **应用Kubernetes配置**
   ```bash
   kubectl apply -f k8s/namespace.yaml
   kubectl apply -f k8s/configmap.yaml
   kubectl apply -f k8s/secret.yaml
   kubectl apply -f k8s/host-services.yaml  # 如果使用方法三访问宿主机服务
   kubectl apply -f k8s/deployment.yaml
   kubectl apply -f k8s/service.yaml
   kubectl apply -f k8s/ingress.yaml  # 如果需要
   kubectl apply -f k8s/ssh-service.yaml  # 启用SSH隧道服务
   ```

6. **验证部署**
   ```bash
   # 检查Pod状态
   kubectl get pods -n serverm
   
   # 检查服务状态
   kubectl get services -n serverm
   
   # 查看日志
   kubectl logs -n serverm -l app=serverm-server
   ```

## 步骤6：注意事项和优化

1. **配置管理优化**
   - 考虑使用Vault或云服务商的密钥管理服务管理敏感信息
   - 可以使用Helm Chart统一管理所有配置

2. **数据库考虑**
   - 对于生产环境，建议使用托管的PostgreSQL、Redis和TDengine服务
   - 如果必须在K8s中部署，使用StatefulSet确保数据持久性

3. **SSH隧道特殊处理**
   - 反向SSH需要宿主机网络权限，可以考虑使用hostNetwork或特权容器
   - 端口范围需要在Service中正确配置
   - 确保公网IP可从外部访问，且防火墙允许相应端口的入站流量

4. **端口冲突**
   - 确保宿主机上的服务端口没有被Kubernetes占用
   - 确保防火墙允许Kubernetes节点访问相关服务端口

5. **资源限制**
   - 根据实际负载调整CPU和内存限制
   - 配置水平自动扩展(HPA)应对流量变化

6. **监控和日志**
   - 集成Prometheus和Grafana监控应用状态
   - 配置ELK或类似系统收集和分析日志

7. **数据备份**
   - 定期备份PostgreSQL数据库
   - 考虑使用Velero等工具备份K8s集群状态

8. **环境变量列表**
   - `CONFIG_PATH` - 配置文件路径（默认：/app/config/configs/config.yaml）
   - `DB_PASSWORD` - 数据库密码
   - `REDIS_PASSWORD` - Redis密码
   - `EMAIL_NAME` - 邮箱地址
   - `EMAIL_PASSWORD` - 邮箱应用密码
   - `SSH_TUNNEL_USERNAME` - SSH隧道用户名
   - `SSH_TUNNEL_PASSWORD` - SSH隧道密码