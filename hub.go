// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// modified by susilonurcahyo@gmail.com

package main

import (
	"log"
	"encoding/json"
)
// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:	
			// new client connected, add to clients map	
			log.Printf("client connected: %s %s", client.clientInfo.ClientId, client.clientInfo.Username)
			h.clients[client] = true	
			// broadcast new user information
			err := broadcastUserStatus(client, h.clients, "join")
			if err != nil {
				log.Printf("Cannot broadcast new user info: %s", err)
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				// client disconnected, remove from clients map				
				log.Printf("client disconnected: %s %s", client.clientInfo.ClientId,  client.clientInfo.Username)
				delete(h.clients, client)
				close(client.send)
				// broadcast user leave information
				err := broadcastUserStatus(client, h.clients, "leave")
				if err != nil {
					log.Printf("Cannot broadcast new user info: %s", err)
				}
			}
		case message := <-h.broadcast:
			// scan all connected clients, only send to destination
			// TODO publish to kafka
			for client := range h.clients {	
				// unmarshal message to get destination id
				var messages Messages
				err := json.Unmarshal(message, &messages)
				if err != nil {
					log.Fatalf("Error occured during unmarshaling. Error: %s", err.Error())
				}				
				if (client.clientInfo.ClientId == messages.Did) {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				} 
			}
		}
	}
}
