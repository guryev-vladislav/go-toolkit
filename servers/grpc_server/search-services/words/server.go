package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	wordspb "github.com/guryev-vladislav/go-toolkit/servers/grpc_server/search-services/proto/words"
	"github.com/guryev-vladislav/go-toolkit/servers/grpc_server/search-services/words/words"
	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
	wordspb.UnimplementedWordsServer
}

type Config struct {
	GRPCPort string `yaml:"grpc_port" env:"WORDS_GRPC_PORT" env-default:"28082"`
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

const maxMessageSize = 4 * 1024

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.Phrase) > maxMessageSize {
		return nil, status.Errorf(codes.ResourceExhausted,
			"message size %d bytes exceeds maximum allowed %d bytes",
			len(in.Phrase), maxMessageSize)
	}

	return &wordspb.WordsReply{
		Words: words.Norm(in.GetPhrase()),
	}, nil

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
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	log.Printf("Words gRPC server starting on %s", address)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
