# 在Kubernetes集群内使用内部数据库服务运行serverM项目

本文档详细说明如何使用集群内部提供的数据库服务（PostgreSQL、Redis和TDengine）来运行serverM项目。

## 前提条件

- 已安装并配置好Kubernetes集群（Kind创建的3节点集群）
- 已安装kubectl命令行工具并配置好连接到集群
- 已构建并推送Docker镜像到镜像仓库（已使用阿里云个人镜像仓库）

## 部署步骤

### 1. 确保命名空间存在

```bash
# 创建serverm命名空间
kubectl apply -f k8s/namespace.yaml

# 检查命名空间是否创建成功
kubectl get namespace serverm
```

### 2. 部署集群内部的数据库服务

```bash
# 部署PostgreSQL数据库
kubectl apply -f k8s_internal/postgresql.yaml

# 部署Redis缓存
kubectl apply -f k8s_internal/redis.yaml

# 部署TDengine时序数据库
kubectl apply -f k8s_internal/tdengine.yaml

# 检查数据库服务是否部署成功
kubectl get pods -n serverm -l app in (postgres,redis,tdengine)
```

### 3. 准备敏感配置信息

使用generate-secrets.sh脚本生成敏感信息的base64编码，然后更新secret.yaml文件：

```bash
# 运行脚本生成编码
cd k8s
./generate-secrets.sh

# 编辑secret.yaml文件，替换占位符为生成的编码值
# 对于内部数据库，数据库密码需要与数据库服务的配置保持一致
# PostgreSQL密码: admin123
# Redis密码: 留空（内部Redis配置未设置密码）
# TDengine密码: taosdata
```

应用Secret配置：

```bash
kubectl apply -f k8s/secret.yaml
```

### 4. 应用内部数据库配置

```bash
# 应用使用内部数据库的配置
kubectl apply -f k8s_internal/serverm-internal-config.yaml

# 将ConfigMap重命名为serverm-config，确保应用能正确识别
kubectl -n serverm patch configmap serverm-internal-config -p '{"metadata":{"name":"serverm-config"}}' --type=merge
```

### 5. 部署serverM应用

```bash
# 部署应用
kubectl apply -f k8s/deployment.yaml

# 部署内部服务
kubectl apply -f k8s/service.yaml

# 如果需要外部访问，部署Ingress
kubectl apply -f k8s/ingress.yaml

# 如果需要SSH隧道服务，部署SSH服务
kubectl apply -f k8s/ssh-service.yaml
```

## 验证部署

### 检查Pod状态

```bash
# 检查所有Pod状态
kubectl get pods -n serverm

# 查看serverM应用日志
kubectl logs -n serverm -l app=serverm-server
```

### 检查服务状态

```bash
# 检查所有服务
kubectl get services -n serverm

# 验证数据库连接（可选，需要安装相应的客户端工具）
kubectl run -it --rm --restart=Never postgres-client --image=postgres:13-alpine -- psql -h postgres -U admin -d serverm
```

### 访问应用

如果部署了Ingress，可以通过配置的域名访问应用。否则，可以使用端口转发进行测试：

```bash
# 端口转发到serverM服务
kubectl port-forward -n serverm svc/serverm-server 8080:80

# 然后在浏览器访问 http://localhost:8080
```

## 常见问题排查

1. **Pod启动失败**
   - 查看Pod详细信息：`kubectl describe pod -n serverm [pod-name]`
   - 检查镜像是否能正常拉取
   - 检查配置是否正确挂载

2. **数据库连接问题**
   - 确保数据库服务Pod已正常运行
   - 验证数据库配置（主机名、端口、用户名、密码）是否正确
   - 查看应用日志中的连接错误信息

3. **权限问题**
   - 确保服务账户有足够的权限
   - 检查Secret和ConfigMap的访问权限

## 扩展和优化

1. **持久化存储**：在生产环境中，应配置持久化存储来保存数据库数据

2. **资源限制**：根据实际负载调整各组件的资源限制

3. **监控和日志**：配置Prometheus和Grafana监控服务状态，使用ELK收集和分析日志

4. **高可用性**：对于生产环境，考虑配置多副本和适当的健康检查机制