package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortlink-go/internal/handler"
	"shortlink-go/internal/i18n"
	"shortlink-go/internal/middleware"
	"shortlink-go/internal/repository"
	"shortlink-go/internal/service"
	"shortlink-go/pkg/logging"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initConfig() {
	wd, _ := os.Getwd()
	log.Printf("Loading logging from: %s/config.yaml", wd)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
}

func startServer(r *gin.Engine) {
	addr := viper.GetString("server.addr")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 启动服务器
	go func() {
		logging.Logger.Info("Server is running on " + addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.Logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// 关闭 Redis 和 DB （如果支持 Shutdown 方法的话，可以扩展）
	conn := repository.RedisPool.Get()
	// 使用 defer 延迟关闭连接
	defer func() {
		if err := conn.Close(); err != nil {
			logging.Logger.Warn("Redis connection close failed", zap.Error(err))
		}
	}()

	logging.Logger.Info("Server exiting")
}

func main() {

	initConfig()
	// 2. 初始化日志系统
	logging.InitLoggerFromConfig()

	// 3. 使用全局 Logger
	logging.Logger.Info("Application started")

	repository.InitDB(logging.Logger, logging.AtomicLevel)
	repository.InitRedis()

	// 初始化 i18n（加载 TOML 文件）
	bundle, err := i18n.InitI18n([]string{
		"./i18n/en.toml",
		"./i18n/zh.toml",
	}, "en")
	if err != nil {
		panic(err)
	}

	gin.Default()
	r := gin.New()
	r.Use(gin.Recovery()) // 显式添加 Recovery 中间件

	// 注册全局错误中间件
	r.Use(middleware.GlobalErrorMiddleware())
	r.Use(middleware.ZapGinLogger(logging.Logger))
	r.Use(middleware.CorsMiddleware())
	// 使用 i18n 中间件
	r.Use(middleware.I18nMiddleware(bundle))

	api := r.Group("/api")
	{
		api.POST("/shortlink", handler.CreateShortLinkHandler)
		api.GET("/shortlink", handler.ListShortLinksHandler)
		api.PUT("/shortlink/status/:id", handler.UpdateShortLinkStatusHandler)
		api.PUT("/shortlink/:id", handler.UpdateShortLinkHandler)

		api.POST("/whitelist", handler.CreateWhitelistDomainHandler)
		api.GET("/whitelist", handler.ListWhitelistDomainsHandler)
		api.DELETE("/whitelist/:id", handler.DeleteWhitelistDomainHandler)
	}

	// 使用中间件调用 RedirectToTargetURLHandler（避免与 /handler 冲突）
	r.Use(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next() // 只处理 GET 请求
			return
		}
		// 调用处理函数（所有逻辑集中在此）
		handler.RedirectToTargetURLHandler(c)
	})

	c := cron.New()

	// 添加定时任务：每十分钟执行一次
	_, addErr := c.AddFunc("*/10 * * * *", func() {
		if err := service.StatisticalData(); err != nil {
			logging.Logger.Fatal("Failed to generate sitemaps via cron job", zap.Error(err))
		}
	})

	if addErr != nil {
		logging.Logger.Fatal("Failed to schedule cron job", zap.Error(addErr))
	}

	c.Start()

	startServer(r)
}
