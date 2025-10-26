package app

import (
	"fmt"
	"net"
	"os"

	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
	"github.com/mystaline/clefinport-be/pkg/provider"
	"github.com/mystaline/clefinport-be/services/wallet_service/internal/route"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func RunGRPCServer(
	serviceProvider provider.IServiceProvider,
) error {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb_wallet.RegisterWalletServiceServer(s, route.SetupWalletGRPC(serviceProvider))

	reflection.Register(s)

	fmt.Println("🚀 gRPC Wallet server running on port", grpcPort)
	return s.Serve(lis)
}
