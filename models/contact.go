package models

import (
	"GinChat/utils"
	"fmt"
	"gorm.io/gorm"
)

// 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint // 谁的关系
	TargetId uint // 对应的谁
	Type     int  // 对应类型 1好友 2群
	Desc     string
}

func (table *Contact) TableName() string {
	return "Contact"
}

func SearchFriends(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("owner_id=? and type=1", userId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.TargetId)
		fmt.Println(">>>>>>>>>>> ", v)
	}
	friends := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&friends)
	fmt.Println("<<<<<<<<<<<<<< ", friends)
	return friends
}

//func AddFriend(userId uint, targetId uint) (int, string) {
//	user := UserBasic{}
//	if targetId != 0 {
//		user = FindByID(targetId)
//		if user.Salt != "" {
//			if userId == user.ID {
//				return -1, "不能自己加自己为好友！"
//			}
//			contact0 := Contact{}
//			utils.DB.Where("owner_id=? and target_id=? and type=1", userId, targetId).Find(&contact0)
//			if contact0.ID != 0 {
//				return -1, "不能重复添加好友！"
//			}
//			// 事务
//			tx := utils.DB.Begin()
//			// 事务一旦开始，不论什么异常都会rollback
//			defer func() {
//				if r := recover(); r != nil {
//					tx.Rollback()
//				}
//			}()
//			contact := Contact{}
//			contact.OwnerId = userId
//			contact.TargetId = targetId
//			contact.Type = 1
//			if err := utils.DB.Create(&contact).Error; err != nil {
//				tx.Rollback()
//				return -1, "添加好友失败！"
//			}
//			contact1 := Contact{}
//			contact1.OwnerId = targetId
//			contact1.TargetId = userId
//			contact1.Type = 1
//			if err := utils.DB.Create(&contact1).Error; err != nil {
//				tx.Rollback()
//				return -1, "添加好友失败！"
//			}
//			tx.Commit()
//
//			return 0, "添加好友成功！"
//		}
//		return -1, "没有找到此用户！"
//	}
//	return -1, "好友ID不能为空！"
//}

func AddFriend(userId uint, targetName string) (int, string) {
	//user := UserBasic{}

	if targetName != "" {
		targetUser := FindUserByName(targetName)
		//fmt.Println(targetUser, " userId        ", )
		if targetUser.Salt != "" {
			if targetUser.ID == userId {
				return -1, "不能加自己"
			}
			contact0 := Contact{}
			utils.DB.Where("owner_id =?  and target_id =? and type=1", userId, targetUser.ID).Find(&contact0)
			if contact0.ID != 0 {
				return -1, "不能重复添加"
			}
			tx := utils.DB.Begin()
			//事务一旦开始，不论什么异常最终都会 Rollback
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()
			contact := Contact{}
			contact.OwnerId = userId
			contact.TargetId = targetUser.ID
			contact.Type = 1
			if err := utils.DB.Create(&contact).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			contact1 := Contact{}
			contact1.OwnerId = targetUser.ID
			contact1.TargetId = userId
			contact1.Type = 1
			if err := utils.DB.Create(&contact1).Error; err != nil {
				tx.Rollback()
				return -1, "添加好友失败"
			}
			tx.Commit()
			return 0, "添加好友成功"
		}
		return -1, "没有找到此用户"
	}
	return -1, "好友ID不能为空"
}

func SearchUserByGroupId(commuintyId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type = 2", commuintyId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.OwnerId)
	}
	return objIds
}
