package main

import (
	"context"
	"fmt"
	"time"

	protos "../clienteproto"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	fmt.Printf("conectado putita\n")
	defer conn.Close()
	client := protos.NewSolicitudClient(conn)
	order := protos.Order{}
	order.Id = "1234"
	order.Nombre = "elpichula"
	order.Valor = 15
	order.Tienda = "ctm"
	order.Destino = "tuvieja"
	order.Prioritario = 1
	fmt.Printf("%v\n", order)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sample, err := client.ShowOrder(ctx, &order)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", sample)
}
