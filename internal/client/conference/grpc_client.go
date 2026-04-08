package conference

import (
	"context"
	"fmt"
	"time"

	pb "github.com/avito-internships/test-backend-1-mrpixik/api/proto/conference"
)

type GRPCClient struct {
	client pb.ConferenceServiceClient
}

func (c *GRPCClient) CreateConference(ctx context.Context, bookingID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.client.CreateConference(ctx, &pb.CreateConferenceRequest{BookingId: bookingID})
	if err != nil {
		return "", fmt.Errorf("conference service: %w", err)
	}
	return resp.ConferenceLink, nil
}
