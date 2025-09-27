# serverM Kubernetes部署指南

本目录包含了将serverM项目部署到Kubernetes集群所需的所有配置文件。

## 文件说明

- **namespace.yaml** - 创建Kubernetes命名空间
- **configmap.yaml** - 存储非敏感配置信息
- **secret.yaml** - 存储敏感配置信息（需要替换为实际的base64编码值）
- **deployment.yaml** - 定义后端服务的部署配置
- **service.yaml** - 定义集群内部服务访问方式
- **ingress.yaml** - 定义外部访问方式（可选）
- **ssh-service.yaml** - 定义SSH隧道服务（可选）
- **generate-secrets.sh** - 辅助脚本，用于生成base64编码的敏感信息

## 部署准备

1. **修改配置文件**
   - 更新 `configmap.yaml` 中的公共配置，特别是 `public_server_ip`
   - 使用 `generate-secrets.sh` 脚本生成敏感信息的base64编码，并更新 `secret.yaml`
   - 更新 `deployment.yaml` 中的镜像地址

2. **构建Docker镜像**
   ```bash
   docker build -t your-registry/serverm:latest .
   docker push your-registry/serverm:latest
   ```

## 部署步骤

按照以下顺序应用配置文件：

```bash
# 创建命名空间
kubectl apply -f namespace.yaml

# 创建配置和密钥
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml

# 部署应用
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# 可选：配置外部访问和SSH服务
kubectl apply -f ingress.yaml
kubectl apply -f ssh-service.yaml
```

## 验证部署

```bash
# 检查Pod状态
kubectl get pods -n serverm

# 检查服务状态
kubectl get services -n serverm

# 查看日志
kubectl logs -n serverm -l app=serverm-server
```

## 注意事项

- 确保PostgreSQL和Redis服务可访问
- 对于SSH隧道功能，可能需要特殊的网络配置
- 生产环境中建议使用外部托管的数据库服务
- 定期备份数据库和重要配置

详细的部署说明请参考项目根目录下的 `K8S_DEPLOYMENT.md` 文件。