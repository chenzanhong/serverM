#!/bin/bash

# 辅助脚本：生成base64编码的敏感信息，用于Kubernetes Secret

echo "=== 敏感信息Base64编码生成器 ==="
echo ""

# 数据库密码
echo "请输入数据库密码："
read -s db_password
db_password_encoded=$(echo -n "$db_password" | base64)
echo "数据库密码Base64编码：$db_password_encoded"
echo ""

# Redis密码（如果有）
echo "请输入Redis密码（可选，直接回车跳过）："
read -s redis_password
if [ -n "$redis_password" ]; then
    redis_password_encoded=$(echo -n "$redis_password" | base64)
    echo "Redis密码Base64编码：$redis_password_encoded"
else
    echo "Redis密码为空，跳过"
fi
echo ""

# 邮箱地址
echo "请输入163邮箱地址："
read email_name
email_name_encoded=$(echo -n "$email_name" | base64)
echo "邮箱地址Base64编码：$email_name_encoded"
echo ""

# 邮箱应用密码
echo "请输入163邮箱应用密码："
read -s email_password
email_password_encoded=$(echo -n "$email_password" | base64)
echo "邮箱应用密码Base64编码：$email_password_encoded"
echo ""

# SSH隧道用户名
echo "请输入SSH隧道用户名："
read ssh_tunnel_username
ssh_tunnel_username_encoded=$(echo -n "$ssh_tunnel_username" | base64)
echo "SSH隧道用户名Base64编码：$ssh_tunnel_username_encoded"
echo ""

# SSH隧道密码
echo "请输入SSH隧道密码："
read -s ssh_tunnel_password
ssh_tunnel_password_encoded=$(echo -n "$ssh_tunnel_password" | base64)
echo "SSH隧道密码Base64编码：$ssh_tunnel_password_encoded"
echo ""

echo "=== 编码完成，请将以上编码值复制到secret.yaml文件中对应的字段 ==="
echo ""
echo "使用示例："
echo "kubectl apply -f k8s/namespace.yaml"
echo "kubectl apply -f k8s/configmap.yaml"
echo "kubectl apply -f k8s/secret.yaml  # 确保已替换编码值"
echo "kubectl apply -f k8s/deployment.yaml"
echo "kubectl apply -f k8s/service.yaml"
echo "kubectl apply -f k8s/ingress.yaml  # 如果需要外部访问"
echo "kubectl apply -f k8s/ssh-service.yaml  # 如果需要SSH隧道服务"