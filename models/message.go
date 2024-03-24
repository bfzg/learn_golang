package models

import (
	"context"
	"encoding/json"
	"fmt"
	"ginchat/global"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Message struct {
	Model
	FormId   int64 `json:"userId"`   // 信息发送者
	TargetId int64 `json:"targetId"` //消息接收者
	Type     int   //消息类型
	Media    int
	Content  string
	Pic      string `json:"url"`
	Url      string
	Desc     string
	Amount   int //其他数据大小
}

func (m *Message) MsgTableName() string {
	return "message"
}

// Node 构造连接
type Node struct {
	Conn      *websocket.Conn //连接
	Addr      string          //客户端地址
	DataQueue chan []byte     //消息
	GroupSets set.Interface   //好友 群
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁，绑定node 时需要线程安全
var rwLocker sync.RWMutex

func Chat(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	Id := query.Get("UserId")
	userId, err := strconv.ParseInt(Id, 10, 64) //将Id 转换成十进制 64位
	if err != nil {
		zap.S().Info("类型转换失败", err)
		return
	}

	//升级位socket
	var isvalida = true

	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	//获取socket连接，构造消息节点
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}

	//将userId和Node绑定
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//服务发送消息
	go sendProc(node) //创建一个协程
	//服务接收消息
	go recProc(node)
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				zap.S().Info("发送消息失败", err)
				return
			}
			fmt.Println("数据发送成功")
		}
	}
}

func recProc(node *Node) {
	for {
		//获取信息
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			zap.S().Info("接收消息失败", err)
			return
		}

		bordMsg(data)
	}
}

// 全局 channel
var upSendChan chan []byte = make(chan []byte, 1024)

func bordMsg(data []byte) {
	upSendChan <- data
}

// init方法，运行message包前调用
func init() {
	go UdpSendProc()
	go UdpRecProc()
}

func UdpSendProc() {
	udpConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 3000,
		Zone: "",
	})
	if err != nil {
		zap.S().Info("拨号udp端口失败", err)
		return
	}

	defer udpConn.Close()

	for {
		select {
		case data := <-upSendChan:
			_, err := udpConn.Write(data)
			if err != nil {
				zap.S().Info("写入udp消息失败", err)
				return
			}
		}
	}
}

// 完成udp 数据的接收，启动udp 服务 获取udp客户端的写入的消息
func UdpRecProc() {
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{ //用于在指定网络和地址上监听连接。
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 3000,
	})
	if err != nil {
		zap.S().Info("监听udp端口失败", err)
		return
	}

	defer udpConn.Close()

	for { //无线循环写法
		var buf [1024]byte
		n, err := udpConn.Read(buf[0:]) //这里是将数据取出来表示将数据从0到最后全部截取下来
		if err != nil {
			zap.S().Info("接收udp消息失败", err)
			return
		}

		//处理发送逻辑
		dispatch(buf[0:n]) //0:n，这表示的是从切片 buf 中取出从下标 0 到下标 n-1 的部分，这里的 n 是上面读取操作返回的读取的字节数。这样的操作可以用来截取实际读取到的数据，因为 UDP 数据报的大小是不固定的，可能会小于缓冲区的大小，所以需要根据实际读取的字节数来截取有效的数据部分。
	}
}

// 解析消息 聊天类型判断
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		zap.S().Info("解析消息失败", err)
		return
	}

	fmt.Println("解析数据:", msg, "msg.FormId", msg.FormId, "targetId:", msg.TargetId, "type:", msg.Type)

	switch msg.Type {
	case 1: //私聊
		sendMsgAndSave(msg.TargetId, data)
	case 2: //群发
		sendGroupMsg(uint(msg.FormId), uint(msg.TargetId), data)
	}
}

// 发送消息 并存储聊天记录到redis
func sendMsgAndSave(userId int64, msg []byte) {
	rwLocker.RLock()              //读锁保证clientMap的写操作其他goroutine无法对clientMap进行写操作，但可以进行写操作
	node, ok := clientMap[userId] //保证对方在线
	rwLocker.RUnlock()

	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.FormId))

	if ok {
		//如果当前用户在线，将消息转发到当前用户的websocket连接中，然后进行存储
		node.DataQueue <- msg
	}

	var key string
	if userId > jsonMsg.FormId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}

	//创建记录
	res, err := global.RedisDB.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
		return
	}

	score := float64(cap(res)) + 1
	ress, e := global.RedisDB.ZAdd(ctx, key, &redis.Z{score, msg}).Result()
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Println(ress)
}

func sendGroupMsg(formId, target uint, data []byte) (int, error) {
	//群发的逻辑：1获取到群里所有用户，然后向除开自己的每一位用户发送消息
	userIDs, err := FindUsers(target)
	if err != nil {
		return -1, err
	}

	for _, userId := range *userIDs {
		if formId != userId {
			sendMsgAndSave(int64(userId), data)
		}
	}

	return 0, nil
}

func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))

	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}

	var rels []string
	var err error
	if isRev {
		rels, err = global.RedisDB.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = global.RedisDB.ZRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err)
	}
	return rels
}
