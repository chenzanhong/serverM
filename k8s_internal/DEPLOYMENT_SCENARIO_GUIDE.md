# serverM在Kind集群中的部署与测试指南

## 场景说明

本文档适用于以下部署场景：

1. 在Kind创建的Kubernetes集群中部署serverM后端服务
2. 在本地宿主机上运行前端项目
3. 在另一台内网服务器A上安装代理，测试监控功能

## 前提条件

1. Kind集群已正确安装和配置
2. kubectl已正确配置连接到Kind集群
3. 宿主机上的数据库服务（PostgreSQL、Redis、TDengine）已启动并允许访问
4. 本地宿主机和服务器A处于同一内网

## 部署步骤

### 1. 准备配置文件

确保以下配置文件已正确设置：

#### 1.1 配置public_server_ip

`public_server_ip`已设置为宿主机的内网IP地址，确保服务器A能访问到：

```yaml
# 在configmap.yaml和serverm-internal-config.yaml中
public_server_ip: "192.168.1.100"  # 替换为您宿主机的实际内网IP
```

#### 1.2 配置host-services.yaml

确保Endpoints正确指向宿主机内网IP：

```yaml
# 在host-services.yaml中
addresses:
- ip: 192.168.1.100  # 与public_server_ip保持一致
```

### 2. 部署内部数据库服务（推荐）

如果您选择使用集群内部数据库（而不是宿主机数据库），请执行：

```bash
# 创建命名空间
kubectl create namespace serverm

# 部署内部数据库服务
kubectl apply -f k8s_internal/postgresql.yaml
kubectl apply -f k8s_internal/redis.yaml
kubectl apply -f k8s_internal/tdengine.yaml
```

### 3. 应用配置

根据您选择的数据库部署方式，应用相应的配置：

#### 使用内部数据库

```bash
kubectl apply -f k8s_internal/serverm-internal-config.yaml
```

#### 使用宿主机数据库

```bash
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/host-services.yaml
```

### 4. 准备敏感信息

```bash
# 生成base64编码的敏感信息并更新secret.yaml
# 然后应用secret
kubectl apply -f k8s/secret.yaml
```

### 5. 部署应用

```bash
# 部署应用
kubectl apply -f k8s/deployment.yaml

# 部署服务
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ssh-service.yaml
```

### 6. 配置端口转发

为了从本地宿主机访问服务，配置端口转发：

```bash
# 为API服务配置端口转发
kubectl port-forward -n serverm svc/serverm-server 8080:80

# 在另一个终端为SSH服务配置端口转发
kubectl port-forward -n serverm svc/serverm-ssh 30022:22
```

## 本地宿主机前端配置

1. 在本地宿主机上运行前端项目
2. 配置前端API地址指向：`http://localhost:8080`

## 内网服务器A上的代理安装

### 1. 获取安装脚本

在服务器A上，通过浏览器访问前端界面（使用宿主机内网IP）：

```
http://192.168.1.100:8080
```

登录系统并获取代理安装脚本。

### 2. 安装代理

在服务器A上执行获取到的脚本：

```bash
chmod +x install_script.sh
./install_script.sh
```

脚本将：
- 下载agent程序
- 安装为系统服务
- 建立反向SSH隧道连接到serverM服务器

### 3. 验证代理连接

代理安装完成后，可以在serverM前端界面查看服务器A的状态。如果连接成功，您将看到：
- 服务器A显示在线状态
- 可以查看系统监控数据
- 可以通过反向SSH隧道进行远程操作

## 注意事项

1. **防火墙配置**：
   - 确保宿主机防火墙允许端口8080和30022的入站连接
   - 确保服务器A可以访问宿主机的这些端口

2. **IP地址一致性**：
   - `public_server_ip`必须与`host-services.yaml`中的IP地址一致
   - 确保使用的是宿主机的实际内网IP，而不是localhost或127.0.0.1

3. **端口冲突**：
   - 确保本地宿主机的8080和30022端口未被占用
   - 如果有冲突，可以修改端口转发配置

## 故障排查

1. **代理无法连接**：
   - 检查`public_server_ip`配置是否正确
   - 验证服务器A能否ping通宿主机内网IP
   - 检查防火墙设置

2. **数据库连接失败**：
   - 确认使用了正确的数据库连接配置（内部数据库或宿主机数据库）
   - 验证数据库服务是否正常运行

3. **前端无法访问API**：
   - 确认端口转发是否正常工作
   - 检查服务是否成功部署并运行

## 扩展说明

1. **持久化数据**：
   - 对于生产环境，建议为数据库配置持久化存储
   - 可以使用PersistentVolumeClaim替代emptyDir

2. **高可用性**：
   - 考虑使用多副本部署提高应用可用性
   - 对于数据库服务，使用StatefulSet确保稳定的网络标识