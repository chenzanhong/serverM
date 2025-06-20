---
title: ServerM
language_tabs:
  - shell: Shell
  - http: HTTP
  - javascript: JavaScript
  - ruby: Ruby
  - python: Python
  - php: PHP
  - java: Java
  - go: Go
toc_footers: []
includes: []
search: true
code_clipboard: true
highlight_theme: darkula
headingLevel: 2
generator: "@tarslib/widdershins v4.0.30"

---

# ServerM
- 可能还存有部分接口的介绍随着项目的开发未及时更新

Base URLs:

# Authentication

# Default

## GET 查询用户所有主机信息/agent/list

GET /agent/list

该接口用于查询当前用户的所有主机信息，支持按时间范围过滤。

##请求格式
- **URL**: `/agent/list`
- **Method**: `GET`
- **Content-Type**: `application/json`
- **Authorization**: `your_jwt_token`

## 请求参数
| 参数名 | 类型   | 必填 | 说明                                                                 |
|--------|--------|------|--------------------------------------------------------------------|
| from   | string | 否   | 起始时间，格式为 `RFC3339`（如 `2023-01-01T00:00:00Z`），默认为 `1970-01-01T00:00:00Z` |
| to     | string | 否   | 结束时间，格式为 `RFC3339`（如 `2023-12-31T23:59:59Z`），默认为 `9999-12-31T23:59:59Z` |

## 响应格式
- **Content-Type**: `application/json`
- **响应示例**:
```json
[
    {
        "id": 4,
        "host_name": "host1",
        "ip": "127.0.0.4",
        "os": "Linux",
        "platform": "Ubuntu 20.04",
        "kernel_arch": "x86_64",
        "host_info_created_at": "2025-05-18T13:50:27.602156Z",
        "token": ""
    },
    {
        "id": 9,
        "host_name": "virtual machine",
        "ip": "",
        "os": "Linux",
        "platform": "ubuntu",
        "kernel_arch": "x86_64",
        "host_info_created_at": "2025-05-18T13:52:40.66778Z",
        "token": ""
    },
    {
        "id": 10,
        "host_name": "1234",
        "ip": "192.168.130.128",
        "os": "Linux",
        "platform": "ubuntu",
        "kernel_arch": "x86_64",
        "host_info_created_at": "2025-05-18T13:59:08.883244Z",
        "token": ""
    },
    {
        "id": 12,
        "host_name": "5678",
        "ip": "192.168.130.129",
        "os": "Linux",
        "platform": "ubuntu",
        "kernel_arch": "x86_64",
        "host_info_created_at": "2025-05-19T16:21:40.504292Z",
        "token": ""
    }
]

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|from|query|string| 否 |none|
|to|query|string| 否 |none|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "hosts": [
    {
      "id": 10,
      "user_name": "user1",
      "host_name": "user_test3",
      "ip": "192.168.130.129",
      "os": "Linux",
      "platform": "ubuntu",
      "kernel_arch": "amd64",
      "host_info_created_at": "2025-06-10T21:29:22.301204Z",
      "company_id": 1,
      "mem_threshold": 0.9,
      "cpu_threshold": 0.9
    },
    {
      "id": 4,
      "user_name": "user1",
      "host_name": "host1",
      "ip": "127.0.0.4",
      "os": "Linux",
      "platform": "Ubuntu 20.04",
      "kernel_arch": "x86_64",
      "host_info_created_at": "2025-06-10T21:20:55.215489Z",
      "company_id": 1,
      "mem_threshold": 0.8,
      "cpu_threshold": 0.8
    }
  ]
}
```

```json
{
  "error": "无效的 from 时间格式"
}
```

```json
{
  "error": "无效的 to 时间格式"
}
```

```json
{
  "error": "Failed to query host_info",
  "details": ""
}
```

```json
{
  "error": "Failed to scan host_info",
  "details": ""
}
```

