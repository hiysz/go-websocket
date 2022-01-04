package gowebsocket

import (
	"github.com/gorilla/websocket"
	"github.com/hiysz/localcache"
	"log"
	"net/http"
)

type (
	handleConnectFunc func(*Conn)
	handleMessageFunc func(*Conn, []byte)
)

type wsServer struct {
	Upgrader             *websocket.Upgrader
	sessionManager       *ConnManager
	connectHandler       handleConnectFunc
	disconnectHandler    handleConnectFunc
	textMessageHandler   handleMessageFunc
	binaryMessageHandler handleMessageFunc
}

func New() *wsServer {
	storage := localcache.New()
	sessionManager := newConnManager(storage)
	go sessionManager.run()

	return &wsServer{
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		sessionManager: sessionManager,
	}
}

func (g *wsServer) HandleConnect(f handleConnectFunc) {
	g.connectHandler = f
}

func (g *wsServer) HandleDisconnect(f handleConnectFunc) {
	g.disconnectHandler = f
}

func (g *wsServer) HandleTextMessage(f handleMessageFunc) {
	g.textMessageHandler = f
}

func (g *wsServer) HandleBinaryMessage(f handleMessageFunc) {
	g.binaryMessageHandler = f
}

func (g *wsServer) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	var (
		conn *websocket.Conn
		err  error
	)

	if conn, err = g.Upgrader.Upgrade(w, r, nil); err != nil {
		return err
	}

	session, err := newConnection(conn)
	if err != nil {
		return err
	}

	g.sessionManager.registry <- session

	for {
		message, err := session.ReadMessage()
		if err != nil {
			break
		}

		switch message.Type {
		case websocket.TextMessage:
			if g.textMessageHandler != nil {
				g.textMessageHandler(session, message.Data)
			}
		case websocket.BinaryMessage:
			if g.binaryMessageHandler != nil {
				g.binaryMessageHandler(session, message.Data)
			}
		default:
			log.Print("未支持的消息类型.\n")
		}
	}

	session.Close()

	if g.disconnectHandler != nil {
		g.disconnectHandler(session)
	}

	return nil
}

func (g *wsServer) MessageBroadcast(message *MessageBody) error {
	g.sessionManager.broadcast <- message
	return nil
}

func (g *wsServer) MessageUnicast(toID string, message *MessageBody) error {
	g.sessionManager.unicast <- &inUnicast{toID: toID, message: message}
	return nil
}
