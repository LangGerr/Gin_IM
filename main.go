package main

import (
	"GinChat/router"
	"GinChat/utils"
	"github.com/spf13/viper"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	utils.InitRedis()
	r := router.Router()
	r.Run(viper.GetString("port.server"))
}
