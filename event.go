package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage    = "send_message"
	EventRecieveMessage = "recieve_message"
	EventChangeRoom     = "change_room"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type RecieveMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

type ChatroomEvent struct {
	Name string `json:"name"`
}

func SendMessage(event Event, c *Client) error {
	var chatEvent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var broadcastMessage RecieveMessageEvent

	broadcastMessage.From = chatEvent.From
	broadcastMessage.Message = chatEvent.Message
	broadcastMessage.Sent = time.Now()

	data, err := json.Marshal(broadcastMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	var outgoingMessage Event
	outgoingMessage.Payload = data
	outgoingMessage.Type = EventRecieveMessage

	for client := range c.manager.clients {
		if client.chatroom == c.chatroom {
			client.egress <- outgoingMessage
		}
	}

	return nil
}

func ChatroomHandler(event Event, c *Client) error {
	var chatroom ChatroomEvent

	if err := json.Unmarshal(event.Payload, &chatroom); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	c.chatroom = chatroom.Name
	return nil
}
