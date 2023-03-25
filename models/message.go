package models

import (
	"GinChat/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// 通过在后面加上valid 可以实现验证是否符合规则
type Message struct {
	gorm.Model
	UserId     int64  //发送者
	TargetId   int64  //接受者
	Type       int    //发送类型  1私聊  2群聊  3心跳
	Media      int    //消息类型  1文字 2表情包 3语音 4图片 /表情包
	Content    string //消息内容
	CreateTime uint64 //创建时间
	ReadTime   uint64 //读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

/*
发送消息：发送者id，接收者id，消息类型，发送内容， 发送类型
校验token，关系
*/
type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 初始化
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLock sync.RWMutex

// 聊天
func Chat(write http.ResponseWriter, request *http.Request) {
	// 1.获取参数并校验合法性
	// 校验token
	//token := query.Get("token")
	query := request.URL.Query()
	userId := query.Get("userId")
	//targetId := query.Get("targetId")
	//context := query.Get("context")
	//msgType := query.Get("type")
	isValida := true // 这个应该是用根据userId 在数据库中查token 和我们传进来的token比较
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {

			return isValida
		},
	}).Upgrade(write, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 2. 获取conn
	node := &Node{
		// 这个conn和前端websocket初始化的连接是一致的，猜测都是指向同一个路由/chat
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	// 3. 用户关系
	// 4. userId和node绑定 并枷锁
	rwLock.Lock()
	id, _ := strconv.ParseInt(userId, 10, 64)
	clientMap[id] = node
	fmt.Println(id)
	rwLock.Unlock()
	// 5. 完成发送逻辑
	go sendProc(node)
	// 6. 完成接受逻辑
	go recvProc(node)
	sendMsg(id, []byte("欢迎进入聊天系统"))
}

// 这个sendProc recvProc协程表示的客户端，是用户发送消息/接收消息的协程
// 而下面的的udpSendProc udpRecvProc初始化就有的协程是服务器端，作用是作为客户端发送消息的中转
/*
	1.在前端页面中有websocket，和这个node.conn指向的是同一个路由，调用协程sendProc发送消息，服务器端的updRecvProc通过upd可以监听任何地址发来的消息，之后服务器会根据消息类型判断是私聊还是群发
	以私聊为例
	2.服务器端调用方法sendMsg，向私聊对象targetId发送消息
	3.此时targetId的客户 的recvProc协程一直监听自身DataQueue管道中数据，从而接收到数据

	udpSendProc方法是在广播的时候会向udpsendChan管道中传数据，因此会调用该协程
*/
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws]sendProc >>>>>> data: ", string(data))
			// 这个写是将服务器中的内容写到 客户端 打印出来
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {
		// 这个读文件是通过协程 读取客户端发给服务器的内容
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		broadMsg(data)
		fmt.Println("[ws]recvProc <<<<<< data: ", string(data))
	}
}

func sendMsg(userId int64, msg []byte) {
	// 这个userId 指的是targetID
	rwLock.RLock()
	node, ok := clientMap[userId]
	rwLock.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	// 接收者
	targetIdStr := strconv.Itoa(int(userId))
	// 发送者
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userId: ", userId, "msg: ", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	//命令返回有序集中，指定区间内的成员。
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result()
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)
}

// 需要重写方法 才能完整的将msg转为byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)

}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine ")
}

// 完成upd数据发送的协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(219, 216, 65, 75),
		Port: viper.GetInt("port.udp"),
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpSendProc: data ", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 完成接受
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: viper.GetInt("port.udp"),
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("updRecvProc <<<<<< data: ", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 这一步是服务器在做，判断是私聊还是群聊等类型
	switch msg.Type {
	case 1:
		// 私信
		fmt.Println("dispatch data: ", string(data))
		sendMsg(msg.TargetId, data)
	// 群发
	case 2:
		sendGroupMsg(msg.TargetId, data) // 发送的群ID， 消息内容
		// 广播
		//case 3:
		//	sendAllMsg()
	}
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息！")
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		// 排除给自己的
		if targetId != int64(userIds[i]) {
			sendMsg(targetId, msg)
		}
	}
}

func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	//contact.TargetId = comId
	contact.Type = 2
	community := Community{}
	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到该群聊！"
	}
	utils.DB.Where("owner_id=? and target_id=? and type=2", userId, comId).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "您已加入此群聊！"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功！"
	}
}
