package main

import (
	"github.com/justmax437/avalonBacker/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)

func main() {
	grpcServer := grpc.NewServer()
	api.RegisterGameServiceServer(grpcServer,
		NewGameService(
			//NewMemoryStorage(30*time.Minute),
			NewMongoSessionStorage(),
			NewVoteStorage(),
		),
	)

	reflection.Register(grpcServer)

	err := grpcServer.Serve(getServerSocket())
	if err != nil {
		log.Fatal("failed to start gRPC server: ", err)
	}
}

func getServerSocket() net.Listener {
	var sock net.Listener
	if port, exist := os.LookupEnv("PORT"); exist {
		socket, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatal("failed to open socket - ", err)
		}
		sock = socket
	} else {
		log.Fatal("server port number env was not found (check $PORT)")
	}
	return sock
}
