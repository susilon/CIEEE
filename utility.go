/*
    Example Implementation of E2EE Chat Application
    Author : susilonurcahyo@gmail.com    
	Desc : 	
    Encryption are done at (client) HTML files not at the server
    Private Key are stored in browser javascript variable
    Server cannot decrypt and read messages
*/

package main

import (
	"bytes"
	"log"
	"encoding/json"
)

func connectedClients(clients map[*Client]bool) ([]byte, error){
	output := make([]ClientInfo, len(clients))

	ctr := 0
	for client := range clients {	
		output[ctr] = *client.clientInfo
		ctr = ctr + 1
	}

	return json.Marshal(output)
}

func broadcastUserStatus(c *Client, clients map[*Client]bool, status string) (error) {
	log.Printf("broadcasting user : %s", status)

	msg, err := json.Marshal(&c.clientInfo)
	
	for client := range clients {	
		messages := &Messages{Sid: "server", Did: client.clientInfo.ClientId, Msg:string(msg), Cmd:status }
		jsonmessages, _ := json.Marshal(messages)
		message := []byte(jsonmessages)	
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))	
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(clients, client)
		}
	}

	return err
}
