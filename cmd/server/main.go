package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"havana/protobuf"
	"log"
	"net"
	"time"
)

type helloServer struct {
}

func (h *helloServer) SayHello(ctx context.Context, helloRequest *pb.HelloRequest) (*pb.HelloResponse, error) {
	println("SayHello:", helloRequest.Greeting)
	return &pb.HelloResponse{Reply: "Hello!"}, nil
}

func (h *helloServer) LotsOfReplies(helloRequest *pb.HelloRequest, stream pb.HelloService_LotsOfRepliesServer) error {
	println("LotsOfReplies:", helloRequest.Greeting)

	for {
		now := time.Now().String()

		println("LotsOfReplies Send:", now)
		err := stream.Send(&pb.HelloResponse{Reply: now})

		if err != nil {
			println("LotsOfReplies:", err.Error())
			return err
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (h *helloServer) LotsOfGreetings(stream pb.HelloService_LotsOfGreetingsServer) error {
	for {
		hello, err := stream.Recv()

		if err != nil {
			println("LotsOfGreetings:", err.Error())

			reply := "LotsOfGreetings SendAndClose 교신 종료"
			println(reply)
			return stream.SendAndClose(&pb.HelloResponse{Reply: reply})
		}

		println("LotsOfGreetings:", hello.Greeting)
	}

	return nil
}

func (h *helloServer) BidiHello(stream pb.HelloService_BidiHelloServer) error {
	done := make(chan struct{}, 2)

	go func() {
		for {
			now := time.Now().String()
			println("BidiHello Send:", now)

			err := stream.Send(&pb.HelloResponse{Reply: now})

			if err != nil {
				println("BidiHello Send:", err.Error())
				break
			}

			time.Sleep(1 * time.Second)
		}

		done <- struct{}{}
	}()

	go func() {
		for {
			hello, err := stream.Recv()

			if err != nil {
				println("BidiHello Recv:", err.Error())
				break
			} else {
				println("BidiHello Recv:", hello.Greeting)
			}
		}

		done <- struct{}{}
	}()

	for i := 0; i < 2; i++ {
		<-done
	}

	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", "0.0.0.0:27015")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterHelloServiceServer(grpcServer, &helloServer{})

	println("Server ready.")
	grpcServer.Serve(lis)
}
