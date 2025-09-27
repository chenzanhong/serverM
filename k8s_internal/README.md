# Kubernetes集群内部数据库服务配置

本目录包含在Kubernetes集群内部提供数据库服务的配置文件，适用于测试环境。

## 配置文件说明

- **postgresql.yaml** - PostgreSQL数据库StatefulSet和Service配置
- **redis.yaml** - Redis缓存StatefulSet和Service配置
- **tdengine.yaml** - TDengine时序数据库StatefulSet和Service配置
- **serverm-internal-config.yaml** - 使用内部数据库服务的serverM配置

## 部署步骤

1. 部署数据库服务
   ```bash
   kubectl apply -f k8s_internal/postgresql.yaml
   kubectl apply -f k8s_internal/redis.yaml
   kubectl apply -f k8s_internal/tdengine.yaml
   ```

2. 部署serverM应用（使用内部数据库）
   ```bash
   # 先确保serverm命名空间存在
   kubectl apply -f k8s/namespace.yaml
   
   # 应用内部配置
   kubectl apply -f k8s_internal/serverm-internal-config.yaml
   
   # 应用secret（与外部配置相同）
   kubectl apply -f k8s/secret.yaml
   
   # 部署应用
   kubectl apply -f k8s/deployment.yaml
   kubectl apply -f k8s/service.yaml
   ```

## 注意事项

1. 这些配置适用于测试环境，生产环境请考虑使用持久化存储和备份策略
2. 默认配置使用emptyDir作为存储，Pod重启后数据将丢失
3. 请根据实际需求调整资源限制和副本数