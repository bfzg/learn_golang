package models

import (
	"encoding/json"
	"fmt"
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

		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			zap.S().Info("json解析失败", err)
			return
		}

		if msg.Type == 1 {
			zap.S().Info("这是一条私信:", msg.Content)
			tarNode, ok := clientMap[msg.TargetId]
			if !ok {
				zap.S().Info("不存在对应的node", msg.TargetId)
				return
			}

			tarNode.DataQueue <- data
			fmt.Println("发送成功：", string(data))
		}
	}
}
