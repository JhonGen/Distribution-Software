package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	protos "../clienteproto"
	"google.golang.org/grpc"
)

type Solicitud struct {
	time.Time
	Order       *protos.Order
	Seguimiento int
	Status      string
	Intentos    int
}

type LogisticaServer struct {
	protos.UnimplementedSolicitudServer
	queuedPymes  []Solicitud
	queuedRetail []Solicitud
}

func orderInSlice(a *protos.Order, list []Solicitud) bool {
	for _, b := range list {
		if b.Order.Id == a.Id && a.Nombre == b.Order.Nombre {
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
	if !orderInSlice(order, s.queuedPymes) && !orderInSlice(order, s.queuedRetail) {
		solicitud := Solicitud{}
		t := time.Now()
		solicitud.Time = t
		solicitud.Order = order
		solicitud.Status = "En bodega"
		solicitud.Seguimiento = rand.Intn(999999999)
		solicitud.Intentos = 0
		if solicitud.Order.TipoCliente == "pymes" {
			s.queuedPymes = append(s.queuedPymes, solicitud)
		}
		confirmation.ConfirmationMessage = "Orden añadida satisfactoriamente, su codigo de seguimiento es: " + strconv.Itoa(solicitud.Seguimiento)
		if solicitud.Order.TipoCliente == "retail" {
			s.queuedRetail = append(s.queuedRetail, solicitud)
		}
		return confirmation, nil
	} else {
		confirmation.ConfirmationMessage = "La orden ya esta en cola"
		return confirmation, nil
	}
}

func (s *LogisticaServer) GetStatus(ctx context.Context, numero *protos.CodigoSeguimiento) (*protos.Status, error) {
	estado := &protos.Status{}
	for _, solicitud := range s.queuedPymes {
		if int32(solicitud.Seguimiento) == numero.Codigo {
			estado.State = solicitud.Status
			return estado, nil
		}
	}
	for _, solicitud := range s.queuedRetail {
		if int32(solicitud.Seguimiento) == numero.Codigo {
			estado.State = solicitud.Status
			return estado, nil
		}
	}

	estado.State = "No existe el pedido"
	return estado, nil
}
