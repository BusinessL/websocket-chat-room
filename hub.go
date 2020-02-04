package main

import "encoding/json"

// 将连接器初始化
var h = hub{
	connections: make(map[*connection]bool),
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
}

// 处理ws的逻辑实现
func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
			c.data.Ip = c.ws.RemoteAddr().String()
			c.data.Type = "handshake"
			c.data.UserList = user_list
			data_b, _ := json.Marshal(c.data)
			// 将数据放入数据管道
			c.send <- data_b
		case c := <-h.unregister:
			// 判断map里是否存在要删的数据
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case data := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- data:
				default:
					delete(h.connections, c)
					close(c.send)
				}
			}
		}
	}
}
