package service

import (
	"GinChat/models"
	"GinChat/utils"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// GetUserList
// Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code", "message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := models.GetUserList()
	c.JSON(200, gin.H{"code": 0, "message": data})
}

// FindUserByNameAndPwd
// Summary 验证用户
// @Tags 用户模块
// @param name formData string false "用户名"
// @param password formData string false "用户密码"
// @Success 200 {string} json {"code", "message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	Name := c.Request.FormValue("name")
	Password := c.Request.FormValue("password")
	user := models.FindUserByName(Name)
	if user.ID == 0 {
		c.JSON(200, gin.H{"code": -1, "message": "该用户不存在！"})
		return
	}
	flag := utils.ValidPassword(Password, user.Salt, user.Password)
	if !flag {
		c.JSON(-1, gin.H{"code": -1, "message": "密码不正确！"})
		return
	}
	// 这个不能传入Password 因为这个是123123 而数据库中的是已经加密了的  而且这一步多余了吧，因为上面就已经查询过一边了
	data := models.FindUserByNameAndPwd(Name, utils.MakePassword(Password, user.Salt))
	if data.ID == 0 {
		c.JSON(200, gin.H{"code": -1, "message": "用户名或密码不正确！"})
		return
	}
	c.JSON(200, gin.H{
		"code":    0, // 0 表示成功 -1 表示失败
		"message": "用户登录验证成功！",
		"data":    user,
	})
}

// CreateUser
// Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "用户密码"
// @param repassword query string false "确认密码"
// @param phone query string false "电话号码"
// @param email query string false "电子邮箱"
// @Success 200 {string} json {"code", "message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("Identity")
	//user.Phone = c.Request.FormValue("phone")
	//user.Email = c.Request.FormValue("email")
	salt := fmt.Sprintf("%06d", rand.Int31())
	user.Salt = salt
	data := models.FindUserByName(user.Name)
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{"code": -1, "message": "用户名或密码不能为空！"})
		return
	}
	if data.Name != "" {
		c.JSON(200, gin.H{"code": -1, "message": "用户名已注册！"})
		return
	}
	if password != repassword {
		c.JSON(-1, gin.H{"code": -1, "message": "两次密码不一致！"})
		return
	}
	user.Password = password
	if _, err := govalidator.ValidateStruct(user); err != nil {
		fmt.Println("updateUser err : ", err)
		c.JSON(200, gin.H{"code": -1, "message": "修改用户失败，修改参数不匹配！"})
		return
	}
	user.Password = utils.MakePassword(password, salt)
	//user.Password = password
	models.CreateUser(user)
	c.JSON(200, gin.H{"code": 0, "message": "新增用户成功！", "data": user})
}

// DeleteUser
// Summary 删除用户
// @Tags 用户模块
// @param id query string false "用户id"
// @Success 200 {string} json {"code", "message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	var user models.UserBasic
	utils.DB.Where("id=?", c.Query("id")).First(&user)
	models.DeleteUser(user)
	c.JSON(200, gin.H{"code": 0, "message": "删除用户成功！", "data": user})
}

// UpdateUser
// Summary 修改用户
// @Tags 用户模块
// @param id formData string false "用户id"
// @param name formData string false "用户名"
// @param password formData string false "用户密码"
// @param phone formData string false "电话号码"
// @param email formData string false "电子邮箱"
// @Success 200 {string} json {"code", "message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	if _, err := govalidator.ValidateStruct(user); err != nil {
		fmt.Println("updateUser err : ", err)
		c.JSON(200, gin.H{"code": -1, "message": "修改用户失败，修改参数不匹配！"})
		return
	}
	models.UpdateUser(user)
	c.JSON(200, gin.H{"code": 0, "message": "修改用户成功！", "data": user})
}

// 防止跨域站点伪造请求 upgrader.CheckOrigin. 这将确定是否允许来自不同域的传入请求连接，如果不是，它们将被CORS错误击中。
// 将websocket 升级
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	MsgHandler(ws, c)
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		// 这个subscribe的作用是订阅 utils.PublishKey这个频道，也就是说能够接收到该频道的消息
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println(err)
		}
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		// 这一步应该是将消息返回给客户端
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func SearchFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("userId"))
	friends := models.SearchFriends(uint(userId))
	//c.JSON(200, gin.H{
	//	"code":    0, //0 成功 -1 失败
	//	"message": "查询好友列表成功!",
	//	"data":    friends,
	//})
	utils.RespOkList(c.Writer, friends, len(friends))
}

func AddFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	code, msg := models.AddFriend(uint(userId), targetName)
	if code == 0 {
		utils.RespOk(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// 新建群
func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	name := c.Request.FormValue("name")
	desc := c.Request.FormValue("desc")
	community := models.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Desc = desc
	code, msg := models.CreateCommunity(community)
	if code == 0 {
		utils.RespOk(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// 加载群列表
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	data, msg := models.LoadCommumity(uint(ownerId))
	if len(data) == 0 {
		utils.RespFail(c.Writer, msg)
	} else {
		utils.RespList(c.Writer, 0, data, msg)
	}
}

// 加入群 userId uint comId string
func JoinGroups(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	comId := c.Request.FormValue("comId")
	code, msg := models.JoinGroup(uint(userId), comId)
	if code == 0 {
		utils.RespOk(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}
