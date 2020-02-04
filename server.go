package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	// 创建路由
	router := mux.NewRouter()
	go h.run()

	router.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe("127.0.0.1:8080", router); err != nil {
		fmt.Print("err", err)
	}
}
