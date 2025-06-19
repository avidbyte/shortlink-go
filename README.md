
刷新依赖并下载,自动添加缺失的依赖，并移除未使用的依赖。
```shell
go mod tidy
```

安装 lumberjack 日志库
```shell
go get gopkg.in/natefinch/lumberjack.v2@v2.2.1

# robfig/cron
go get github.com/robfig/cron/v3
```

```shell
go get -u github.com/gin-contrib/cors
```

windows 打包命令
```text
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o shortlink-app
```

mac 打包命令
```text
GOOS=linux GOARCH=amd64 go build -o shortlink-app
```
