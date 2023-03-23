package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

func InitConfig() {
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app inited ")
}

func InitMySQL() (err error) {
	// 处理日志 自定义日志模板，打印sql语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, // 慢sql阈值
			LogLevel:      logger.Info, // 级别
			Colorful:      true,        // 彩色
		})

	DB, err = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic("failed to connect mysql database")
		return err
	}
	fmt.Println("config mysql inited ")

	//DB.AutoMigrate(models.UserBasic{})
	//user := models.UserBasic{}
	//DB.Find(&user)
	//fmt.Println(user)
	return nil
}

func InitRedis() (err error) {
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		Password:     viper.GetString("redis.password"),
		MinIdleConns: viper.GetInt("redis.minIdleConn"),
	})
	//pong, err := Red.Ping().Result()
	//if err != nil {
	//	panic("failed to connect redis database")
	//	return err
	//}
	//fmt.Println("config Redis inited", pong)
	return nil
}

const (
	PublishKey = "websocket"
)

// Publish 发布消息到redis
func Publish(ctx context.Context, channel string, msg string) error {
	var err error
	// ctx是页面请求的
	fmt.Println("Publish......", msg)
	err = Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// subscribe 订阅redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)
	// 订阅之后 接收消息
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Subscribe......", msg.Payload)
	return msg.Payload, err
}
