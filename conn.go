package gowebsocket

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Conn struct {
	id        string
	conn      *websocket.Conn
	inChan    chan *MessageBody
	outChan   chan *MessageBody
	closeChan chan byte
	config    *Config

	mutex    sync.Mutex
	isClosed bool
}

func newConnection(conn *websocket.Conn) (*Conn, error) {
	sessionID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	session := &Conn{
		id:        sessionID.String(),
		conn:      conn,
		inChan:    make(chan *MessageBody, 1000),
		outChan:   make(chan *MessageBody, 1000),
		closeChan: make(chan byte, 1),
		config:    newConfig(),
	}

	go session.readLoop()
	go session.writeLoop()

	return session, nil
}

func (s *Conn) ReadMessage() (data *MessageBody, err error) {
	select {
	case data = <-s.inChan:
	case <-s.closeChan:
		err = errors.New("connection is closeed.")
	}
	return
}

func (s *Conn) WriteMessage(data *MessageBody) (err error) {
	select {
	case s.outChan <- data:
	case <-s.closeChan:
		err = errors.New("connection is closeed.")
	}
	return err
}

func (s *Conn) Close() {
	_ = s.conn.Close()
	s.mutex.Lock()
	if !s.isClosed {
		close(s.closeChan)
		s.isClosed = true
	}
	s.mutex.Unlock()
	log.Println("connection breaklink client:", s.conn.RemoteAddr().String())
}

func (s *Conn) readLoop() {
	var (
		msgType int
		data    []byte
		err     error
	)

	s.conn.SetReadLimit(s.config.MaxMessageSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	s.conn.SetPongHandler(func(appData string) error {
		log.Print("收到pong \n")

		err := s.conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		return err
	})
	for {
		if msgType, data, err = s.conn.ReadMessage(); err != nil {
			s.Close()
			return
		}

		select {
		case s.inChan <- &MessageBody{Type: msgType, Data: data}:
		case <-s.closeChan:
			s.Close()
			return
		}
	}
}

func (s *Conn) writeLoop() {
	ticker := time.NewTicker(s.config.PingPeriod)
	defer func() {
		ticker.Stop()
	}()
	var (
		msgBody *MessageBody
		err     error
	)

	for {
		select {
		case msgBody = <-s.outChan:
			_ = s.conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))

			if err = s.conn.WriteMessage(msgBody.Type, msgBody.Data); err != nil {
				s.Close()
				return
			}
		case <-s.closeChan:
			s.Close()
			return
		case <-ticker.C:
			s.ping()
		}

	}

}

func (s *Conn) ping() {
	_ = s.conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))
	log.Print("发送ping \n")
	if err := s.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		return
	}
}

func (s *Conn) GetID() string {
	return s.id
}
