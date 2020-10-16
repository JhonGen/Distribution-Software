package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	protos "../protos"
	"google.golang.org/grpc"
)

var (
	tipoCamion    = flag.String("tipo_camion", "", "tipo de camion que se desea implementar")
	tiempoEntrega = flag.Int("tiempo_entrega", 0, "tiempo que tarda en realizar una entrega")
	delay         = flag.Int("delay", 0, "tiempo de espera de solicitud de un camion")
	nro_camion    = flag.Int("delay", 0, "camion a escojer")
)

func intentarEntrega(camion *protos.Camion) *protos.Camion {
	chance1 := rand.Intn(100)
	chance2 := rand.Intn(100)
	if chance1 >= 80 {
		camion.Orden1 = nil

	}
	if chance2 >= 80 {
		camion.Orden2 = nil
	}
	i := 1
	for i <= *tiempoEntrega {
		time.Sleep(time.Second)
		i += 1
	}
	return camion
}

func main() {
	flag.Parse()
	fmt.Printf("tipo_camion: %v\n", *tipoCamion)
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	camionCliente := protos.NewSolicitudClient(conn)
	camion := protos.Camion{}
	camion.Tipo = *tipoCamion
	camion.Estado = "En espera a recibir paquetes"
	camion.TiempoEspera = int32(*delay)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for true {

		camion, err2 := camionCliente.RetirarOrden(ctx, &camion)
		if err2 != nil {
			panic(err2)
		}
		camion = intentarEntrega(camion)
	}

}
