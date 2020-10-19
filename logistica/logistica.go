package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"time"

	protos "../protos"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

type Solicitud struct {
	time.Time
	Order       *protos.Order
	Seguimiento int
	Status      string
}

type LogisticaServer struct {
	protos.UnimplementedSolicitudServer
	queuedPymes        []Solicitud
	queuedPrioritarios []Solicitud
	queuedRetail       []Solicitud
	queuedReparto      []Solicitud
	queuedBalance      []Solicitud
}
type MensajeFinanzas struct {
	Order  []string
	Status string
}

func getIndex(cola []Solicitud, value Solicitud) int {
	for p, v := range cola {
		if v == value {
			return p
		}
	}
	return -1
}

func orderInSlice(a *protos.Order, list []Solicitud) bool {
	for _, b := range list {
		if b.Order.Id == a.Id && a.Nombre == b.Order.Nombre {
			return true
		}
	}
	return false
}

func remove(slice []Solicitud, s int) []Solicitud {
	return append(slice[:s], slice[s+1:]...)
}

func sumarIntentos(a *protos.Order, list []Solicitud, reparto []Solicitud) ([]Solicitud, []Solicitud) {

	solicitud := Solicitud{}
	if a != nil && reparto != nil {
		for i, b := range reparto {

			if b.Order.Id == a.Id && a.Nombre == b.Order.Nombre {

				fmt.Printf("nombre del fallo= %v\n", b.Order.Nombre)
				reparto[i].Order.Intentos += 1
				solicitud.Order = reparto[i].Order
				solicitud.Seguimiento = reparto[i].Seguimiento
				solicitud.Status = reparto[i].Status
				//inReparto = i
				break
			}
		}

		if a != nil {

			if a.TipoCliente == "retail" {

				if solicitud.Order.Intentos >= 3 {

				} else if solicitud.Order.Intentos < 3 {
					if len(list) > 1 {

						list = append(list, solicitud)
						copy(list[2:], list[1:])

						list[1] = solicitud
						fmt.Printf("cantidad de intentos: %v  \n", solicitud.Order.Intentos)
						return list, reparto
					} else if len(list) <= 1 {
						list = append(list, solicitud)
						fmt.Printf("cantidad de intentos: %v  \n", solicitud.Order.Intentos)
						return list, reparto

					}

				}
			}
		}
		if a.TipoCliente == "pymes" {
			if a.Valor <= 10*(1+int32(solicitud.Order.Intentos)) || solicitud.Order.Intentos >= 2 {
			} else {
				if len(list) > 1 {

					list = append(list, solicitud)
					copy(list[2:], list[1:])

					list[1] = solicitud
					fmt.Printf("cantidad de intentos: %v  \n", solicitud.Order.Intentos)
					return list, reparto

				} else if len(list) <= 1 {
					list = append(list, solicitud)
					fmt.Printf("cantidad de intentos: %v  \n", solicitud.Order.Intentos)
					return list, reparto

				}
			}
		}
	}

	return list, reparto
}

