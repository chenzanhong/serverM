#!/bin/bash

# serverM项目一键部署脚本 - 使用集群内部数据库
# 适用于Kind创建的Kubernetes集群

echo "========================================"
echo "      serverM项目部署脚本（内部数据库）      "
echo "========================================"

# 检查kubectl是否安装
if ! command -v kubectl &> /dev/null; then
    echo "错误: kubectl命令未找到，请先安装kubectl并配置集群连接"
    exit 1
fi

# 创建命名空间
echo "1. 创建serverm命名空间..."
kubectl apply -f k8s/namespace.yaml
if [ $? -ne 0 ]; then
    echo "创建命名空间失败，请检查kubectl配置"
    exit 1
fi

# 部署内部数据库服务
echo "2. 部署内部数据库服务..."
kubectl apply -f k8s_internal/postgresql.yaml
kubectl apply -f k8s_internal/redis.yaml
kubectl apply -f k8s_internal/tdengine.yaml

# 等待数据库服务就绪
echo "3. 等待数据库服务就绪..."
sleep 10

# 应用内部数据库配置
echo "4. 应用内部数据库配置..."
kubectl apply -f k8s_internal/serverm-internal-config.yaml

# 将内部配置重命名为serverm-config
echo "5. 配置应用使用内部数据库..."
kubectl -n serverm delete configmap serverm-config 2>/dev/null || true
sleep 2
kubectl -n serverm patch configmap serverm-internal-config -p '{"metadata":{"name":"serverm-config"}}' --type=merge

# 应用Secret配置（假设已经准备好了）
echo "6. 应用Secret配置..."
kubectl apply -f k8s/secret.yaml

# 部署应用
echo "7. 部署serverM应用..."
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 部署可选服务
echo "8. 部署可选服务..."
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/ssh-service.yaml

echo "========================================"
echo "部署完成！正在验证部署状态..."
echo "========================================"

# 等待一段时间让Pod启动
sleep 10

# 显示部署状态
echo "数据库Pod状态:"
kubectl get pods -n serverm -l app in (postgres,redis,tdengine)

echo "\n应用Pod状态:"
kubectl get pods -n serverm -l app=serverm-server

echo "\n服务状态:"
kubectl get services -n serverm

echo "\n========================================"
echo "部署说明:"
echo "1. 应用将使用集群内部的数据库服务"
echo "2. PostgreSQL地址: postgres:5432"
echo "3. Redis地址: redis:6379"
echo "4. TDengine地址: tdengine:6030"
echo "\n使用以下命令访问应用:"
echo "kubectl port-forward -n serverm svc/serverm-server 8080:80"
echo "然后在浏览器访问 http://localhost:8080"
echo "========================================"