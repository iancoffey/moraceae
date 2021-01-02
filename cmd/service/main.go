package main

import (
	"github.com/iancoffey/moraceae/proto"
	"google.golang.org/grpc"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

var (
	num  = rand.Intn(1000)
	port = ":10000"
)

type server struct {
	proto.UnimplementedGreeterServer
	name string
}

func (s *server) Greet(ctx context.Context, in *proto.GreetRequest) (*proto.GreetReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &proto.GreetReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	viper.SetEnvPrefix("MORACEAE")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	pflag.Int("PORT", 19000, "The port of this service") // each service gets its own port of course
	pflag.String("SERVICE_NAME", "", "service name")     //  each service gets a unique name

	servicePort := viper.GetInt("PORT")
	serviceName := viper.GetString("SERVICE_NAME")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	log.Printf("name=node-%d port=%d service-name=%s", num, servicePort, serviceName)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", servicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterGreeterServer(s, &server{name: serviceName})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
