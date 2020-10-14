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
	consulta    = flag.String("id", "", "id de la consulta")
)

func main() {
	flag.Parse()
	fmt.Printf("tipo_cliente: %v\n", *tipoCliente)
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	client := protos.NewSolicitudClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
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
			fmt.Println(scanner.Text())
			lineaSeparada := strings.Split(scanner.Text(), ",")
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
			confirmation, err4 := client.MakeOrder(ctx, &order)
			fmt.Printf("%v\n", confirmation)
			fmt.Printf("%v\n", sample)

			if err4 != nil {
				panic(err3)
			}

			if err3 != nil {
				panic(err4)
			}

		}

		if err4 := scanner.Err(); err4 != nil {
			log.Fatal(err4)
		}
	}
	defer conn.Close()

}
