# go-websocket
使用Go语言基于 `github.com/gorilla/websocket` 封装的WebSocket网络层脚手架。

## Install
```bash
go get github.com/hiysz/go-websocket
```

## example
Using [Gin](https://github.com/gin-gonic/gin):
```go
package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hiysz/go-websocket"
)

func main() {
	router := gin.Default()
	g := gowebsocket.New()
	router.GET("/ws", func(ctx *gin.Context) {
		_ = g.HandleRequest(ctx.Writer, ctx.Request)
	})

	g.HandleTextMessage(func(s *gowebsocket.Conn, message []byte) {
		fmt.Println(string(message))
		_ = g.MessageBroadcast(&gowebsocket.MessageBody{Type: gowebsocket.TextMessage, Data: []byte("hello, MessageBroadcast")})
		_ = g.MessageUnicast(s.GetID(), &gowebsocket.MessageBody{Type: gowebsocket.TextMessage, Data: []byte("hello, MessageUnicast")})
	})

	g.HandleBinaryMessage(func(s *gowebsocket.Conn, message []byte) {
		fmt.Println(string(message))
	})

	g.HandleDisconnect(func(s *gowebsocket.Conn) {
		fmt.Println(s.GetID(), " connection disconnected!")
	})

	_ = router.Run(":8081")
}

```
