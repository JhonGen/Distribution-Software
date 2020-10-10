package logistica

import{
  "context"
  "encoding/jason"
  "flag"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "math"
  "net"
  "sync"
  "time"

  "google.golang.org/grpc"


}

type LogisticaServer struct{}

func main(){
  listener, err := net.Listen("tcp", ":4040")
  if err !=nil {
    panic(err)
  }
  srv := grpc.NewServer()
  protos.RegisterMostrarOrdenService(srv, &LogisticaServer{})

  if peticion := srv.Serve(listener); peticion != nil{
    panic(err)
  }
}

func (s *LogisticaServer) MostrarOrden(ctx context.Context, *protos Orden)(*protos.Response,error){

}

func (s *LogisticaServer) Ordenar(ctx context.Context, *proto Orden)(*proto.Response,error){
  return nil;
}