```json
{
  "error": "Error occurred during iteration",
  "details": ""
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|string|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## GET 查询指定主机的信息/monitor/{hostname}

GET /agent/monitor/{hostname}

该接口用于查询特定主机的详细信息，包括 CPU、内存、网络和进程等数据。

## 请求格式
- **URL**: `/monitor/:hostname`(hostname填写要查询的主机名)
- **Method**: `GET`
- **Content-Type**: `application/json`

## 请求参数
### URL 参数
| 参数名     | 类型   | 必填 | 说明         |
|------------|--------|------|--------------|
| host_name  | string | 是   | 主机名       |

### Query 参数
| 参数名 | 类型   | 必填 | 说明                                                                 |
|--------|--------|------|--------------------------------------------------------------------|
| type   | string | 否   | 查询类型，默认为 `all`（返回所有信息），可选值：`cpu`, `memory`, `net`, `process` |
| from   | string | 否   | 起始时间，格式为 `RFC3339`（如 `2023-01-01T00:00:00Z`），默认为 `1970-01-01T00:00:00Z` |
| to     | string | 否   | 结束时间，格式为 `RFC3339`（如 `2023-12-31T23:59:59Z`），默认为 `9999-12-31T23:59:59Z` |

## 响应格式
- **Content-Type**: `application/json`
  - **响应示例**:
```json 
{
  "cpu": 
  [
    {
      "data": 
            { 
              "cores_num": 6, 
              "cpu_info_created_at": "0001-01-01T00:00:00Z", 
              "id": 0, 
              "model_name": "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz", 
              "percent": 25.5
            }, 
      "time": "2025-03-10T10:17:16Z"
    } 
  ], 
  "host": 
  {
    "host_info_created_at": "2025-03-10T18:17:16.189247Z", 
    "hostname": "my-host", 
    "id": 2, 
    "kernel_arch": "x86_64", 
    "os": "Linux", 
    "platform": "Ubuntu 20.04"
  }, 
  "memory": 
  [ 
    { 
      "data": { 
        "available": "8GB",
        "free": "8GB",
        "id": 0,
        "mem_info_created_at": "0001-01-01T00:00:00Z",
        "total": "16GB",
        "used": "8GB",
        "user_percent": 50
      },
      "time": "2025-03-10T10:17:16Z"
    } 
  ], 
  "net": [ 
    { 
      "data": { 
        "bytes_recv": 1024, 
        "bytes_sent": 2048, 
        "id": 0, 
        "name": "eth0", 
        "net_info_created_at": "0001-01-01T00:00:00Z"
      }, 
      "time": "2025-03-10T10:17:16Z"
    } 
  ], 
  "process": [
    {
      "data": 
      { 
        "cmdline": "/usr/bin/python3", 
        "cpu_percent": 10.5, 
        "id": 0, 
        "mem_percent": 5.5, 
        "pid": 1234, 
        "pro_info_created_at": "0001-01-01T00:00:00Z"
      }, 
      "time": "2025-03-10T10:17:16Z" } 
  ]
}

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|hostname|path|string| 是 |none|
|Authorization|header|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "cpu": [
    {
      "data": {
        "cores_num": 6,
        "cpu_info_created_at": "0001-01-01T00:00:00Z",
        "id": 0,
        "model_name": "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz",
        "percent": 25.5
      },
      "time": "2025-03-10T10:17:16Z"
    }
  ],
  "host": {
    "host_info_created_at": "2025-03-10T18:17:16.189247Z",
    "hostname": "my-host",
    "id": 2,
    "kernel_arch": "x86_64",
    "os": "Linux",
    "platform": "Ubuntu 20.04"
  },
  "memory": [
    {
      "data": {
        "available": "8GB",
        "free": "8GB",
        "id": 0,
        "mem_info_created_at": "0001-01-01T00:00:00Z",
        "total": "16GB",
        "used": "8GB",
        "user_percent": 50
      },
      "time": "2025-03-10T10:17:16Z"
    }
  ],
  "net": [
    {
      "data": {
        "bytes_recv": 1024,
        "bytes_sent": 2048,
        "id": 0,
        "name": "eth0",
        "net_info_created_at": "0001-01-01T00:00:00Z"
      },
      "time": "2025-03-10T10:17:16Z"
    }
  ],
  "process": [
    {
      "data": {
        "cmdline": "/usr/bin/python3",
        "cpu_percent": 10.5,
        "id": 0,
        "mem_percent": 5.5,
        "pid": 1234,
        "pro_info_created_at": "0001-01-01T00:00:00Z"
      },
      "time": "2025-03-10T10:17:16Z"
    }
  ]
}
```

> 500 Response

```json
{
  "error": ""
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» cpu_info|[object]|true|none||none|
|»» cpu_info_created_at|string|false|none||none|
|»» id|integer|false|none||none|
|»» model_name|string|false|none||none|
|»» percent|number|false|none||none|
|»» time|string|false|none||none|
|» host_info|object|true|none||none|
|»» host_info_created_at|string|true|none||none|
|»» hostname|string|true|none||none|
|»» id|integer|true|none||none|
|»» kernel_arch|string|true|none||none|
|»» os|string|true|none||none|
|»» platform|string|true|none||none|
|» mem_info|[object]|true|none||none|
|»» available|string|false|none||none|
|»» free|string|false|none||none|
|»» id|integer|false|none||none|
|»» mem_info_created_at|string|false|none||none|
|»» total|string|false|none||none|
|»» used|string|false|none||none|
|»» user_percent|integer|false|none||none|
|»» time|string|false|none||none|
|» net_info|[object]|true|none||none|
|»» bytes_recv|integer|false|none||none|
|»» bytes_sent|integer|false|none||none|
|»» id|integer|false|none||none|
|»» name|string|false|none||none|
|»» net_info_created_at|string|false|none||none|
|»» time|string|false|none||none|
|» pro_info|[object]|true|none||none|
|»» cmdline|string|false|none||none|
|»» cpu_percent|number|false|none||none|
|»» id|integer|false|none||none|
|»» mem_percent|number|false|none||none|
|»» pid|integer|false|none||none|
|»» pro_info_created_at|string|false|none||none|
|»» time|string|false|none||none|

## POST 安装采集器接口

POST /agent/install

该接口用于向指定主机安装并启动采集客户端，接收参数包括：主机ip，端口号，用户名，密码以及主机名，安装成功后会定时（默认30s）向服务端发送采集得到的客户端信息。

> Body 请求参数

```json
"// {\n//     \"host\": \"x.x.x.x\",\n//     \"user\": \"czh\",\n//     \"password\": \"123456\",\n//     \"port\": 22,\n//     \"host_name\": \"czh-centos\",\n//     \"os\": \"Linux\",\n//     \"platform\": \"centos\",\n//     \"kernel_arch\": \"amd64\",\n//     \"mem_threshold\": 90,\n//     \"cpu_threshold\": 95\n// }\n{\n    \"host\": \"192.168.202.130\",\n    \"user\": \"czh\",\n    \"password\": \"password\",\n    \"port\": 22,\n    \"host_name\": \"czh-ubuntu\",\n    \"os\": \"Linux\",\n    \"platform\": \"ubuntu\",\n    \"kernel_arch\": \"amd64\",\n    \"mem_threshold\": 90,\n    \"cpu_threshold\": 95\n}"
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|

> 返回示例

> 200 Response

```json
"// 响应一个脚本文件，不是json"
```

> 400 Response

```json
{
  "error": ""
}
```

> 401 Response

```json
{
  "code": 401,
  "success": false,
  "message": "未找到用户信息"
}
```

> 409 Response

```json
{
  "error": "host_name '111' already exists"
}
```

> 500 Response

```json
{
  "error": "Failed to check host_name "
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|
|» host_name|string|true|none||none|
|» token|string|true|none||none|

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» error|string|true|none||none|

状态码 **500**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» error|string|true|none||none|

## POST 添加系统信息/agent/addSystemInfo

POST /agent/addSystemInfo

增加服务器信息
接收参数：
type RequestData struct {
	CPUInfo  []model.CPUInfo   `json:"cpu_info"`  // CPU 信息
	HostInfo model.HostInfo    `json:"host_info"` // 主机信息
	MemInfo  model.MemoryInfo  `json:"mem_info"`  // 内存信息
	ProInfo  model.ProcessInfo `json:"pro_info"`  // 进程信息
	NetInfo  model.NetworkInfo `json:"net_info"`  // 网络信息
}

> Body 请求参数

```json
{
  "host_info": {
    "host_name": "abc",
    "os": "Linux",
    "platform": "Ubuntu 20.04",
    "kernel_arch": "x86_64",
    "token": ""
  },
  "cpu_info": [
    {
      "model_name": "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz",
      "cores_num": 6,
      "percent": 25.5
    },
    {
      "model_name": "Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz",
      "cores_num": 6,
      "percent": 25.5
    }
  ],
  "mem_info": {
    "total": "16GB",
    "available": "8GB",
    "used": "8GB",
    "free": "8GB",
    "user_percent": 50
  },
  "pro_info": {
    "pid": 1234,
    "cpu_percent": 10.5,
    "mem_percent": 5.5,
    "cmdline": "/usr/bin/python3"
  },
  "net_info": {
    "name": "eth0",
    "bytes_recv": 1024,
    "bytes_sent": 2048
  }
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 否 |none|
|» host_info|body|object| 是 |none|
|»» host_name|body|string| 是 |none|
|»» os|body|string| 是 |none|
|»» platform|body|string| 是 |none|
|»» kernel_arch|body|string| 是 |none|
|»» token|body|string| 是 |none|
|» cpu_info|body|[object]| 是 |none|
|»» model_name|body|string| 是 |none|
|»» cores_num|body|integer| 是 |none|
|»» percent|body|number| 是 |none|
|» mem_info|body|object| 是 |none|
|»» total|body|string| 是 |none|
|»» available|body|string| 是 |none|
|»» used|body|string| 是 |none|
|»» free|body|string| 是 |none|
|»» user_percent|body|integer| 是 |none|
|» pro_info|body|object| 是 |none|
|»» pid|body|integer| 是 |none|
|»» cpu_percent|body|number| 是 |none|
|»» mem_percent|body|number| 是 |none|
|»» cmdline|body|string| 是 |none|
|» net_info|body|object| 是 |none|
|»» name|body|string| 是 |none|
|»» bytes_recv|body|integer| 是 |none|
|»» bytes_sent|body|integer| 是 |none|

> 返回示例

> 201 Response

```json
{
  "message": "System information inserted successfully"
}
```

> 400 Response

```json
{
  "error": "Invalid JSON data:"
}
```

```json
{
  "error": "Failed to update heartbeat and status"
}
```

```json
{
  "error": "Failed to insert host info: "
}
```

```json
{
  "error": "Failed to insert host and token info: "
}
```

```json
{
  "error": "Failed to insert system info: "
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|201|[Created](https://tools.ietf.org/html/rfc7231#section-6.3.2)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## GET 查询主机实时信息

GET /agent/monitor/status/{hostname}

该接口用于查询特定主机的最新一条详细信息，包括 CPU、内存、网络和进程等数据，用于实时信息展示。

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|hostname|path|string| 是 |none|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "alert_messages": "CPU与内存告警",
  "data": {
    "cpu_info": [
      {
        "id": 0,
        "model_name": "General Purpose Processor",
        "cores_num": 1,
        "percent": 3.044412607440411,
        "cpu_info_created_at": "2025-06-13T00:04:31.117430674+08:00"
      },
      {
        "id": 0,
        "model_name": "General Purpose Processor",
        "cores_num": 1,
        "percent": 3.044412607440411,
        "cpu_info_created_at": "2025-06-13T00:04:31.117431064+08:00"
      }
    ],
    "host_info": {
      "id": 0,
      "user_name": "",
      "host_name": "password",
      "ip": "",
      "port": 0,
      "os": "linux",
      "platform": "centos-7.9.2009 rhel",
      "kernel_arch": "x86_64",
      "host_info_created_at": "2025-06-13T00:04:31.121717609+08:00",
      "token": "48f5ac5eff7511e5",
      "cpu_threshold": 0,
      "mem_threshold": 0
    },
    "mem_info": {
      "id": 0,
      "total": "1.93G",
      "available": "0.83G",
      "used": "0.82G",
      "free": "0.27G",
      "user_percent": 42.47,
      "mem_info_created_at": "2025-06-13T00:04:31.117487125+08:00"
    },
    "pro_info": [
      {
        "id": 0,
        "pid": 5109,
        "cpu_percent": 0.6379552277413544,
        "mem_percent": 0.30149877,
        "pro_info_created_at": "2025-06-13T00:04:31.1350454+08:00"
      },
      {
        "id": 0,
        "pid": 27411,
        "cpu_percent": 0.3130104701376684,
        "mem_percent": 4.5762753,
        "pro_info_created_at": "2025-06-13T00:04:31.137759913+08:00"
      }
    ],
    "net_info": [
      {
        "id": 0,
        "name": "eth0",
        "bytes_recv": 2703761386,
        "bytes_sent": 316717010,
        "net_info_created_at": "2025-06-13T00:04:31.138430754+08:00"
      },
      {
        "id": 0,
        "name": "lo",
        "bytes_recv": 5312406381,
        "bytes_sent": 5312406381,
        "net_info_created_at": "2025-06-13T00:04:31.138431065+08:00"
      }
    ]
  }
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» alert_messages|string|true|none||none|
|» data|object|true|none||none|
|»» cpu_info|[object]|true|none||none|
|»»» id|integer|true|none||none|
|»»» model_name|string|true|none||none|
|»»» cores_num|integer|true|none||none|
|»»» percent|number|true|none||none|
|»»» cpu_info_created_at|string|true|none||none|
|»» host_info|object|true|none||none|
|»»» id|integer|true|none||none|
|»»» user_name|string|true|none||none|
|»»» host_name|string|true|none||none|
|»»» ip|string|true|none||none|
|»»» port|integer|true|none||none|
|»»» os|string|true|none||none|
|»»» platform|string|true|none||none|
|»»» kernel_arch|string|true|none||none|
|»»» host_info_created_at|string|true|none||none|
|»»» token|string|true|none||none|
|»»» cpu_threshold|integer|true|none||none|
|»»» mem_threshold|integer|true|none||none|
|»» mem_info|object|true|none||none|
|»»» id|integer|true|none||none|
|»»» total|string|true|none||none|
|»»» available|string|true|none||none|
|»»» used|string|true|none||none|
|»»» free|string|true|none||none|
|»»» user_percent|number|true|none||none|
|»»» mem_info_created_at|string|true|none||none|
|»» pro_info|[object]|true|none||none|
|»»» id|integer|true|none||none|
|»»» pid|integer|true|none||none|
|»»» cpu_percent|number|true|none||none|
|»»» mem_percent|number|true|none||none|
|»»» pro_info_created_at|string|true|none||none|
|»» net_info|[object]|true|none||none|
|»»» id|integer|true|none||none|
|»»» name|string|true|none||none|
|»»» bytes_recv|integer|true|none||none|
|»»» bytes_sent|integer|true|none||none|
|»»» net_info_created_at|string|true|none||none|

## POST 删除采集器接口

POST /agent/delete

该接口用于删除指定主机的采集器客户端。

> Body 请求参数

```json
{
  "ip": "192.168.56.131",
  "host_name": "virtual machine"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |your_jwt_token|
|body|body|object| 否 |none|
|» ip|body|string| 是 |none|
|» host_name|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "采集器删除成功"
}
```

```json
{
  "请求有误": null
}
```

```json
{
  "error": ""
}
```

> 401 Response

```json
{
  "error": ""
}
```

> 500 Response

```json
{
  "error": "服务器错误"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## POST 修改预警阈值

POST /agent/setthreshold

该接口用于设置内存预警、CPU预警的阈值

> Body 请求参数

```json
{
  "hostname": "czh-centos",
  "mem_threshold": 50.5,
  "cpu_threshold": 60
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» hostname|body|string| 是 |none|
|» mem_threshold|body|number| 是 |none|
|» cpu_threshold|body|number| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "设置阈值成功"
}
```

> 400 Response

```json
{
  "message": "请求有误"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|

### 返回数据结构

## GET 获取预警详情信息接口

GET /agent/getwarning

该接口可以获取指定预警的预警详情信息

> Body 请求参数

```yaml
{}

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|

> 返回示例

> 200 Response

```json
{
  "data": [
    {
      "id": 1,
      "host_name": "web-server-01",
      "warning_type": "CPU与内存",
      "warning_title": "主机 web-server-01 发生 CPU与内存，CPU使用率: 85.00%，内存使用率: 92.00%",
      "warning_time": "2025-06-12T10:30:00+08:00"
    },
    {
      "id": 2,
      "host_name": "web-server-01",
      "warning_type": "CPU",
      "warning_title": "主机 web-server-01 发生 CPU，CPU使用率: 88.00%，内存使用率: 75.00%",
      "warning_time": "2025-06-12T10:45:00+08:00"
    }
  ]
}
```

> 400 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

# 登陆注册

## POST 个人注册/agent/register

POST /agent/register

用户注册，需输入用户名、密码、邮箱，系统会进行校验（输入是否非空，格式是否正确，用户名或邮箱是否已经存在）

> Body 请求参数

```json
{
  "name": "user100",
  "password": "user100pwd",
  "email": "0@qq.com",
  "company": "szu"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 否 |none|
|» name|body|string| 是 |none|
|» password|body|string| 是 |none|
|» email|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "注册成功"
}
```

> 400 Response

```json
{
  "message": "请求数据格式错误"
}
```

```json
{
  "message": "用户名不能为空"
}
```

```json
{
  "message": "邮箱不能为空"
}
```

```json
{
  "message": "邮箱格式不正确"
}
```

```json
{
  "message": "密码长度应该不小于6，不大于16"
}
```

```json
{
  "message": "用户名已存在"
}
```

```json
{
  "message": "邮箱已存在"
}
```

```json
{
  "mmessage": "数据库查询用户名失败"
}
```

```json
{
  "message": "数据库查询邮箱失败"
}
```

```json
{
  "message": "用户创建失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|422|[Unprocessable Entity](https://tools.ietf.org/html/rfc2518#section-10.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **422**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **500**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

## POST 登录/agent/login

POST /agent/login

用户登录，需要输入用户名和密码，系统会进行校验（输入是否非空，格式是否正确，用户名与密码是否匹配）

> Body 请求参数

```json
{
  "name": "czh",
  "password": "123456"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 否 |none|
|» name|body|string| 是 |none|
|» password|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "登录成功",
  "role": "USER",
  "// \"ADMIN\"或\"ROOT\" \"username\"": "czh",
  "companyId": 2,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImN6aCIsImV4cCI6MTc0MTMxNDc5NX0.CmwBSimhVB5cJ7V10HV_Do1bncLzXgZt3ikGbLmq3G0"
}
```

> 400 Response

```json
{
  "message": "请求数据格式错误"
}
```

```json
{
  "message": "用户不存在"
}
```

```json
{
  "message": "密码错误"
}
```

> 500 Response

```json
{
  "message": "生成 token 错误"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|
|» role|string|true|none||none|
|» username|string|true|none||none|
|» companyId|integer|true|none||none|
|» token|string|true|none||none|

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **401**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **500**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

## POST 发送注册时的验证码邮件

POST /registertoken

注册时请求发送含6位验证码token的邮件

> Body 请求参数

```json
{
  "email": "1737370258@qq.com"
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 否 |none|
|» email|body|string| 是 |none|

> 返回示例

> 200 Response

```json
"// 响应邮件，无json"
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

# 密码找回

## POST 重置密码请求/agent/request_reset_password

POST /agent/request_reset_password

该函数用于处理用户请求重置密码的请求。当用户提交包含其电子邮件地址的请求时，该函数会验证请求数据的有效性，查找对应的用户，生成一个唯一的重置密码token，将该token保存到数据库中，并向用户发送包含重置密码链接的电子邮件。

参数说明：
Header中带一个Authorization，
参数：
{
    "email":""
}

处理流程：
绑定JSON请求体到结构体request中。
根据电子邮件地址查找用户。
如果用户不存在，返回404状态码和相应消息。
如果用户存在，生成一个唯一的重置密码token。
将token保存到数据库中。
调用sendResetPasswordEmail函数发送重置密码邮件。
返回200状态码和成功消息。

> Body 请求参数

```json
{
  "email": ""
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 否 |none|
|» email|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "重置密码请求成功"
}
```

> 400 Response

```json
{
  "message": "请求数据格式错误"
}
```

> 404 Response

```json
{
  "message": "用户未找到"
}
```

```json
{
  "message": "数据库查询失败"
}
```

```json
{
  "message": "保存 token 失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## POST 重置密码/reset_password

POST /reset_password

该函数用于处理用户提交的重置密码请求。当用户点击邮件中的重置密码链接并输入新密码时，该函数会验证请求数据的有效性，检查token的有效性，更新用户的密码，并重置token。

参数说明：
{
    “token”:"" // 申请密码重置时生成的token
    "new_password":""
}

处理流程：
绑定JSON请求体到结构体request中。
根据token查找用户。
如果token无效或用户不存在，返回404状态码和相应消息。
如果用户存在且token有效，更新用户的密码。
重置用户的token（将其设置为null或清空）。
返回200状态码和成功消息。

> Body 请求参数

```json
{
  "token": "",
  "new_password": ""
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|body|body|object| 否 |none|
|» token|body|string| 是 |none|
|» new_password|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "重置密码成功"
}
```

> 400 Response

```json
{
  "message": "请求数据格式错误"
}
```

> 404 Response

```json
{
  "message": "无效的重置密码 token"
}
```

```json
{
  "message": "数据库查询失败"
}
```

```json
{
  "message": "密码重置失败"
}
```

```json
{
  "message": "密码重置成功，但是 token 重置失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

# 用户信息

## GET 获取当前用户信息/agent/userInfo

GET /agent/userInfo

获得当前登录用户的信息

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "message": "获取用户信息成功",
  "user": {
    "id": 1,
    "name": "root",
    "realname": "",
    "// 默认为空，当注册公司时才会提示进行更新 \"email\"": "root@example.com",
    "password": "123456",
    "company_id": 1,
    "role_id": 1,
    "is_verified": false
  }
}
```

> 401 Response

```json
{
  "message": "用户不存在"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|
|» user|object|true|none||none|
|»» id|integer|true|none||none|
|»» name|string|true|none||none|
|»» realname|string|true|none||默认为空，当注册公司时才会提示进行更新|
|»» email|string|true|none||none|
|»» password|string|true|none||none|
|»» company_id|integer|true|none||none|
|»» role_id|integer|true|none||none|
|»» is_verified|boolean|true|none||none|

## POST 更新用户信息/agent/updateUserInfo

POST /agent/updateUserInfo

更新当前登录用户的用户名、密码、邮箱,或者进行实名
参数：
```json
{
	"new_name":"",
	"new_password":"",
	"new_email":"",
        "realname":"",
}
```

> Body 请求参数

```json
{
  "new_name": "",
  "new_password": "",
  "new_email": "",
  "realname": ""
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» new_name|body|string| 是 |none|
|» new_password|body|string| 是 |none|
|» new_email|body|string| 是 |none|
|» realname|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "更新成功"
}
```

> 409 Response

```json
{
  "message": "更新用户名错误：新用户名已存在",
  "error": ""
}
```

```json
{
  "message": "更新用户名失败",
  "error": ""
}
```

```json
{
  "message": "更新密码失败",
  "error": ""
}
```

```json
{
  "message": "更新邮箱失败",
  "error": ""
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|409|[Conflict](https://tools.ietf.org/html/rfc7231#section-6.5.8)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## GET （管理员）获取所有用户的信息/agent/allUserInfo

GET /agent/allUserInfo

管理员获取所有用户的信息（隐藏密码字段的值）

> 返回示例

> 200 Response

```json
{
  "message": "获取用户信息成功",
  "users": [
    {
      "id": 1,
      "name": "root",
      "email": "root@example.com",
      "password": "",
      "role_id": 1,
      "is_verified": false
    },
    {
      "id": 2,
      "name": "user1",
      "email": "user1@qq.com",
      "password": "",
      "role_id": 2,
      "is_verified": false
    },
    {
      "id": 3,
      "name": "user2",
      "email": "user2@qq.com",
      "password": "",
      "role_id": 2,
      "is_verified": false
    },
    {
      "id": 4,
      "name": "user3",
      "email": "user3@qq.com",
      "password": "",
      "role_id": 2,
      "is_verified": false
    },
    {
      "id": 5,
      "name": "user4",
      "email": "user4@qq.com",
      "password": "",
      "role_id": 2,
      "is_verified": false
    },
    {
      "id": 6,
      "name": "user5",
      "email": "user5@qq.com",
      "password": "",
      "role_id": 2,
      "is_verified": false
    }
  ]
}
```

> 401 Response

```json
{
  "message": "用户没有权限"
}
```

> 500 Response

```json
{
  "message": "获取所有用户信息失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

# 系统/公司管理员操作

## GET 查询公司信息（含成员）/agent/get-company-info

GET /agent/get-company-info

# 公司信息管理接口

允许系统管理员指定公司及其成员信息，也允许公司管理员查询其所属公司的详细信息（包括成员信息）。

## 请求方式

- **Method**: `GET`
- **Content-Type**: `application/json`
- **Authorization**: `your_jwt_token`

## 请求参数

可通过以下查询参数来指定要查询的公司。如果不提供该参数，则默认查询当前登录用户（需要相应的管理员权限）所在的公司信息。

| 参数名         | 类型   | 必需 | 描述                                       |
|--------------|------|----|------------------------------------------|
| company-name | 字符串 | 否  | 要查询的公司名称。若不提供，则查询当前用户所属公司的信息。 |

### 示例请求

- 查询特定公司的信息：
    ```
    GET /agent/company-info?company-name=ABC公司
    ```

- 查询当前登录管理员所属公司的信息（无需指定`company-name`参数）：
    ```
    GET /agent/company-info
    ```

## 注意事项

- 确保调用此接口时，用户具有足够的权限（系统管理员或公司管理员权限），以便成功执行对应的操作。
- 当使用`company-name`作为查询参数时，请确保提供的公司名称准确无误，以避免查询不到预期的结果。

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "data": {
    "Company": {
      "id": 1,
      "name": "Company",
      "admin": "root",
      "member_num": 5,
      "system_num": 5,
      "description": "测试用公司"
    },
    "Members": [
      {
        "username": "user1",
        "email": "user1@qq.com",
        "role_id": 1
      },
      {
        "username": "user2",
        "email": "user2@qq.com",
        "role_id": 0
      },
      {
        "username": "user3",
        "email": "user3@qq.com",
        "role_id": 0
      },
      {
        "username": "user4",
        "email": "user4@qq.com",
        "role_id": 0
      },
      {
        "username": "user5",
        "email": "user5@qq.com",
        "role_id": 0
      }
    ]
  },
  "message": "获取公司信息成功"
}
```

> 401 Response

```json
{
  "message": "公司不存在"
}
```

```json
{
  "message": "非管理员，权限不足"
}
```

```json
{
  "message": "非系统管理员或指定公司的管理员，权限不足"
}
```

```json
{
  "message": "数据库查询管理员失败"
}
```

```json
{
  "message": "数据库查询公司失败"
}
```

```json
{
  "message": "数据库查询公司成员失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## GET 获取公司列表/agent/get-company-list

GET /agent/get-company-list

# 获取所有公司信息接口

允许系统管理员获取平台中所有公司的基本信息（不包含成员信息）。

## 请求方式

- **Method**: `GET`
- **Content-Type**: `application/json`
- **Authorization**: `your_jwt_token`

## 请求URL
```
GET /agent/get-company-list
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "data": [
    {
      "id": 1,
      "name": "Company",
      "description": "测试用公司",
      "admin_id": 2,
      "adminName": "user1",
      "adminEmail": "user1@qq.com",
      "membernum": 5,
      "systemnum": 5
    },
    {
      "id": 3,
      "name": "向阳开",
      "description": "",
      "admin_id": 1,
      "adminName": "root",
      "adminEmail": "root@example.com",
      "membernum": 1,
      "systemnum": 3
    }
  ],
  "message": "查询成功"
}
```

> 403 Response

```json
{
  "message": "非系统管理员，权限不足"
}
```

> 500 Response

```json
{
  "message": "数据库查询公司失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## GET 查询当前公司所有成员信息/agent/getmemberinfo

GET /agent/getmemberinfo

# 获取公司成员信息接口

允许管理员根据公司名称查询特定公司的所有成员信息。如果不提供公司名称，则默认查询当前登录管理员所属公司的成员信息。

## 请求方式

- **Method**: `GET`
- **Content-Type**: `application/json`
- **Authorization**: `your_jwt_token`

## 请求URL
```
GET /agent/getmemberinfo
```
## 请求参数

| 参数名         | 类型   | 必需 | 描述                                       |
|--------------|------|----|------------------------------------------|
| company-name | 字符串 | 否  | 要查询成员信息的公司名称。若不提供，则查询当前登录管理员所属公司的成员信息。 |

### 示例请求

- 查询特定公司的成员信息：
    ```
    GET /api/company-members?company-name=ABC公司
    ```

- 查询当前登录管理员所属公司的成员信息（无需指定`company-name`参数）：
    ```
    GET /api/company-members
    ```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|company|query|string| 否 |none|
|Authorization|header|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "message": "获取公司成员信息成功",
  "members": [
    {
      "username": "user1",
      "email": "user1@qq.com",
      "role_id": 1
    },
    {
      "username": "user2",
      "email": "user2@qq.com",
      "role_id": 0
    },
    {
      "username": "user3",
      "email": "user3@qq.com",
      "role_id": 0
    },
    {
      "username": "user4",
      "email": "user4@qq.com",
      "role_id": 0
    },
    {
      "username": "user5",
      "email": "user5@qq.com",
      "role_id": 0
    }
  ]
}
```

> 400 Response

```json
{
  "errors": "The company is not exist"
}
```

> 401 Response

```json
{
  "errors": "The user is unauthorized, the token is wrong "
}
```

> 403 Response

```json
{
  "errors": "The user does not have access permission"
}
```

> 500 Response

```json
{
  "errors": "failed to query the company"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» username|string|true|none||none|
|» email|string|true|none||none|

## POST 添加成员/agent/addMember

POST /agent/addMember

公司管理员添加成员（会判断当前用户是否为公司管理员）添加后会向指定成员发送信息
参数：
{
    "username":"",
    "email":""
}

> Body 请求参数

```json
{
  "username": "",
  "email": ""
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» username|body|string| 是 |none|
|» email|body|string| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "管理员[用户名]成功添加成员[用户名]"
}
```

> 400 Response

```json
{
  "message": "请求数据格式错误"
}
```

> 403 Response

```json
{
  "message": "非公司管理员，权限不足"
}
```

```json
{
  "message": "该用户已属于当前公司"
}
```

```json
{
  "message": "该用户已属于其他某个公司"
}
```

```json
{
  "message": "邮箱已存在"
}
```

```json
{
  "message": "数据库查询管理员失败"
}
```

```json
{
  "message": "数据库查询用户名失败"
}
```

```json
{
  "message": "数据库查询邮箱失败"
}
```

```json
{
  "message": "用户创建失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|422|[Unprocessable Entity](https://tools.ietf.org/html/rfc2518#section-10.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## POST 管理员批量删除账号/agent/deleteMembers

POST /agent/deleteMembers

管理员批量删除账号
参数：
{
    "usernames": ["Alice", "Bob", "Charlie"]
}

> Body 请求参数

```json
{
  "usernames": [
    "Alice",
    "Bob",
    "Charlie"
  ]
}
```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |none|
|body|body|object| 否 |none|
|» usernames|body|[string]| 是 |none|

> 返回示例

> 200 Response

```json
{
  "message": "管理员[用户名]成功删除成员"
}
```

> 400 Response

```json
{
  "message": "请求数据格式错误"
}
```

> 403 Response

```json
{
  "message": "非公司管理员，权限不足"
}
```

```json
{
  "message": "数据库查询管理员失败"
}
```

```json
{
  "message": "事务开始失败"
}
```

```json
{
  "message": "数据库查询用户[用户名]失败；已取消当前所有用户删除操作"
}
```

```json
{
  "message": "部分用户删除失败；已取消当前所有用户删除操作"
}
```

```json
{
  "message": "事务提交失败，取消当前所有用户删除操作"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## POST 添加sshkey

POST /agent/sshkey

通过该接口添加sshkey，用于远程登录连接

> Body 请求参数

```
hostname: user1
sshkey: "123456"

```

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|hostname|query|string| 否 |主机名|
|sshkey|query|string| 否 |ssh私钥|
|body|body|string| 否 |none|

> 返回示例

> 200 Response

```json
{
  "message": "SSH密钥添加成功"
}
```

> 400 Response

```json
{
  "error": "请求头中缺少Token"
}
```

> 401 Response

```json
{
  "message": "未登录"
}
```

> 403 Response

```json
{
  "message": "非系统/公司管理员，权限不足"
}
```

> 500 Response

```json
{
  "message": "数据库查询失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» error|string|true|none||none|

## GET 获取指定用户接收到的消息

GET /agent/info/recivelist

获取该用户作为接收者收到的所有信息

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |token用于验证用户名|

> 返回示例

> 200 Response

```json
{
  "message": "获取接收到的通知成功",
  "receiveNotices": [
    {
      "id": 1,
      "send": "张三",
      "receive": "root",
      "content": "邀请你加入公司",
      "state": "unprocessed",
      "create_at": "2025-04-24T10:55:40.299042Z"
    },
    {
      "id": 2,
      "send": "李四",
      "receive": "root",
      "content": "邀请你加入公司",
      "state": "processed",
      "create_at": "2025-04-24T10:55:40.299042Z"
    },
    {
      "id": 3,
      "send": "王五",
      "receive": "root",
      "content": "邀请你加入公司",
      "state": "expired",
      "create_at": "2025-04-24T10:55:40.299042Z"
    }
  ]
}
```

> 401 Response

```json
{
  "message": "用户不存在"
}
```

> 500 Response

```json
{
  "error": "消息查询失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|
|» receiveNotices|[object]|true|none||none|
|»» id|integer|true|none||none|
|»» send|string|true|none||none|
|»» receive|string|true|none||none|
|»» content|string|true|none||none|
|»» state|any|true|none||none|

*anyOf*

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»»» *anonymous*|string|false|none||none|

*or*

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»»» *anonymous*|string|false|none||none|

*continued*

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|»» create_at|string|true|none||none|

状态码 **401**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|
|» 01JNMN5QRHAC7YVBQZM8D88V4A|string|true|none||none|

## GET 获取指定用户发送的消息

GET /agent/info/sendlist

获取该用户作为接收者发送的所有信息

### 请求参数

|名称|位置|类型|必选|说明|
|---|---|---|---|---|
|Authorization|header|string| 否 |token用于验证用户名|

> 返回示例

> 200 Response

```json
{
  "message": "获取发送的通知成功",
  "sendNotices": [
    {
      "id": 1,
      "receive": "张三",
      "send": "user1",
      "content": "邀请你加入公司",
      "state": "unprocessed",
      "create_at": "2025-04-24T10:55:40.299042Z"
    },
    {
      "id": 2,
      "receive": "李四",
      "send": "user1",
      "content": "邀请你加入公司",
      "state": "processed",
      "create_at": "2025-04-24T10:55:40.299042Z"
    },
    {
      "id": 3,
      "receive": "王五",
      "send": "user1",
      "content": "邀请你加入公司",
      "state": "expired",
      "create_at": "2025-04-24T10:55:40.299042Z"
    }
  ]
}
```

> 401 Response

```json
{
  "error": "token识别错误"
}
```

> 500 Response

```json
{
  "error": "查询信息错误"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» receive|string|true|none|接收者|none|
|» content|string|true|none|内容|none|
|» processed|boolean|true|none|是否已处理|none|
|» create_at|string|true|none|创建时间|none|

## POST 消息处理接口

POST /agent/info/manage

同意/拒绝指定消息内容接口，通过该接口可以同意/拒绝加入公司的消息邀请或者创建公司的申请，先校验审批人是否有权限，审批后修改数据库对应的值

> Body 请求参数

```json
"// id必须传，其他可选\r\n{\r\n    \"id\": 1,\r\n    \"receive\": \"\",\r\n    \"send\": \"\",\r\n    \"content\": \"\",\r\n    \"create_at\": \"\"\r\n}\r\n"
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|mode|query|string| 否 ||同意消息或者拒绝消息|
|Authorization|header|string| 否 ||token用于校验用户|
|body|body|object| 否 ||none|
|» receive|body|string| 是 | 接受者|none|
|» send|body|string| 是 | 发送者|none|
|» content|body|string| 是 | 内容|none|
|» create_at|body|string| 是 | 时间|none|

> 返回示例

> 200 Response

```json
{
  "message": "消息处理成功"
}
```

> 400 Response

```json
{
  "message": "消息已过期"
}
```

> 401 Response

```json
{
  "error": "缺少token"
}
```

> 403 Response

```json
{
  "error": "审批人不具有权限"
}
```

> 500 Response

```json
{
  "error": "数据库处理错误"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|403|[Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## GET 获取日志接口

GET /agent/log

获取日志信息

> Body 请求参数

```json
{
  "from": "2023-01-01T00:00:00Z",
  "to": "2023-02-01T00:00:00Z",
  "type": "添加服务器",
  "result": "成功:添加成功",
  "server_name": "服务器9"
}
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|
|body|body|object| 否 ||none|
|» from|body|string| 是 | 起始时间|none|
|» to|body|string| 是 | 终止时间|none|
|» type|body|string| 是 | 操作类型|none|
|» result|body|string| 是 | 操作结果|none|
|» server_name|body|string| 是 | 相关服务器|none|

> 返回示例

> 200 Response

```json
{
  "": "string"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» *anonymous*|string|false|none||标签|

# 文件传输

## POST 两服务器间单文件传输

POST /agent/transfer

## 请求方法

`POST`

## 请求头

- `Content-Type: application/json`
- `Authorization: Bearer <JWT_TOKEN>` (如果使用JWT进行身份验证)

## 请求体

| 参数名          | 必选 | 类型   | 描述                                       |
|-----------------|------|--------|--------------------------------------------|
| source_server    | 是   | string | 源服务器IP地址                             |
| source_user      | 是   | string | 源服务器SSH登录用户名                      |
| source_auth      | 是   | string | 源服务器SSH认证信息（例如密码或私钥）      |
| source_path      | 是   | string | 源文件路径                                 |
| target_server    | 是   | string | 目标服务器IP地址                           |
| target_user      | 是   | string | 目标服务器SSH登录用户名                    |
| target_auth      | 是   | string | 目标服务器SSH认证信息（例如密码或私钥）    |
| target_path      | 是   | string | 目标文件路径                               |

### 示例请求体

```json
{
  "source_server": "192.168.1.1",
  "source_user": "user1",
  "source_auth": "password1",
  "source_path": "/path/to/source/file.txt",
  "target_server": "192.168.1.2",
  "target_user": "user2",
  "target_auth": "password2",
  "target_path": "/path/to/target/file.txt"
}

> Body 请求参数

```json
{
  "source_server": "192.168.1.1",
  "source_user": "user1",
  "source_auth": "password1",
  "source_path": "/path/to/source/file.txt",
  "target_server": "192.168.1.2",
  "target_user": "user2",
  "target_auth": "password2",
  "target_path": "/path/to/target/file.txt"
}
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|
|body|body|object| 否 ||none|
|» source_server|body|string| 是 ||none|
|» source_user|body|string| 是 ||none|
|» source_auth|body|string| 是 ||none|
|» source_path|body|string| 是 ||none|
|» target_server|body|string| 是 ||none|
|» target_user|body|string| 是 ||none|
|» target_auth|body|string| 是 ||none|
|» target_path|body|string| 是 ||none|

> 返回示例

> 200 Response

```json
{}
```

```json
{
  "message": "解析请求有误"
}
```

```json
{
  "message": "该源服务器不属于用户（所在公司）"
}
```

```json
{
  "message": "该目标服务器不属于用户（所在公司）"
}
```

```json
{
  "message": "创建与源服务器的连接失败"
}
```

```json
{
  "message": "创建与目标服务器的连接失败"
}
```

```json
{
  "message": "查询源服务器是否属于用户（所在公司）失败"
}
```

```json
{
  "message": "查询目标服务器是否属于用户（所在公司）失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## POST 文件下载

POST /agent/download

## 接口地址
```
POST /agent/download
```

## 🔹 请求参数（JSON Body）

| 字段名   | 类型   | 必填 | 描述               |
|----------|--------|------|--------------------|
| `server` | string | 是   | 远程服务器 IP 地址 |
| `path`   | string | 是   | 远程文件路径       |
| `user`   | string | 是  | 登录远程服务器所需的的用户名 |
| `auth`   | string  |是   | 登录远程服务器所需的认证信息（如密码或密钥） |

> Body 请求参数

```json
{
  "server": "192.168.202.128",
  "path": "/home/czh/勇往直前1.ogg",
  "user": "czh",
  "auth": "password"
}
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|
|body|body|object| 否 ||none|
|» server|body|string| 是 ||none|
|» path|body|string| 是 ||none|
|» user|body|string| 是 ||none|
|» auth|body|string| 是 ||none|

> 返回示例

> 200 Response

```json
{}
```

```json
{
  "message": "路径是一个目录: <nil>"
}
```

```json
{
  "message": "解析请求失败: [error]"
}
```

```json
{
  "message": "该服务器不属于用户（所在公司）"
}
```

```json
{
  "message": "文件不存在"
}
```

> 401 Response

```json
{
  "message": "未登录"
}
```

```json
{
  "message": "查询服务器与用户（所在公司）的关系失败"
}
```

```json
{
  "message": "获取文件信息失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

## POST 文件上传

POST /agent/upload

客户端上传文件到指定的服务器

特殊的：
Content-Type: multipart/form-data
## 接口地址
```
POST /agent/download
```

## 🔹 请求参数（form-data）

| 字段名   | 类型   | 必填 | 描述               |
|----------|--------|------|--------------------|
| `file`     | file       |  是 | 本地要上传的文件  |
| `server` | string | 是   | 远程服务器 IP 地址 |
| `path`   | string | 是   | 远程文件路径       |
| `user`   | string | 是   | 登录远程服务器所需的的用户名 |
| `auth`   | string | 是   | 登录远程服务器所需的认证信息（如密码或密钥） |

> Body 请求参数

```yaml
file: ""
server: 192.168.202.128
path: /home/czh/勇往直前1.ogg
user: czh
auth: password

```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|
|body|body|object| 否 ||none|
|» file|body|string(binary)| 否 ||none|
|» server|body|string| 否 ||none|
|» path|body|string| 否 ||none|
|» user|body|string| 否 ||none|
|» auth|body|string| 否 ||none|

> 返回示例

> 200 Response

```json
{
  "message": "文件上传完成"
}
```

```json
{
  "message": "解析请求失败"
}
```

```json
{
  "message": "该服务器不属于用户（所在公司）"
}
```

```json
{
  "message": "文件上传失败"
}
```

```json
{
  "message": "获取要上传的文件失败"
}
```

> 401 Response

```json
{
  "message": "未登录"
}
```

> 500 Response

```json
{
  "message": "查询服务器是否属于用户（所在公司）失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

# 日志

## POST 获取用户操作日志

POST /agent/getuseroperationlogs

获取用户的操作日志，支持按时间段、用户名、操作类型进行筛选。
POST

用户个人操作日志查询：
可以不用传username
可选参数fromTime、toTime、operation用于筛选

系统管理员查询所有操作日志：
所有参数都可选，username、fromTime、toTime、operation均用于筛选

> Body 请求参数

```json
"{\r\n    \"username\":\"\", // 用户查询个人操作日志时不需要，（系统管理员）查询所有操作日志时可选，root\r\n    \"fromTime\":\"2025-05-29T21:11:48.576+0800\", // 起始时间，可选\r\n    \"toTime\":\"2028-05-29T23:11:48.576+0800\", // 结束时间，可选\r\n    \"operation\":\"\" // 操作类型。可选,文件上传、文件下载、两服务器间单文件传输、添加服务器等\r\n}"
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|
|body|body|object| 否 ||none|
|» username|body|string| 否 ||用户查询个人操作日志时不需要，（系统管理员）查询所有操作日志时可选|
|» fromTime|body|string| 否 ||起始时间，可选|
|» toTime|body|string| 否 ||结束时间，可选|
|» operation|body|string| 否 ||操作类型。可选|

> 返回示例

> 200 Response

```json
{
  "logs": [
    {
      "level": "error",
      "ts": "2025-05-29T21:00:50.305+0800",
      "msg": "文件下载",
      "username": "czh",
      "detail": "创建与目标服务器的连接失败，请检查目标服务器是否正确"
    },
    {
      "level": "error",
      "ts": "2025-05-29T21:00:56.041+0800",
      "msg": "两服务器间单文件传输",
      "username": "czh",
      "detail": "创建与源服务器的连接失败，请检查源服务器是否正确"
    },
    {
      "level": "error",
      "ts": "2025-05-29T21:10:35.019+0800",
      "msg": "文件上传",
      "username": "czh",
      "detail": "创建与目标服务器的连接失败，请检查目标服务器是否正确"
    }
  ]
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

状态码 **200**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» logs|[object]|true|none||操作日记集|
|»» level|string|true|none||日志级别|
|»» ts|string|true|none||时间戳|
|»» msg|string|true|none||操作类型|
|»» username|string|true|none||用户名|
|»» detail|string|true|none||详情|

# 脚本

## GET 获取安装代理程序的脚本

GET /agent/agentscript

获取安装代理程序的脚本
GET
需要Authorization
/agentscript （无需Authorization）供用户在要被监控的服务器上通过curl或wget命令进行agent脚本的获取，例如：
curl http://x.x.x.x:8080/agentscript?hostname=czh-centos -o install.sh

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 获取反向ssh配置的脚本

GET /agent/sshscript

获取反向ssh配置的脚本
需要Authorization

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|

> 返回示例

> 200 Response

```json
{}
```

> 503 Response

```json
{
  "message": "获取反向ssh配置的脚本失败：没有可用端口号了"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|503|[Service Unavailable](https://tools.ietf.org/html/rfc7231#section-6.6.4)|none|Inline|

### 返回数据结构

## GET 获取包含agent和反向ssh配置的脚本

GET /agent/combinedscript

获取包含agent和反向ssh配置的脚本
GET
需要Authorization

/combinedscript （无需Authorization）供用户在要被监控的服务器上通过curl或wget命令进行agent脚本的获取，例如：
curl http://x.x.x.x:8080/combinedscript?hostname=czh-centos -o install.sh

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|hostname|query|string| 否 ||none|
|Authorization|header|string| 否 ||none|

> 返回示例

> 200 Response

```json
{}
```

> 500 Response

```json
{
  "error": "template parse error"
}
```

> 503 Response

```json
{
  "message": "获取反向ssh配置的脚本失败：没有可用端口号了"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|
|503|[Service Unavailable](https://tools.ietf.org/html/rfc7231#section-6.6.4)|none|Inline|

### 返回数据结构

## GET 获取删除代理服务和反向ssh服务的脚本

GET /uninstallcombinedscript

删除代理服务的脚本
无需任何参数
curl http://x.x.x.x:8080/uninstallcombinedscript -o uninstall.sh

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 获取删除代理服务的脚本

GET /uninstallagentscript

无需任何参数

curl http://x.x.x.x:8080/uninstallagentscript -o uninstall.sh

> 返回示例

> 200 Response

```json
{}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### 返回数据结构

## GET 获取生成配置反向ssh的脚本所需的端口号

GET /agent/port/get

获取生成反向ssh配置脚本所需的端口号，以确保所用端口号在用户要监控的服务器上可用，然后动态生成配置反向ssh的脚本
GET
需要Authorization

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|Authorization|header|string| 否 ||none|

> 返回示例

> 200 Response

```json
{
  "message": "端口号获取成功",
  "port": 10000
}
```

> 503 Response

```json
{
  "message": "端口号获取失败：没有可用端口号了",
  "port": -1
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|503|[Service Unavailable](https://tools.ietf.org/html/rfc7231#section-6.6.4)|none|Inline|

### 返回数据结构

# 公司业务

## POST 注册团队申请接口

POST /agent/registercompany

邀请团队接口，接收公司名称、统一社会信用代码、法人姓名、公司管理员姓名和公司管理员邮箱注册，向系统管理员提交邀请团队申请。

> Body 请求参数

```json
"{\r\n    \"company\": \"软件工程\",\r\n    \"social_credit_code\": \"91310115MA1K3YJ123\", // 18位\r\n    \"legal_name\": \"陈xx\",\r\n    \"admin_name\": \"陈xx\",\r\n    \"admin_email\": \"xxx@qq.com\"\r\n}"
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|body|body|object| 否 ||none|
|» company|body|string| 是 | 公司名|none|
|» social_credit_code|body|string| 是 | 社会信用代码|none|
|» legal_name|body|string| 是 | 法人姓名|none|
|» admin_name|body|string| 是 | 公司管理员姓名|none|
|» admin_email|body|string| 是 | 公司管理员邮箱|none|

> 返回示例

> 200 Response

```json
{
  "message": "发出团队申请"
}
```

> 400 Response

```json
{
  "message": "请求头中缺少Token"
}
```

> 401 Response

```json
{
  "message": "未找到用户信息"
}
```

> 500 Response

```json
{
  "message": "数据库查询公司失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **400**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» error|string|true|none||none|

## POST 邀请加入公司接口

POST /agent/joincompany

公司管理员通过该接口传输要邀请成员姓名、用户名和邮箱从而邀请该成员加入公司

> Body 请求参数

```json
{
  "realname": "user6",
  "username": "user6",
  "email": "user6@qq.com"
}
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|body|body|object| 否 ||none|
|» realname|body|string| 是 ||none|
|» username|body|string| 是 ||none|
|» email|body|string| 是 ||none|

> 返回示例

> 200 Response

```json
{
  "message": "发出加入公司邀请"
}
```

> 400 Response

```json
{
  "message": "公司不存在"
}
```

> 401 Response

```json
{
  "message": "用户不存在"
}
```

> 500 Response

```json
{
  "message": "数据库查询管理员失败"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

状态码 **401**

|名称|类型|必选|约束|中文名|说明|
|---|---|---|---|---|---|
|» message|string|true|none||none|
|» 01JNMN5QRHAC7YVBQZM8D88V4A|string|true|none||none|

## POST 更换管理员

POST /agent/replaceadmin

公司管理员提交新管理员姓名、新管理员用户名和新管理员邮箱，申请更换公司管理员。

> Body 请求参数

```json
{
  "realname": "user2",
  "username": "user2",
  "email": "user2@qq.com"
}
```

### 请求参数

|名称|位置|类型|必选|中文名|说明|
|---|---|---|---|---|---|
|body|body|object| 否 ||none|
|» realname|body|string| 是 ||none|
|» username|body|string| 是 ||none|
|» email|body|string| 是 ||none|

> 返回示例

> 200 Response

```json
{
  "message": "更换管理员申请提交成功"
}
```

> 400 Response

```json
{
  "message": "请求格式有误"
}
```

> 401 Response

```json
{
  "message": "没有权限"
}
```

> 500 Response

```json
{
  "message": "服务器出现错误"
}
```

### 返回结果

|状态码|状态码含义|说明|数据模型|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|
|401|[Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)|none|Inline|
|500|[Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1)|none|Inline|

### 返回数据结构

# 数据模型

