package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	protos "../clienteproto"
	"google.golang.org/grpc"
)

var (
	tipoCliente = flag.String("tipo_cliente", "", "tipo de cliente")
	delay       = flag.Int("delay", 0, " tiempo de espera en segundos")
	consulta    = flag.Int("codigo", 0, "id de la consulta")
)

func ShowMakeOrder(linea string, client protos.SolicitudClient) {
	fmt.Println(linea)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	lineaSeparada := strings.Split(linea, ",")
	order := protos.Order{}
	order.Id = lineaSeparada[0]
	order.Nombre = lineaSeparada[1]
	valorInt, _ := strconv.ParseInt(lineaSeparada[2], 10, 32)
	order.Valor = int32(valorInt)
	order.Tienda = lineaSeparada[3]
	order.Destino = lineaSeparada[4]
	if *tipoCliente == "pymes" {
		prioritarioBool, _ := strconv.ParseBool(lineaSeparada[5])
		order.Prioritario = prioritarioBool
	}
	if *tipoCliente == "retail" {
		order.Prioritario = true
	}
	sample, err3 := client.ShowOrder(ctx, &order)
	if err3 != nil {
		panic(err3)
	}
	confirmation, err4 := client.MakeOrder(ctx, &order)
	if err4 != nil {
		panic(err4)
	}
	fmt.Printf("%v\n", confirmation)
	fmt.Printf("%v\n", sample)

	i := 1
	for i <= *delay {
		time.Sleep(time.Second)
		i += 1
	}

}

func obtenerEstado(codigo int, client protos.SolicitudClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	codigoSeguimiento := protos.CodigoSeguimiento{}
	codigoSeguimiento.Codigo = int32(codigo)
	status, err := client.GetStatus(ctx, &codigoSeguimiento)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", status)
}

func main() {
	flag.Parse()
	fmt.Printf("tipo_cliente: %v\n", *tipoCliente)
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	client := protos.NewSolicitudClient(conn)

	if err != nil {
		panic(err)
	}
	if *tipoCliente != "" {
		file, err2 := os.Open("../" + *tipoCliente + ".csv")
		if err2 != nil {
			log.Fatal(err2)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ShowMakeOrder(scanner.Text(), client)

		}

		if err4 := scanner.Err(); err4 != nil {
			log.Fatal(err4)
		}
	}

	defer conn.Close()
	obtenerEstado(*consulta, client)
}
