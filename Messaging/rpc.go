package Messaging

import (
	"Concord/CustomErrors"
	"context"
	"errors"
	"google.golang.org/grpc"
	"net"
)

const RPC_ADDRESS = ""
const RPC_PORT = "9000"

//Compile chat proto files
//protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative .\Messaging\chat.proto

type Server struct {
	UnimplementedDirectMessageServiceServer
	messageHub *Hub
}

func StartRPCServer(messageHub *Hub) {
	lis, err := net.Listen("tcp", RPC_ADDRESS+":"+RPC_PORT)
	if err != nil {
		CustomErrors.LogError(5023, CustomErrors.LOG_FATAL, true, err)
	}

	s := Server{messageHub: messageHub}
	grpcServer := grpc.NewServer()
	RegisterDirectMessageServiceServer(grpcServer, &s)

	CustomErrors.LogError(0, CustomErrors.LOG_INFO, false, errors.New("GRPC server started listening on "+RPC_ADDRESS+":"+RPC_PORT))
	err = grpcServer.Serve(lis)
	if err != nil {
		CustomErrors.LogError(5024, CustomErrors.LOG_FATAL, true, err)
	}

}

func (server *Server) DirectMessageUser(ctx context.Context, message *DirectMessage) (*DirectMessageResponse, error) {
	server.messageHub.hubDirectMessageUser <- message

	matchedUsers := server.messageHub.checkUsersExist(message.RecipientIDs)
	if len(message.RecipientIDs) == len(matchedUsers) {
		return &DirectMessageResponse{ErrorCode: 200, ErrorMsg: "ok", SuccessfulDeliveryIDs: matchedUsers}, nil
	}
	err := errors.New("not all recipients received the direct message")
	return &DirectMessageResponse{ErrorCode: 5028, ErrorMsg: err.Error(), SuccessfulDeliveryIDs: matchedUsers}, err
}
