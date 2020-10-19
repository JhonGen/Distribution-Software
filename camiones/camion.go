package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	protos "../protos"
	"google.golang.org/grpc"
)

var (
	tipoCamion    = flag.String("tipo_camion", "", "tipo de camion que se desea implementar")
	tiempoEntrega = flag.Int("tiempo_entrega", 0, "tiempo que tarda en realizar una entrega")
	delay         = flag.Int("delay", 0, "tiempo de espera de solicitud de un camion")
	nro_camion    = flag.String("nro_camion", "", "camion a escojer")
)

func intentarEntrega(camion *protos.Camion) *protos.Camion {
	chance1 := rand.Intn(100)
	chance2 := rand.Intn(100)
	if camion.Orden1 != nil {
		if chance1 <= 80 {
			fmt.Printf("Orden " + camion.Orden1.Nombre + " Entregada exitosamente\n")
			//ReporteEntrega(camion.Orden1)

			t := time.Now()

			value := strconv.Itoa(int(camion.Orden1.Valor))
			f, err := os.OpenFile("Camion1.txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString(camion.Orden1.Id + "," + camion.Orden1.TipoCliente + "," + value + "," + camion.Orden1.Tienda + "," + t.String() + "\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()

			camion.Orden1 = nil

		} else {
			fmt.Printf("Orden " + camion.Orden1.Nombre + " Entrega fallida\n")
			camion.Estado = "Con paquete de vuelta"
		}
	}
	if camion.Orden2 != nil {
		if chance2 <= 80 {
			fmt.Printf("Orden " + camion.Orden2.Nombre + " Entregada exitosamente\n")
			camion.Orden2 = nil
		} else {
			fmt.Printf("Orden " + camion.Orden2.Nombre + " Entrega fallida\n")
			camion.Estado = "Con paquete de vuelta"
		}
	}
	i := 1
	for i <= *tiempoEntrega {
		time.Sleep(time.Second)
		i += 1
	}
	if camion.Orden1 == nil && camion.Orden2 == nil {
		camion.Estado = "Camion en Espera"
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
	camion.Orden1 = nil
	camion.Orden2 = nil
	f, err := os.Create("Camion1.txt")
	f.Close()

	for true {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		camion, err2 := camionCliente.RetirarOrden(ctx, &camion)
		if err2 != nil {
			panic(err2)
		}

		fmt.Printf(camion.Estado + "\n")
		camion = intentarEntrega(camion)
		camion, err3 := camionCliente.DevolverOrden(ctx, camion)
		if err3 != nil {
			panic(err3)
		}

		cancel()
	}

}
