// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// modified by susilonurcahyo@gmail.com

package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"encoding/json"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {		
        return true
    },
}

// Client Information
type ClientInfo struct {
	ClientId string
	Username string
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// client info
	clientInfo *ClientInfo
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()

		// create messages object from message string
		var messages Messages
		log.Printf("Message: %s", message)
		if (string(message) != ""){
			err = json.Unmarshal(message, &messages)
			if err != nil {
				log.Fatalf("Error occured during unmarshaling. Please check message format, Error: %s", err.Error())
			}

			// message for server			
			if (messages.Cmd == "list") {
				// return to requester
				listclient, _ := connectedClients(c.hub.clients)
				messages.Sid = "server"
				messages.Did = c.clientInfo.ClientId
				messages.Msg = string(listclient)
			}

			// recreate messages object			
			newmessages := &Messages{Sid: c.clientInfo.ClientId, Did: messages.Did, Msg:messages.Msg, Cmd:messages.Cmd }		
			jsonmessages, err := json.Marshal(newmessages)
			if err != nil {
				log.Fatalf("Error occured during creating message object, Error: %s", err.Error())
			}	
			message = []byte(jsonmessages)
		} else {
			log.Printf("Empty Messages")
			newmessages := &Messages{Sid: "", Did: "", Msg:"", Cmd:"" }
			jsonmessages, _ := json.Marshal(newmessages)
			message = []byte(jsonmessages)
		}		

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unknown error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))	

		c.hub.broadcast <- message	
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	id := r.URL.Query().Get("id")
	username := r.URL.Query().Get("username")
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), clientInfo: &ClientInfo{ClientId:id, Username: username}}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
