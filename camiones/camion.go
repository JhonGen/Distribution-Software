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

func intentarEntrega(camion *protos.Camion, cliente protos.SolicitudClient) *protos.Camion {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	chance1 := rand.Intn(100)
	chance2 := rand.Intn(100)
	defer cancel()
	if camion.Orden1 != nil {
		if chance1 <= 80 {
			fmt.Printf("Orden " + camion.Orden1.Nombre + " Entregada exitosamente\n")
			//ReporteEntrega(camion.Orden1)
			camion.Orden1.Apruebo = true
			confirmacion, _ := cliente.ReporteEntrega(ctx, camion.Orden1)
			fmt.Printf("%v\n", confirmacion)
			t := time.Now()

			value := strconv.Itoa(int(camion.Orden1.Valor))
			inte := strconv.Itoa(int(camion.Orden1.Intentos))
			f, err := os.OpenFile("camion"+*nro_camion+".txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString(camion.Orden1.Id + "," + camion.Orden1.TipoCliente + "," + value + "," + camion.Orden1.Tienda + "," + inte + "," + t.String() + "\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()

			camion.Orden1 = nil

		} else {
			fmt.Printf("Orden " + camion.Orden1.Nombre + " Entrega fallida\n")
			confirmacion, _ := cliente.ReporteEntrega(ctx, camion.Orden1)
			fmt.Printf("%v\n", confirmacion)
			value := strconv.Itoa(int(camion.Orden1.Valor))
			inte := strconv.Itoa(int(camion.Orden1.Intentos))
			f, err := os.OpenFile("camion"+*nro_camion+".txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString(camion.Orden1.Id + "," + camion.Orden1.TipoCliente + "," + value + "," + camion.Orden1.Tienda + "," + inte + "," + "0" + "\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()
			camion.Estado = "Con paquete de vuelta"
		}
	}
	if camion.Orden2 != nil {
		if chance2 <= 80 {
			t := time.Now()
			camion.Orden2.Apruebo = true
			confirmacion, _ := cliente.ReporteEntrega(ctx, camion.Orden2)
			fmt.Printf("%v\n", confirmacion)
			value := strconv.Itoa(int(camion.Orden2.Valor))
			inte := strconv.Itoa(int(camion.Orden2.Intentos))
			f, err := os.OpenFile("camion"+*nro_camion+".txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString(camion.Orden2.Id + "," + camion.Orden2.TipoCliente + "," + value + "," + camion.Orden2.Tienda + "," + inte + "," + t.String() + "\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()
			fmt.Printf("Orden " + camion.Orden2.Nombre + " Entregada exitosamente\n")
			camion.Orden2 = nil
		} else {
			confirmacion, _ := cliente.ReporteEntrega(ctx, camion.Orden2)
			fmt.Printf("%v\n", confirmacion)
			value := strconv.Itoa(int(camion.Orden2.Valor))
			inte := strconv.Itoa(int(camion.Orden2.Intentos))
			f, err := os.OpenFile("camion"+*nro_camion+".txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString(camion.Orden2.Id + "," + camion.Orden2.TipoCliente + "," + value + "," + camion.Orden2.Tienda + "," + inte + "," + "0" + "\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()
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
	conn, err := grpc.Dial("10.10.28.47:4040", grpc.WithInsecure())
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
	f1, _ := os.Create("camion1.txt")
	f1.Close()
	f2, _ := os.Create("camion2.txt")
	f2.Close()
	f3, _ := os.Create("camion3.txt")
	f3.Close()

	for true {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		camion, err2 := camionCliente.RetirarOrden(ctx, &camion)
		if err2 != nil {
			panic(err2)
		}

		fmt.Printf(camion.Estado + "\n")
		camion = intentarEntrega(camion, camionCliente)
		camion, err3 := camionCliente.DevolverOrden(ctx, camion)
		if err3 != nil {
			panic(err3)
		}

		cancel()
	}

}
