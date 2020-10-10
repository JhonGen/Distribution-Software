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
  protos "github.com/JhonGen/TAREA1_SD/tree/main/clienteproto"

}

type LogisticaServer struct{}

func main(){
  listener, err := net.Listen("tcp", ":4040")
  if err !=nil {
    panic(err)
  }
}

func (s *LogisticaServer) ShowOrder(ctx context.Context, )(*protos.Response,error){

}

func (s *LogisticaServer) MakeOrder(ctx context.Context, )(*protos.Response,error){
  return nil;
}

func (s *LogisticaServer) GetStatus(ctx context.Context, id *protos.Order)(*protos.Response, error){

}
