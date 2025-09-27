# serverM 容器化部署优化说明

本文档总结了serverM项目的容器化部署优化，包括Dockerfile改进、配置加载机制优化、健康检查实现以及docker-compose配置。

## 一、Dockerfile 改进

### 1.1 主要改进点

1. **构建优化**
   - 优化缓存层，单独下载依赖
   - 使用 `-ldflags="-w -s"` 减小二进制文件体积
   - 添加构建工具支持

2. **配置灵活性**
   - 增加环境变量支持
   - 移除硬编码配置路径，支持Docker卷挂载

3. **健康检查**
   - 添加内置健康检查指令
   - 暴露SSH隧道所需端口

4. **错误处理增强**
   - SSH配置优化
   - 目录结构预创建

### 1.2 推荐使用方式

```bash
# 构建镜像
docker build -t serverm:latest .

# 运行容器
docker run -d \
  --name serverm-server \
  -p 8080:8080 \
  -p 30022:22 \
  -v ./config/configs:/app/config/configs \
  -v ./logs:/app/logs \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_NAME=serverm \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e REDIS_HOST=redis \
  -e REDIS_PORT=6379 \
  -e PUBLIC_SERVER_IP=localhost \
  serverm:latest
```

## 二、配置加载机制优化

### 2.1 主要改进点

1. **环境变量覆盖**
   - 扩展了环境变量覆盖范围，支持所有配置项
   - 支持通过环境变量动态调整服务行为

2. **默认配置**
   - 添加默认配置生成机制，当配置文件不存在时自动使用默认配置
   - 关键参数添加默认值，增强鲁棒性

3. **错误处理**
   - 改进运行时调用失败处理逻辑
   - 添加详细日志记录
   - 添加配置有效性验证（如端口范围检查）

4. **Docker适配**
   - 添加Docker默认配置路径
   - 优化配置加载流程，支持容器环境

### 2.2 环境变量列表

| 环境变量名 | 说明 | 默认值 |
|------------|------|--------|
| CONFIG_PATH | 配置文件路径 | /app/config/configs/config.yaml |
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_NAME | 数据库名 | serverm |
| DB_USER | 数据库用户名 | postgres |
| DB_PASSWORD | 数据库密码 | postgres |
| REDIS_HOST | Redis主机 | localhost |
| REDIS_PORT | Redis端口 | 6379 |
| REDIS_PASSWORD | Redis密码 | "" |
| REDIS_DB | Redis数据库索引 | "0" |
| PUBLIC_SERVER_IP | 公网服务器IP | localhost |
| SCRIPT_START_PORT | 反向SSH起始端口 | "30000" |
| SCRIPT_END_PORT | 反向SSH结束端口 | "30100" |
| SSH_TUNNEL_USERNAME | SSH隧道用户名 | "" |
| SSH_TUNNEL_PASSWORD | SSH隧道密码 | "" |
| GIN_MODE | Gin运行模式 | release |

## 三、健康检查实现

### 3.1 API健康检查端点

添加了 `/health` 公共端点，无需认证即可访问：

```
GET /health
```

返回示例：
```json
{
  "status": "ok",
  "service": "serverM"
}
```

### 3.2 Docker健康检查配置

Dockerfile中添加了健康检查指令：

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

## 四、Docker Compose 配置

### 4.1 功能概述

提供了完整的开发环境配置，包括：

1. **数据库服务**
   - PostgreSQL
   - Redis
   - TDengine (可选)

2. **应用服务**
   - 自动构建并运行serverM应用
   - 端口映射和卷挂载配置

3. **健康检查**
   - 所有服务都配置了健康检查
   - 依赖服务启动顺序控制

### 4.2 使用方法

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs serverm

# 停止服务
docker-compose down

# 停止并删除卷
docker-compose down -v
```

### 4.3 注意事项

1. 首次启动时，数据库需要一些时间初始化
2. 确保配置文件路径正确映射
3. 生产环境中应修改默认密码和敏感配置

## 五、Kubernetes 部署优化

### 5.1 推荐配置

在Kubernetes部署中，建议：

1. 使用ConfigMap存储配置文件
2. 使用Secret存储敏感信息
3. 添加健康检查探针
4. 设置资源限制和请求
5. 配置适当的存活和就绪探针

### 5.2 资源请求和限制示例

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 5.3 探针配置示例

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

## 六、常见问题排查

### 6.1 容器无法启动

1. 检查环境变量配置
2. 检查数据库连接是否正常
3. 查看容器日志：`docker logs serverm-server`

### 6.2 反向SSH连接失败

1. 确认`PUBLIC_SERVER_IP`配置正确
2. 检查端口范围配置
3. 验证防火墙设置是否允许相应端口

### 6.3 健康检查失败

1. 检查应用是否正常启动
2. 验证`/health`端点是否可访问
3. 增加启动延迟时间

## 七、性能优化建议

1. **数据库优化**
   - 为生产环境配置适当的数据库连接池
   - 考虑使用数据库连接复用

2. **资源优化**
   - 根据实际负载调整容器资源限制
   - 考虑使用多副本部署提高可用性

3. **日志管理**
   - 配置适当的日志级别
   - 考虑使用集中式日志收集系统