func ReportarFinanzas(solicitud Solicitud) {
	mensajeStruct := MensajeFinanzas{}
	var arregloOrden []string
	arregloOrden = append(arregloOrden, solicitud.Order.Id)
	arregloOrden = append(arregloOrden, solicitud.Order.Nombre)
	arregloOrden = append(arregloOrden, strconv.Itoa(int(solicitud.Order.Valor)))
	arregloOrden = append(arregloOrden, solicitud.Order.Tienda)
	arregloOrden = append(arregloOrden, solicitud.Order.Destino)
	arregloOrden = append(arregloOrden, strconv.FormatBool(solicitud.Order.Prioritario))
	arregloOrden = append(arregloOrden, strconv.Itoa(int(solicitud.Order.Intentos)))
	mensajeStruct.Order = arregloOrden
	mensajeStruct.Status = solicitud.Status
	mensaje, _ := json.Marshal(mensajeStruct)
	conn, err := amqp.Dial("amqp://admin:password@10.10.28.47:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"Finanzas", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	fmt.Println(q)
	err = ch.Publish(
		"",
		"Finanzas",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(mensaje),
		},
	)

	failOnError(err, "fallo el publish ctm")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	listener, err := net.Listen("tcp", "10.10.28.47:4040")
	if err != nil {
		panic(err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	protos.RegisterSolicitudServer(grpcServer, &LogisticaServer{})
	fmt.Printf("escuchando\n")
	grpcServer.Serve(listener)

	defer grpcServer.Stop()
}

func (s *LogisticaServer) ShowOrder(ctx context.Context, order *protos.Order) (*protos.Sample, error) {

	return &protos.Sample{}, nil

}

func (s *LogisticaServer) MakeOrder(ctx context.Context, order *protos.Order) (*protos.Confirmation, error) {
	confirmation := &protos.Confirmation{}
	if !orderInSlice(order, s.queuedPymes) && !orderInSlice(order, s.queuedRetail) && !orderInSlice(order, s.queuedPrioritarios) {
		solicitud := Solicitud{}
		t := time.Now()
		solicitud.Time = t
		solicitud.Order = order
		solicitud.Status = "En bodega"
		solicitud.Seguimiento = rand.Intn(999999999)
		if solicitud.Order.TipoCliente == "pymes" {
			if solicitud.Order.Prioritario {
				s.queuedPrioritarios = append(s.queuedPrioritarios, solicitud)
				sort.SliceStable(s.queuedPrioritarios, func(p, q int) bool {
					return s.queuedPrioritarios[p].Order.Valor > s.queuedPrioritarios[q].Order.Valor
				})
			} else {
				s.queuedPymes = append(s.queuedPymes, solicitud)
				sort.SliceStable(s.queuedPymes, func(p, q int) bool {
					return s.queuedPymes[p].Order.Valor > s.queuedPymes[q].Order.Valor
				})
			}

		}
		if solicitud.Order.TipoCliente == "retail" {
			s.queuedRetail = append(s.queuedRetail, solicitud)
			sort.SliceStable(s.queuedRetail, func(p, q int) bool {
				return s.queuedRetail[p].Order.Valor > s.queuedRetail[q].Order.Valor
			})
		}
		confirmation.ConfirmationMessage = "Orden añadida satisfactoriamente, su codigo de seguimiento es: " + strconv.Itoa(solicitud.Seguimiento)
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

func (s *LogisticaServer) RetirarOrden(ctx context.Context, camion *protos.Camion) (*protos.Camion, error) {
	i := int32(1)
	for i <= camion.TiempoEspera {
		if camion.Tipo == "pymes" {
			if camion.Orden1 == nil {
				if len(s.queuedPrioritarios) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedPrioritarios[0])
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]
				} else if len(s.queuedPymes) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedPymes[0])
					camion.Orden1, s.queuedPymes = s.queuedPymes[0].Order, s.queuedPymes[1:]

				}

			}
			if camion.Orden2 == nil {
				if len(s.queuedPrioritarios) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedPrioritarios[0])
					camion.Orden2, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]
				} else if len(s.queuedPymes) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedPymes[0])
					camion.Orden2, s.queuedPymes = s.queuedPymes[0].Order, s.queuedPymes[1:]

				}
			}

		}
		if camion.Tipo == "retail" {

			if camion.Orden1 == nil {
				if len(s.queuedRetail) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedRetail[0])
					camion.Orden1, s.queuedRetail = s.queuedRetail[0].Order, s.queuedRetail[1:]
				} else if len(s.queuedPrioritarios) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedPrioritarios[0])
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]

				}

			}
			if camion.Orden2 == nil {
				s.queuedRetail, s.queuedReparto = sumarIntentos(camion.Orden1, s.queuedRetail, s.queuedReparto)
				if len(s.queuedRetail) > 0 {
					s.queuedReparto = append(s.queuedReparto, s.queuedRetail[0])
					camion.Orden2, s.queuedRetail = s.queuedRetail[0].Order, s.queuedRetail[1:]

				}
			}
		}
		time.Sleep(time.Second)
		if camion.Orden1 == nil && camion.Orden2 == nil {
			fmt.Printf("No hay ordenes en cola, esperando un paquete, intentos: %v\n", i)
		}
		i += 1
	}
	fmt.Printf("Orden creada con exito:\n")
	fmt.Printf("Orden nro 1: %v\n", camion.Orden1)
	fmt.Printf("Orden nro 2: %v\n", camion.Orden2)

	return camion, nil
}

func (s *LogisticaServer) DevolverOrden(ctx context.Context, camion *protos.Camion) (*protos.Camion, error) {
	if camion.Orden1 != nil {
		fmt.Printf("el pedido %v falló", camion.Orden1.Nombre)
		if camion.Tipo == "pymes" {
			s.queuedPymes, s.queuedReparto = sumarIntentos(camion.Orden1, s.queuedPymes, s.queuedReparto)
		} else if camion.Tipo == "retail" {
			s.queuedRetail, s.queuedReparto = sumarIntentos(camion.Orden1, s.queuedRetail, s.queuedReparto)
		}
	}
	if camion.Orden2 != nil {
		fmt.Printf("el pedido %v falló", camion.Orden2.Nombre)
		if camion.Tipo == "pymes" {
			s.queuedPymes, s.queuedReparto = sumarIntentos(camion.Orden2, s.queuedPymes, s.queuedReparto)
		} else if camion.Tipo == "retail" {
			s.queuedRetail, s.queuedReparto = sumarIntentos(camion.Orden2, s.queuedRetail, s.queuedReparto)
		}
	}
	camion.Orden1 = nil
	camion.Orden2 = nil
	return camion, nil
}

func (s *LogisticaServer) ReporteEntrega(ctx context.Context, orden *protos.Order) (*protos.Confirmation, error) {
	solicitud := Solicitud{}
	confirmation := &protos.Confirmation{}
	solicitud.Order = orden
	fmt.Printf("%v\n", orden.Apruebo)
	if orden.Apruebo {
		solicitud.Status = "Entregado"
		s.queuedBalance = append(s.queuedBalance, solicitud)
		confirmation.ConfirmationMessage = "Orden enviada a balance"
		ReportarFinanzas(solicitud)
		return confirmation, nil
	} else {
		if orden.TipoCliente == "retail" {
			if orden.Intentos >= 3 {
				solicitud.Status = "No Entregado"
				confirmation.ConfirmationMessage = "Orden enviada a balance"
				s.queuedBalance = append(s.queuedBalance, solicitud)
				ReportarFinanzas(solicitud)
				return confirmation, nil
			}
		} else if orden.TipoCliente == "pymes" {
			if orden.Valor >= 10*(orden.Intentos) || orden.Intentos >= 2 {
				solicitud.Status = "No Entregado"
				confirmation.ConfirmationMessage = "Orden enviada a balance"
				s.queuedBalance = append(s.queuedBalance, solicitud)
				ReportarFinanzas(solicitud)
				return confirmation, nil

			}
		}
	}
	confirmation.ConfirmationMessage = "Orden se devolvera a la cola"
	return confirmation, nil
}
