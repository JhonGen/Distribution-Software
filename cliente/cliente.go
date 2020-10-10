package cliente

import{
  "google.golang.org/grpc"
}

func main(){
    conn, err := grpc.Dial("10.10.28.46:4040", grpc.WithInsecure())
    if err != nil {
      panic(err)
    }

    client := proto.NewOrderServiceClient(conn)

}
