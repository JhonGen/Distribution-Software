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

func orderInSlice(a *protos.Order, list []*protos.Order) bool {
	for _, b := range list {
		if b.Id == a.Id && a.Nombre == b.Nombre {
			return true
		}
	}
	return false
}

func main() {
	listener, err := net.Listen("tcp", "localhost:4040")
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

	return &protos.Sample{}, nil

}

func (s *LogisticaServer) MakeOrder(ctx context.Context, order *protos.Order) (*protos.Confirmation, error) {
	confirmation := &protos.Confirmation{}
	if !orderInSlice(order, s.queuedOrders) {
		s.queuedOrders = append(s.queuedOrders, order)
		confirmation.ConfirmationMessage = "Order added succesfully"
		return confirmation, nil
	} else {
		confirmation.ConfirmationMessage = "ORDER ALREADY IN QUEUE\n"
		return confirmation, nil
	}
}

func (s *LogisticaServer) GetStatus(ctx context.Context, id *protos.Order) (*protos.Status, error) {
	return nil, nil
}
