
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


TOML 解析库
```shell
go get github.com/BurntSushi/toml@v1.5.0
```

go-i18n 国际化库（v2 版本）
```shell
go get github.com/nicksnyder/go-i18n/v2@v2.6.0
```


查看依赖树
```shell
go mod graph
```



windows 打包命令
```text
$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o shortlink-app
```

mac 打包命令
```text
GOOS=linux GOARCH=amd64 go build -o shortlink-app
```

