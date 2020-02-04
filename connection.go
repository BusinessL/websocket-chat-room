package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

// 抽象出需要的数据结构
// ws连接器   数据   管道

type connection struct {
	// ws连接器
	ws *websocket.Conn
	// 管道
	send chan []byte
	// 数据
	data *Data
}

// 抽象ws连接器
// 处理ws中的各种逻辑
type hub struct {
	// connections 注册了连接器
	connections map[*connection]bool
	// 从连接器发送的信息
	broadcast chan []byte
	// 从连接器注册请求
	register chan *connection
	// 销毁请求
	unregister chan *connection
}

// 先实现ws的读和写
// ws中写数据
func (c *connection) writer() {
	//管道遍历数据
	for message := range c.send {
		//数据写出
		c.ws.WriteMessage(websocket.TextMessage, message)
	}
	c.ws.Close()
}

var user_list = []string{}

// 连接中读数据
func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			// 读不进数据，将用户移除
			h.unregister <- c
			break
		}
		// 读取数据
		json.Unmarshal(message, &c.data)
		// 根据data的type判断该做什么
		switch c.data.Type {
		case "login":
			// 弹出窗口，输入用户名
			c.data.User = c.data.Content
			c.data.From = c.data.Content
			// 登录后，将用户加入到用户列表
			user_list = append(user_list, c.data.Content)
			c.data.UserList = user_list
			// 数据序列化
			data_b, _ := json.Marshal(c.data)
			h.broadcast <- data_b

		case "user":
			c.data.Type = "user"
			data_b, _ := json.Marshal(c.data)
			h.broadcast <- data_b
		case "logout":
			c.data.Type = "logout"
			user_list = remove(user_list, c.data.User)
			data_b, _ := json.Marshal(c.data)
			h.broadcast <- data_b
			h.unregister <- c
		default:
			fmt.Print("========default================")

		}
	}
}

// 删除用户切片中数据
func remove(slice []string, user string) []string {
	count := len(slice)
	if count == 0 {
		return slice
	}

	if count == 1 && slice[0] == user {
		return []string{}
	}

	// 定义新的返回切片
	var n_slice = []string{}
	// 删除传入切片中的指定用户，其他用户放到新的切片
	for i := range slice {
		if slice[i] == user && i == count {
			return slice[:count]
		} else if slice[i] == user {
			n_slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}

	return n_slice
}

// 定义升级器，将http请求升级为ws请求
var upgrader = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ws的回调函数
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// 获取ws对象
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// 创建连接对象
	c := &connection{send: make(chan []byte, 128), ws: ws, data: &Data{}}

	h.register <- c
	go c.writer()
	c.reader()

	defer func() {
		c.data.Type = "logout"
		user_list = remove(user_list, c.data.User)
		data_b, _ := json.Marshal(c.data)
		h.broadcast <- data_b
		h.unregister <- c
	}()
}
