package test

import (
	"Concord/Messaging"
	"context"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestDirectMessageServer(t *testing.T) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(Messaging.RPC_ADDRESS+":"+Messaging.RPC_PORT, grpc.WithInsecure())
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()
	c := Messaging.NewDirectMessageServiceClient(conn)
	message := Messaging.DirectMessage{
		SenderID:     "1001",
		SentTime:     time.Now().Unix(),
		Body:         "Test message from client",
		Attachments:  nil,
		RecipientIDs: nil,
	}

	response, err := c.DirectMessageUser(context.Background(), &message)
	if err != nil {
		t.Logf("Error sending client message: %s", err.Error())
		t.FailNow()
	}
	t.Logf("Server response: %s", response.ErrorMsg)

}
