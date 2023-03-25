package main

import (
	"GinChat/models"
	"GinChat/utils"
)

func main() {
	utils.InitConfig()
	utils.InitMySQL()
	db := utils.DB
	// 迁移 schema
	db.AutoMigrate(models.Message{})
	// create
	//user := &models.UserBasic{Name: "zhangsan"}
	//db.Create(user)

	// read
	//fmt.Println(db.First(user, 1))
}
