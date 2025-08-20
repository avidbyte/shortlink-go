# ShortLink 🚀

一个轻量级的短链接生成与管理平台，支持短链创建、访问统计、状态管理等功能。  
项目采用 **前后端分离架构**：

- **后端**：Go 语言实现
- **前端**：Vue 框架 前端地址 👉 [shortlink-web](https://github.com/avidbyte/shortlink-web)

---

## 🌐 功能特性

- 🔗 短链生成与管理
- 🔄 支持 301/302/307 重定向
- 📊 访问统计（PV/UV）
- ⚙️ 分页查询、禁用/启用状态管理
- ⚡ Redis 缓存加速访问
- 🌍 多语言国际化（i18n）
- 📝 完整的日志记录与错误处理

---

## 🛠️ 运行环境

运行本项目需要以下依赖：

- **Go 1.18+**
- **MySQL 5.7+ / 8.0+** （存储短链及统计数据）
- **Redis 6.0+** （缓存加速与 UV 统计）

请确保本地或服务器已正确安装并启动以上服务。

## 📦 技术栈与依赖

| 库名     | 用途           |
|--------|----------------|
| Gin    | Web 框架       |
| GORM   | ORM 数据库操作 |
| Redis  | 缓存支持       |
| Zap    | 高性能日志库   |
| i18n   | 国际化支持     |
| Cron   | 定时任务       |
| Docker | 容器化部署支持 |

---

## 🚀 快速部署

### 1. 配置环境变量


```shell
# 生产环境配置文件路径
export SHORTLINK_CONFIG_PATH=/shortlink/config.yaml

# 生产环境 i18n 文件路径
export SHORTLINK_I18N_PATH=/shortlink/i18n
```

### 2.启动服务

```shell
./shortlink-app
```

- 前端访问地址：http://localhost:80

- API 访问地址：http://localhost:8080

## 🧪 开发与构建
安装依赖

刷新并整理依赖：

```shell
go mod tidy
```
常用依赖（如需手动安装）：
```shell
go get gopkg.in/natefinch/lumberjack.v2@v2.2.1   # 日志
go get github.com/robfig/cron/v3                 # 定时任务
go get -u github.com/gin-contrib/cors            # CORS 支持
go get gopkg.in/toml.v1@v1.9.5                   # 国际化
```
构建命令

Linux
```shell
GOOS=linux GOARCH=amd64 go build -o shortlink-app
```

macOS
- Intel 芯片 (x86_64)
```shell
GOOS=darwin GOARCH=amd64 go build -o shortlink-app
```

- Apple Silicon (M1/M2/M3)
```shell
GOOS=darwin GOARCH=arm64 go build -o shortlink-app
```

Windows PowerShell：
```shell
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o shortlink-app
```

## 📄 文档与资源
示例 Nginx 配置：docs/nginx-example.conf

依赖树查看：
```shell
go mod graph
```

## 📝 License

MIT License

📬 联系方式

如有问题或建议，请提交 Issue 或联系项目维护者。