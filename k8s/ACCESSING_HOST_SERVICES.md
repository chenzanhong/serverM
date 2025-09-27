# Kubernetes环境中的特殊配置指南

## 1. 访问宿主机上的服务

当使用Kubernetes部署serverM项目，而数据库（PostgreSQL、TDengine）和Redis服务运行在宿主机上时，有以下几种访问方式：

### 方法一：使用host.minikube.internal（推荐用于Minikube）

如果使用Minikube，可以使用`host.minikube.internal`作为主机名，它会自动解析到宿主机IP：

```yaml
# 示例配置
db:
  host: "host.minikube.internal"
  port: 5432
```

### 方法二：直接使用节点IP

可以使用运行Pod的节点IP地址：

```yaml
db:
  host: "192.168.x.x"  # 替换为实际节点IP
  port: 5432
```

### 方法三：使用ClusterIP Service和Endpoints（推荐用于生产环境）

1. 创建一个不选择任何Pod的Service：

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
    targetPort: 6030
```

2. 创建指向宿主机IP的Endpoints：

```yaml
apiVersion: v1
kind: Endpoints
metadata:
  name: host-services
  namespace: serverm
subsets:
- addresses:
  - ip: 192.168.x.x  # 替换为实际宿主机IP
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

3. 然后在配置中使用该Service名称：

```yaml
db:
  host: "host-services"
  port: 5432
```

## 2. 反向SSH服务的公网IP配置

对于反向SSH功能，需要明确以下几点：

### 确定公网IP

1. **使用NodePort暴露SSH服务**：
   - 在`ssh-service.yaml`中，已配置使用NodePort 30022
   - 公网IP应该是运行Pod的Kubernetes节点的公网IP
   - 在`configmap.yaml`的`public_server_ip`中设置该节点的公网IP

2. **使用LoadBalancer或Ingress**：
   - 如果环境支持，可以配置LoadBalancer或Ingress
   - 此时公网IP是LoadBalancer分配的IP或Ingress的IP

### 配置反向SSH连接

1. 确保`ssh-service.yaml`正确配置了NodePort：

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

2. 在被监控服务器上，使用以下命令建立反向SSH隧道：

```bash
ssh -fN -R 10001:localhost:22 ssh_tunnel_username@[公网IP] -p 30022
```

## 3. 部署注意事项

1. **端口冲突**：确保宿主机上的数据库和Redis端口没有被Kubernetes占用
2. **防火墙**：确保宿主机防火墙允许Kubernetes节点访问相关服务端口
3. **性能考虑**：对于生产环境，建议将数据库和Redis也部署到Kubernetes集群中
4. **高可用**：考虑使用StatefulSet部署数据库服务以提供更好的高可用性

## 4. 环境变量优先级

serverM项目支持通过环境变量覆盖配置文件中的敏感信息：
- `DB_PASSWORD` - 数据库密码
- `REDIS_PASSWORD` - Redis密码
- `EMAIL_NAME` - 邮箱地址
- `EMAIL_PASSWORD` - 邮箱应用密码
- `SSH_TUNNEL_USERNAME` - SSH隧道用户名
- `SSH_TUNNEL_PASSWORD` - SSH隧道密码
- `CONFIG_PATH` - 配置文件路径（默认：/app/config/configs/config.yaml）