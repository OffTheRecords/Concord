package Messaging

//https://github.com/gorilla/websocket/tree/master/examples/chat

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//Send messages to the user from rpc/local.
	hubDirectMessageUser chan *DirectMessage
}

func NewHub() *Hub {
	return &Hub{
		register:             make(chan *Client),
		unregister:           make(chan *Client),
		clients:              make(map[*Client]bool),
		hubDirectMessageUser: make(chan *DirectMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {

		//Add client to client list (map)
		case client := <-h.register:
			h.clients[client] = true

		//Remove client from list (map)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		//Send user direct message
		//TODO optimize
		case sendUserDirectMessage := <-h.hubDirectMessageUser:
			for client := range h.clients {
				for i := 0; i < len(sendUserDirectMessage.RecipientIDs); i++ {
					if client.userID == sendUserDirectMessage.RecipientIDs[i] {
						select {
						case client.send <- sendUserDirectMessage:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}
			}
		}
	}
}

func (h *Hub) checkUsersExist(userIDs []string) []string {
	var foundUserIDs []string

	for client := range h.clients {
		for i := 0; i < len(userIDs); i++ {
			if client.userID == userIDs[i] {
				foundUserIDs = append(foundUserIDs, client.userID)
			}
		}
	}
	return removeDuplicateStringsFromSlice(foundUserIDs)
}
