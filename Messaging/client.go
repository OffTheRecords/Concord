package Messaging

import (
	"Concord/Structures"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true //TODO only allow certain origins //https://stackoverflow.com/a/65039729
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	//Client id
	userID string
}

func ClientMessageReceiverHandler(hub *Hub, userID string, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO
		fmt.Println("upgrade:", err)
		return
	}

	err = ws.WriteMessage(1, []byte("Connection successful"))
	if err != nil {
		// TODO
		fmt.Println("write:", err)
		return
	}

	//Register client on hub
	client := &Client{hub: hub, conn: ws, send: make(chan []byte, 256), userID: userID}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	//go client.writePump()
	//go client.readPump()

	//TODO Temp
	err = ws.Close()
	if err != nil {
		return
	}
}

func ClientWSErrorResponse(w http.ResponseWriter, r *http.Request, response Structures.Response) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// TODO
		fmt.Println("upgrade:", err)
		return
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			//TODO
			fmt.Println("close:", err)
		}
	}(ws)

	//JSON response
	responseJson, _ := json.Marshal(response)

	err = ws.WriteMessage(1, []byte(responseJson))
	if err != nil {
		// TODO
		fmt.Println("write:", err)
		return
	}
}