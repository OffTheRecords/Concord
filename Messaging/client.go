package Messaging

import (
	"Concord/CustomErrors"
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
	send chan *DirectMessage

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
	client := &Client{hub: hub, conn: ws, send: make(chan *DirectMessage, 256), userID: userID}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	//go client.readPump()
}

func (c *Client) writePump() {
	//Ticker used to check if client is alive
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			CustomErrors.LogError(5025, CustomErrors.LOG_WARNING, false, err)
			return
		}
	}()
	for {
		select {
		//Process a message from the hub
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				CustomErrors.LogError(5026, CustomErrors.LOG_WARNING, false, err)
				return
			}
			if !ok {
				// The hub closed the channel.
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					CustomErrors.LogError(5027, CustomErrors.LOG_WARNING, false, err)
					return
				}
				return
			}

			//Serialize message into bytes so it can be written
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			messageJson, _ := json.Marshal(message)
			messageBytes := []byte(messageJson)
			_, err = w.Write(messageBytes)
			if err != nil {
				CustomErrors.LogError(5027, CustomErrors.LOG_WARNING, false, err)
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				messageJson, _ = json.Marshal(<-c.send)
				messageBytes = []byte(messageJson)
				_, err = w.Write(messageBytes)
				if err != nil {
					CustomErrors.LogError(5027, CustomErrors.LOG_WARNING, false, err)
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		// Close websocket connection if ping fails
		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				CustomErrors.LogError(5026, CustomErrors.LOG_WARNING, false, err)
				return
			}
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
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
