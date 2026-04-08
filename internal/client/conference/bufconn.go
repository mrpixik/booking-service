package conference

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/avito-internships/test-backend-1-mrpixik/api/proto/conference"
)

const bufSize = 1024 * 1024

func NewBufconnClient() (*GRPCClient, func(), error) {

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()

	pb.RegisterConferenceServiceServer(srv, &fakeServer{})
	go srv.Serve(lis)

	conn, err := grpc.NewClient("passthrough://fake",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		srv.Stop()
		return nil, nil, err
	}

	// функция для gracefull shutdown
	cleanup := func() {
		_ = conn.Close()
		srv.Stop()
	}

	return &GRPCClient{
		client: pb.NewConferenceServiceClient(conn),
	}, cleanup, nil
}
