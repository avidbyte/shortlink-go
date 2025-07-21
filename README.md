🌐 项目简介
ShortLink 是一个轻量级的短链接生成与管理平台，支持短链创建、访问统计、禁用/启用状态管理等功能。项目采用前后端分离架构，后端使用
Go 语言实现，前端为 Vue 框架。

🚀 部署方式

1. 服务

```shell
# 生产环境配置文件路径
export SHORTLINK_CONFIG_PATH=/etc/shortlink/config.yaml

# 生产环境 i18n 文件路径
export SHORTLINK_I18N_PATH=/etc/shortlink/i18n

# 启动应用
./shortlink-go
```

前端访问地址：http://localhost:80
API 访问地址：http://localhost:8080

前端项目地址：https://github.com/avidbyte/shortlink-web


🧪 开发与构建

刷新依赖并下载,自动添加缺失的依赖，并移除未使用的依赖。
```shell
go mod tidy
```

安装依赖

```shell
# 安装 lumberjack 日志库
go get gopkg.in/natefinch/lumberjack.v2@v2.2.1

# 安装 cron 定时任务库
go get github.com/robfig/cron/v3

# 安装 gin cors 支持
go get -u github.com/gin-contrib/cors

# 安装 i18n 国际化支持
go get gopkg.in/toml.v1@v1.9.5
```

构建命令


构建 macOS 可执行文件
Intel 芯片（x86_64）
```shell
GOOS=darwin GOARCH=amd64 go build -o shortlink-app
```
Apple Silicon（M1/M2/M3 等 ARM 芯片）

```shell
GOOS=darwin GOARCH=arm64 go build -o shortlink-app
```

构建 Linux 可执行文件

Windows:
```shell
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o shortlink-app
```

macOS/Linux:
```shell
GOOS=linux GOARCH=amd64 go build -o shortlink-app
```

📊 功能特性
短链生成与管理
支持 301/302/307 重定向
访问统计（PV/UV）
分页查询、状态管理
Redis 缓存加速访问
支持国际化（i18n）
完整的日志记录与错误处理

📦 依赖库

|  库名   |  用途   |
|-----|-----|
|  Gin   |  Web 框架   |
|  GORM   |  ORM 数据库操作   |
|  Redis   |  缓存支持   |
|  Zap   |  高性能日志库   |
|   i18n  |   国际化支持  |
|  Cron   |  定时任务   |
| Docker    |  容器化部署支持   |




📄 文档与资源
示例 Nginx 配置：docs/nginx-example.conf
项目构建与依赖说明：README.md
依赖树查看：
```shell
go mod graph
```

📝 License
MIT License

📬 联系方式
如有问题或建议，请提交 Issue 或联系项目维护者。


