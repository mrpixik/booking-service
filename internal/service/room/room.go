package room

import (
	"context"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
)

//go:generate mockery --name=roomRepo --inpackage --testonly
type roomRepo interface {
	GetAll(ctx context.Context) ([]model.Room, error)
	Create(ctx context.Context, room *model.Room) (*model.Room, error)
}

type RoomService struct {
	repo roomRepo
}

func NewRoomService(repo roomRepo) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) List(ctx context.Context) (*dto.RoomListResponse, error) {
	rooms, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, adapters.ErrUnwrapRepoToService(err)
	}

	result := make([]dto.RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		result = append(result, dto.RoomResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Capacity:    r.Capacity,
			CreatedAt:   r.CreatedAt,
		})
	}

	return &dto.RoomListResponse{Rooms: result}, nil
}

func (s *RoomService) Create(ctx context.Context, req dto.RoomCreateRequest) (*dto.RoomResponse, error) {
	if req.Name == "" {
		return nil, service.ErrInvalidRequest
	}
	if req.Capacity != nil && *req.Capacity <= 0 {
		return nil, service.ErrInvalidRequest
	}

	room := &model.Room{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
	}

	created, err := s.repo.Create(ctx, room)
	if err != nil {
		return nil, adapters.ErrUnwrapRepoToService(err)
	}

	return &dto.RoomResponse{
		ID:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		Capacity:    created.Capacity,
		CreatedAt:   created.CreatedAt,
	}, nil
}
