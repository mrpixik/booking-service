package conference

import (
	"context"

	pb "github.com/avito-internships/test-backend-1-mrpixik/api/proto/conference"
)

type fakeServer struct {
	pb.UnimplementedConferenceServiceServer
}

func (f *fakeServer) CreateConference(ctx context.Context, req *pb.CreateConferenceRequest) (*pb.CreateConferenceResponse, error) {
	return &pb.CreateConferenceResponse{
		ConferenceLink: "https://avito-ktalk.com/" + req.BookingId,
	}, nil
}
