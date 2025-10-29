package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress is used to avoid concurrent writes on the websocket connection
	egress chan []byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

func (c *Client) readMessages() {
	defer func() {
		// we remove/cleanup the connection
		c.manager.removeClient(c)
	}()

	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		log.Println(string(payload))
	}
}

func (c *Client) writeMessages() {
	defer func() {
		// we remove/cleanup the connection
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("Connection closed", err)
				}
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("error writing message: %v", err)
			}

			log.Println("Message sent")
		}
	}
}
