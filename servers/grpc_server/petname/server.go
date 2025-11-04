package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	petname "github.com/dustinkirkland/golang-petname"
	petnamepb "github.com/guryev-vladislav/go-toolkit/servers/grpc_server/petname/proto"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Config struct {
	GRPCPort string `yaml:"grpc_port" env:"PETNAME_GRPC_PORT" env-default:"28081"`
}

type server struct {
	petnamepb.UnimplementedPetnameGeneratorServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *server) Generate(ctx context.Context, req *petnamepb.PetnameRequest) (*petnamepb.PetnameResponse, error) {
	if req.Words <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "words must be greater than 0")
	}

	name := petname.Generate(int(req.Words), req.Separator)

	return &petnamepb.PetnameResponse{
		Name: name,
	}, nil
}

func (s *server) GenerateMany(req *petnamepb.PetnameStreamRequest, stream petnamepb.PetnameGenerator_GenerateManyServer) error {
	if req.Words <= 0 {
		return status.Errorf(codes.InvalidArgument, "words must be greater than 0")
	}
	if req.Names <= 0 {
		return status.Errorf(codes.InvalidArgument, "names must be greater than 0")
	}

	for range req.Names {
		name := petname.Generate(int(req.Words), req.Separator)

		if err := stream.Send(&petnamepb.PetnameResponse{
			Name: name,
		}); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	var cfg Config

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("failed to read config from file: %v", err)
		}
		log.Printf("Config loaded from file: %s", configPath)
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("failed to read config from env: %v", err)
		}
		log.Printf("Config loaded from environment variables")
	}

	address := fmt.Sprintf("0.0.0.0:%s", cfg.GRPCPort)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	petnamepb.RegisterPetnameGeneratorServer(s, &server{})
	reflection.Register(s)

	log.Printf("Petname gRPC server starting on %s", address)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
