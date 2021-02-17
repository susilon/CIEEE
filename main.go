// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
    Example Implementation of E2EE Chat Application
    Author : susilonurcahyo@gmail.com    
	Desc : 	
    Encryption are done at (client) HTML files not at the server
    Private Key are stored in browser javascript variable
    Server cannot decrypt and read messages

	This software is for demo purposes only
	Do not use in production
*/

package main

import (
	"flag"
	"log"
	"net/http"
)

// put your desired port number here
var addr = flag.String("addr", ":8380", "http service address")

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()	
	fileServer := http.FileServer(http.Dir("./html"))
	http.Handle("/", http.StripPrefix("/", fileServer))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	log.Printf("Listening at %s", *addr)
	err := http.ListenAndServe(*addr, nil)	
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
