package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"time"

	protos "../protos"
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
	queuedPymes        []Solicitud
	queuedPrioritarios []Solicitud
	queuedRetail       []Solicitud
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

func sumarIntentos(a *protos.Order, list []Solicitud) {
	solicitud := Solicitud{}
	for _, b := range list {
		if b.Order.Id == a.Id && a.Nombre == b.Order.Nombre {
			solicitud = b
			break
		}
	}
	solicitud.Intentos += 1
	if solicitud.Intentos >= 3 {
		index := getIndex(list, solicitud)
		remove(list, index)

	}
	if len(list) != 0 {
		list = append(list, solicitud)
		copy(list[2:], list[1:])

		list[1] = solicitud
	}

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
	fmt.Printf("%v", camion.TiempoEspera)
	fmt.Printf(camion.Tipo + "\n")
	for i <= camion.TiempoEspera {

		if camion.Orden1 == nil && camion.Orden2 == nil {
			fmt.Printf("No hay ordenes en cola, camion termina el servicio\n")
		}
		fmt.Printf("entre al for\n")
		if camion.Tipo == "pymes" {
			if camion.Orden1 != nil {
				sumarIntentos(camion.Orden1, s.queuedPymes)
				if len(s.queuedPrioritarios) > 0 {
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]
				} else if len(s.queuedPymes) > 0 {
					camion.Orden1, s.queuedPymes = s.queuedPymes[0].Order, s.queuedPymes[1:]

				}

			} else {
				fmt.Printf("basta de hablar de huevito rey\n")
				if len(s.queuedPrioritarios) > 0 {
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]
				} else if len(s.queuedPymes) > 0 {
					camion.Orden1, s.queuedPymes = s.queuedPymes[0].Order, s.queuedPymes[1:]

				}
			}
			if camion.Orden2 != nil {
				sumarIntentos(camion.Orden2, s.queuedPymes)
				if len(s.queuedPrioritarios) > 0 {
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]
				} else if len(s.queuedPymes) > 0 {
					camion.Orden1, s.queuedPymes = s.queuedPymes[0].Order, s.queuedPymes[1:]

				}
			} else {
				if len(s.queuedPrioritarios) > 0 {
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]
				} else if len(s.queuedPymes) > 0 {
					camion.Orden1, s.queuedPymes = s.queuedPymes[0].Order, s.queuedPymes[1:]

				}
			}

		}
		if camion.Tipo == "retail" {
			fmt.Printf("estoy en la primera wea\n")
			if camion.Orden1 != nil {
				sumarIntentos(camion.Orden1, s.queuedRetail)
				if len(s.queuedRetail) > 0 {
					camion.Orden1, s.queuedRetail = s.queuedRetail[0].Order, s.queuedRetail[1:]
				} else if len(s.queuedPrioritarios) > 0 {
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]

				}

			} else {
				fmt.Printf("estoy aca\n")
				if len(s.queuedRetail) > 0 {
					camion.Orden1, s.queuedRetail = s.queuedRetail[0].Order, s.queuedRetail[1:]
				} else if len(s.queuedPrioritarios) > 0 {
					camion.Orden1, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]

				}

			}
			if camion.Orden2 != nil {
				sumarIntentos(camion.Orden1, s.queuedPymes)
				if len(s.queuedRetail) > 0 {
					camion.Orden2, s.queuedRetail = s.queuedRetail[0].Order, s.queuedRetail[1:]

				} else if len(s.queuedPrioritarios) > 0 {
					camion.Orden2, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]

				}

			} else {
				if len(s.queuedRetail) > 0 {
					fmt.Printf("estoy en la segunda wea\n")
					camion.Orden2, s.queuedRetail = s.queuedRetail[0].Order, s.queuedRetail[1:]
				} else if len(s.queuedPrioritarios) > 0 {
					camion.Orden2, s.queuedPrioritarios = s.queuedPrioritarios[0].Order, s.queuedPrioritarios[1:]

				}

			}

		}
		time.Sleep(time.Second)
		i += 1
	}

	return camion, nil

}
func (s *LogisticaServer) DevolverOrden(ctx context.Context, camion *protos.Camion) (*protos.Camion, error) {
	return nil, nil
}
