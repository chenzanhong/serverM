# serverM 

## 项目介绍
- 一个基于golang、gin、gorm、redis、postgresql等开发的服务器监控管理平台。这里只包含了后端部分和客户端代理部分。目前已测试过的可正常监控的Linux服务器发行版为centos和ubuntu。
- 该项目是软件工程综合实训的小组项目（我负责后端功能的实现和后期项目的对接调整，并进行项目展示），目前已实现基本的用户与服务器管理、监控代理（数据采集与上传）、数据可视化等功能。
- 前端：https://github.com/chenzanhong/WebApp 或 https://cnb.cool/szu/ServerMonitor/WebAPP/-/tree/test
- 后端：https://github.com/chenzanhong/serverM 或 https://cnb.cool/szu/ServerMonitor/Platform/-/tree/master
- 监控代理：https://github.com/chenzanhong/agent （完整代理项目）或 https://gitee.com/chenzanhong/agent （只含配置文件和可执行文件）
- 我后期对接并准备展示该项目过程中想到的问题和改进点：IMPROVEMENT.md

---

## 技术栈

- **语言和框架**：
  - [Go](https://golang.org/) + [Gin 框架](https://github.com/gin-gonic/gin)：高性能的 HTTP web 框架，用于构建 RESTful API 接口。

- **数据库**：
  - [PostgreSQL](https://www.postgresql.org/)：关系型数据库，作为主数据存储，持久化用户信息、主机信息及告警记录。
  - [Redis](https://redis.io/)：内存中的键值存储，用作 NoSQL 数据库，缓存监控数据并定期清除过期数据。

- **前端**：
  - [Vue2](https://v2.vuejs.org/) + [Element Plus](https://element-plus.org/)（见 [WebApp GitHub 仓库](https://github.com/chenzanhong/WebApp)）：提供用户友好的操作界面。

- **日志记录**：
  - 使用标准库 `log` 记录系统日志。
  - 使用 [zap](https://github.com/uber-go/zap) 进行结构化的用户操作日志记录。

- **邮件服务**：
  - SMTP + 163邮箱：用于发送注册验证邮件、重置密码邮件以及预警通知邮件等。

- **反向 SSH 连接**：
  - [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) 和 [github.com/pkg/sftp](https://github.com/pkg/sftp)：使用反向ssh服务（客户端代理通过autossh主动连接公网服务器，详细见 ReserverSSH.md 或脚本部分 /backend/server/handle/agent/getscript/getscript.go）用于远程服务器连接及文件传输。

- **并发模型**：
  - Go 原生支持的协程（goroutine）和通道（channel），用于实现高效的并行预警任务处理。
  - 使用 worker pool 模式处理并发请求，增强在高负载下的系统稳定性。

- **上下文管理**：
  - [context](https://golang.org/pkg/context/)：用于控制请求的生命周期和事务处理。

- **同步机制**：
  - 使用 `sync.Map` 和 `sync.Mutex` 来保证并发环境下对共享资源的安全访问（对预警状态表的互斥访问），防止竞态条件的发生。

---

![前端首页](/backend/server/docs/image_show/index.png)
其他前端页面展示见/backend/server/docs/image_show

## 项目架构

## 🏗 项目架构

本项目采用 **前后端分离的设计思想**，后端以 Go/Gin 为核心框架，结合 PostgreSQL、Redis 和 SSH 技术来构建一个服务器监控与管理平台。

---

### 🔧 架构层级说明

| 层级 | 组件 | 描述 |
|------|------|------|
| **前端层（WebApp）** | Vue2 + Element Plus | 用户操作界面，提供主机管理、监控视图、通知服务、用户操作日志、公司管理等功能。详见 [前端仓库](https://github.com/chenzanhong/WebApp) |
| **API 接口层** | Gin 框架 | 提供 RESTful API 接口，处理用户请求、身份认证、数据转发等 |
| **业务逻辑层** | Go modules | 实现主要业务逻辑，包括：用户权限控制、SSH 连接管理、文件传输、告警处理等 |
| **数据访问层** | GORM + PostgreSQL | 使用 ORM 对数据库进行操作，存储用户信息、主机状态、告警记录等 |
| **缓存层** | Redis | 缓存监控数据、阈值设定、会话信息等，提高系统响应速度 |
| **网络通信层** | SSH / SFTP | 使用 `golang.org/x/crypto/ssh` 和 `github.com/pkg/sftp` 实现远程连接与文件传输 |
| **消息通知层** | SMTP 邮件服务 | 基于 163 邮箱实现邮件注册验证、密码重置、系统预警通知等功能 |
| **异步任务层** | Goroutine + Worker Pool | 使用协程池异步处理告警判断，避免阻塞主线程，提升并发能力 |

---

### 🧱 目录结构概览

```
serverM/
├── server/                # 后端主服务
│   ├── main.go              # 程序入口
|   ├── asset/               # 数据库初始化数据文件和默认的脚本文件
│   ├── config/              # 配置文件加载模块
|   ├── docs/                # 早期 swagger 文档，中间弃用，改用 apifox 了
│   ├── handle/              # 控制器层和业务逻辑层，处理 HTTP 请求和实际业务处理
│   |   ├── admin/               # 公司管理员操作
│   |   ├── agent/               # 添加服务器以及代理相关的处理（脚本、预警）
│   |   ├── company /            # 公司的注册与信息列表
│   |   ├── email/               # 邮件模块
│   |   ├── monitor/             # 监控模块：服务器信息的获取，监控数据的接收（从客户端代理）、存储、获取，预警的处理和预警记录的获取，redis中监控数据的定期清除
│   |   ├── notice/              # 通知模块
│   |   ├── transfer/            # 文件传输模块
│   |   └──  user/               # 用户模块：注册、登录、更新，信息获取
|   ├── logs                 # 日志模块
│   ├── middleware/          # 中间件（JWT鉴权、CORS等）
│   ├── model/               # 数据模型定义（对应数据库表结构）和数据初始化模块
│   ├── redis/               # Redis 操作封装
│   └── router/              # 路由定义
├── agent/                   # 客户端代理程序（内网主机运行）
│   └── ...                  # 包括采集系统指标并发送给服务端的功能
└── script/                  # 脚本工具目录（如初始化脚本、部署脚本）
```

---

### ⚙️ 核心模块说明

| 模块名称 | 功能描述 |
|----------|-----------|
| **用户系统** | 支持注册、登录、权限控制、邮箱验证、密码找回 |
| **公司管理** | 公司注册、添加/删除成员、更换管理员 |
| **主机管理** | 添加、删除、查看受控主机，支持反向 SSH 穿透内网 |
| **系统监控** | 客户端定时上传 CPU、内存、网络、进程等信息，服务端接收并缓存至 Redis |
| **预警系统** | 判断是否超过阈值，使用 Worker Pool 异步处理预警逻辑，并通过邮件通知用户 |
| **文件传输** | 支持从 Web 界面向远程服务器上传文件，使用 SFTP 协议 |
| **日志系统** | 使用 zap 和 log 库分别记录用户操作日志和系统运行日志 |
| **反向 SSH 服务** | 公网服务器作为中继，允许内网主机主动建立隧道，实现内网穿透 |
| **邮件服务** | 提供邮件注册验证、密码重置、系统预警通知等功能 |
| **通知服务** | 用于注册公司、公司邀请成员和更换管理员等的通知与确认，以及其他系统公共服务通知 |

---

## 快速开始
1. **安装依赖**
```bash
    cd serverM/server
    go mod tidy
```
如果有拉取不完全的，可以使用 `go get + 依赖` 手动拉取。

2. **环境配置·**
在serverM/server/config/configs目录下创建一个config.yaml文件（或修改提供的config.yaml.example的文件名）添加以下内容：
```yaml
db: # pg数据库配置
  host: localhost # 数据库地址
  port: 5432   # 数据库端口
  name: database # 数据库名
  user: username    # 数据库用户名
  password: password # 数据库密码

redis:
  host: localhost
  port: 6379
  password:   # 没有密码，默认值为空
  db: 0

email:
  email_name: xxxx@163.com
  email_password: # 应用密码，非用户密码，请在163邮箱（网页版） -> 设置 -> POP3/SMTP/IMAP -> 开启服务 -> 开启IMAP/SMTP服务. POP3/SMTP服务 -> 保存开启后弹出窗口显示的应用密码（随后消失不可查看））

smtp_server:
  SMTPServer_host: smtp.163.com # 服务提供商，须与email中的邮箱相匹配
  SMTPServer_port: 465  # 端口号 25 在云服务器上可能会被限制，请用户自行更换

script:
  github_repo_url: https://gitee.com/chenzanhong/agent.git # agent仓库地址
  start_port: 10000 # 配置反向SSH服务可用的最小端口，端口号范围多大，可以配置反向ssh服务的服务器就有多少，建议根据项目实际需求自行更改
  end_port: 45000 # 配置反向SSH服务最大端口
  ssh_tunnel_username: xxxx # 反向SSH服务器用户名，部署该项目的服务器的用户名，建议创建一个专门的用户，并限制权限
  ssh_tunnel_password: xxxx # 反向SSH服务器密码，部署该项目的服务器的密码
  public_server_ip: x.x.x.x # 反向SSH服务器公网IP，部署该项目的服务器的公网ip


# 时序数据库配置，目前未使用，但未重构清理，建议保留这部分，以免运行失败
tdengine: # 时序数据库配置
  host: localhost # 数据库地址
  port: 6030 # 数据库端口
  name: servermonitor # 数据库名称
  user: root # 数据库用户名
  password: taosdata # 数据库密码

```

3. **服务端反向ssh服务配置**
见ReserveSHH.md

4. **启动**
先确保postgresql和redis服务正常运行中
```bash
    cd serverM/server
    go run main.go
```

---
## ❓ 常见问题

### Q: 启动时报错：`connection refused: connect to database failed`

A: 请确认 PostgreSQL 是否已启动，并检查 `config.yaml` 中数据库配置是否正确。

### Q: SSH 连接失败怎么办？

A: 请检查客户端代理是否正常运行、反向 SSH 配置是否正确，以及服务器防火墙是否开放对应端口。

### Q: 收不到邮件怎么办？

A: 请确认邮箱账号和应用密码是否填写正确，是否开启了 SMTP 服务。

### 如何调整代理程序上报监控数据的时间间隔？

A: 修改代理项目的配置文件 /agent/config/config.yaml，调整参数second的值（单位秒） 
```yaml
second: 30 
```

---

## 🔐 安全提示

- 不要将 `config.yaml` 提交到公共仓库
- 建议为反向 SSH 创建专用账户并限制权限
- 邮件密码应使用“应用密码”，而非登录密码
- 生产环境中应启用 HTTPS
