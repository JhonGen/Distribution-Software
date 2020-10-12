package main

import (
	"context"
	"fmt"
	"net"

	protos "../clienteproto"
	"google.golang.org/grpc"
)

type LogisticaServer struct {
	protos.UnimplementedSolicitudServer
	queuedOrders []*protos.Order
}

func main() {
	listener, err := net.Listen("tcp", "10.10.28.47:4040")
	if err != nil {
		panic(err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	protos.RegisterSolicitudServer(grpcServer, &LogisticaServer{})
	fmt.Printf("escuchando")
	grpcServer.Serve(listener)

	defer grpcServer.Stop()
}

func (s *LogisticaServer) ShowOrder(ctx context.Context, order *protos.Order) (*protos.Sample, error) {

	fmt.Printf("%v", order)

	return &protos.Sample{}, nil

}

func (s *LogisticaServer) MakeOrder(ctx context.Context, order *protos.Order) (*protos.Confirmation, error) {
	return nil, nil
}

func (s *LogisticaServer) GetStatus(ctx context.Context, id *protos.Order) (*protos.Status, error) {
	return nil, nil
}
