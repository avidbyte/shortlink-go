package repository

import (
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"shortlink-go/pkg/logging"
	"time"
)

var RedisPool *redis.Pool

func InitRedis() {
	addr := viper.GetString("redis.addr")
	password := viper.GetString("redis.password")

	RedisPool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", addr)
			if err != nil {
				logging.Logger.Error("Failed to connect Redis",
					zap.String("addr", addr),
					zap.Error(err),
				)
				return nil, err
			}

			// 如果设置了密码，执行 AUTH
			if password != "" {
				if _, authErr := conn.Do("AUTH", password); authErr != nil {
					// 记录关闭连接时的错误（如果有的话）
					if closeErr := conn.Close(); closeErr != nil {
						logging.Logger.Error("Failed to close redis connection after AUTH failure",
							zap.String("addr", addr),
							zap.Error(closeErr),
						)
					}
					logging.Logger.Error("Redis AUTH failed",
						zap.String("addr", addr),
						zap.Error(authErr),
					)
					return nil, authErr
				}
			}

			logging.Logger.Error("Redis connection established",
				zap.String("addr", addr),
				zap.Bool("auth", password != ""), // 是否启用认证
			)

			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) > time.Minute {
				_, err := c.Do("PING")
				if err != nil {
					logging.Logger.Warn("Redis connection health check failed",
						zap.String("addr", addr),
						zap.Error(err),
					)
				}
				return err
			}
			return nil
		},
	}
}
