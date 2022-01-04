package gowebsocket

import (
	"github.com/hiysz/localcache"
)

type inUnicast struct {
	toID    string
	message *MessageBody
}

type ConnManager struct {
	storage    localcache.ICache
	registry   chan *Conn
	unregister chan *Conn
	unicast    chan *inUnicast
	broadcast  chan *MessageBody
}

func newConnManager(storage localcache.ICache) *ConnManager {
	return &ConnManager{
		storage:    storage,
		registry:   make(chan *Conn),
		unregister: make(chan *Conn),
		unicast:    make(chan *inUnicast),
		broadcast:  make(chan *MessageBody),
	}
}

func (sm *ConnManager) run() {
	for {
		select {
		case client := <-sm.registry:
			_ = sm.storage.Set(client.id, client)

		case client := <-sm.unregister:
			_, _ = sm.storage.Del(client.id)

		case unicast := <-sm.unicast:
			if client, err := sm.storage.Get(unicast.toID); err == nil {
				select {
				case client.(*Conn).outChan <- unicast.message:
				default:
					close(client.(*Conn).outChan)
					_, _ = sm.storage.Del(unicast.toID)
				}
			}

		case message := <-sm.broadcast:
			sm.storage.FetchAll(func(v interface{}) {
				client := v.(*Conn)
				select {
				case client.outChan <- message:
				default:
					close(client.outChan)
					_, _ = sm.storage.Del(client.id)
				}
			})

		}
	}
}
