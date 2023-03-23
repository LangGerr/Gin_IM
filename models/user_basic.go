package models

import (
	"GinChat/utils"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// 通过在后面加上valid 可以实现验证是否符合规则
type UserBasic struct {
	gorm.Model
	Name          string
	Password      string `valid:"required"`
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Identity      string
	ClientIp      string
	ClientPort    string
	Salt          string
	LoginTime     time.Time
	HeartbeatTime time.Time
	LoginOutTime  time.Time
	IsLogOut      bool
	DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func FindUserByName(name string) UserBasic {
	var user UserBasic
	utils.DB.Where("name=?", name).First(&user)
	return user
}

// 查找某个用户
func FindByID(id uint) UserBasic {
	user := UserBasic{}
	utils.DB.Where("id=?", id).First(&user)
	return user
}

func FindUserByPhone(phone string) UserBasic {
	var user UserBasic
	utils.DB.Where("phone=?", phone).First(&user)
	return user
}

func FindUserByEmail(email string) UserBasic {
	var user UserBasic
	utils.DB.Where("email=?", email).First(&user)
	return user
}

func GetUserList() []*UserBasic {
	data := make([]*UserBasic, 10)
	utils.DB.Find(&data)
	return data
}

func CreateUser(user UserBasic) *gorm.DB {
	return utils.DB.Create(&user)
}

func DeleteUser(user UserBasic) *gorm.DB {
	return utils.DB.Delete(&user)
}

func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name, Password: user.Password, Phone: user.Phone, Email: user.Email})
}

func FindUserByNameAndPwd(name, password string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name=? and password=?", name, password).First(&user)
	// 加入token 进行加密 放入identity字段
	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.Md5Encode(str)
	utils.DB.Model(&user).Where("id=?", user.ID).Update("identity", temp)
	return user
}
