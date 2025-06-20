# software-engineer


## 项目介绍


## 技术栈


## 项目架构


## 项目部署
### 部署
1. **安装依赖**
```bash
    cd serverM/server
    go mod tidy
```
如果有拉取不完全的，可以使用 `go get + 依赖` 手动拉取。


2. **环境配置·**
在serverM/server/config/configs目录下创建一个config.yaml文件（或修改提供的config.yaml.example文件）添加以下内容：
```yaml
db: # 数据库配置
  host: localhost # 数据库地址
  port: "5432"   # 数据库端口
  name: database # 数据库名
  user: username    # 数据库用户名
  password: password # 数据库密码

oss: #aliyun oss 配置
  OSS_REGION: oss-cn-shenzhen # oss区域
  OSS_ACCESS_KEY_ID: # oss key
  OSS_ACCESS_KEY_SECRET:  # oss密钥
  OSS_BUCKET:  # oss bucket

email:
    email_name: xxxx@163.com
    email_password: # 应用密码
    #（在163邮箱 -> 设置 -> POP3/SMTP/IMAP -> 开启服务 -> 开启IMAP/SMTP服务. POP3/SMTP服务 -> 保存开启后弹出窗口显示的应用密码（随后消失不可查看））
    base_url: # ngrok提供的url
    # 密码找回的方式二是发送验证码，可以不用暴露内网，不使用这个base_url
    # 使用ngrok暴露内网（测试环境）
    # ngrok http --url=ox-driven-shortly.ngrok-free.app 80
    # ngrok http http://localhost:8080
    # 安装ngrok:https://dashboard.ngrok.com/get-started/setup/windows

smtp_server:
  SMTPServer_host: smtp.163.com
  SMTPServer_port: 25
```

3. **启动**
```bash
    cd serverM/server
    go run main.go
```

## 其他说明
<!-- 1. **密码找回功能接口测试**
启动ngork服务，通过ngrok http 暴露本地8080端口，并在serverM/server/config/configs中配置base_url；
访问登录localhost:8080/agent/login，获取token；然后在请求头中添加token，访问localhost:8080/agent/request_reset_password，参数为
```json
{
    "email":"xxx@xxx.com"  // 你注册时使用的邮箱
}
```
然后会收到邮箱，点击邮件上的链接跳转至密码找回页面，输入新密码，点击确定，即可更新密码 -->

