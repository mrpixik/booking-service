package room

import (
	"context"
	"testing"
	"time"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func intPtr(v int) *int { return &v }

func TestList(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		setupRepo func(*mockRoomRepo)
		wantErr   error
		wantCount int
		wantNames []string
	}{
		{
			name: "success",
			setupRepo: func(r *mockRoomRepo) {
				r.On("GetAll", mock.Anything).Return([]model.Room{
					{ID: uuid.New().String(), Name: "Room A", CreatedAt: &now},
					{ID: uuid.New().String(), Name: "Room B", CreatedAt: &now},
				}, nil)
			},
			wantCount: 2,
			wantNames: []string{"Room A", "Room B"},
		},
		{
			name: "empty",
			setupRepo: func(r *mockRoomRepo) {
				r.On("GetAll", mock.Anything).Return([]model.Room{}, nil)
			},
			wantCount: 0,
		},
		{
			name: "repo error",
			setupRepo: func(r *mockRoomRepo) {
				r.On("GetAll", mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRoomRepo(t)
			tt.setupRepo(repo)

			svc := NewRoomService(repo)
			result, err := svc.List(context.Background())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result.Rooms, tt.wantCount)
			for i, name := range tt.wantNames {
				assert.Equal(t, name, result.Rooms[i].Name)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		req         dto.RoomCreateRequest
		setupRepo   func(*mockRoomRepo)
		wantErr     error
		checkResult func(*testing.T, *dto.RoomResponse)
	}{
		{
			name: "success with capacity",
			req:  dto.RoomCreateRequest{Name: "New Room", Capacity: intPtr(10)},
			setupRepo: func(r *mockRoomRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(&model.Room{
					ID: uuid.New().String(), Name: "New Room", Capacity: intPtr(10),
				}, nil)
			},
			checkResult: func(t *testing.T, res *dto.RoomResponse) {
				assert.Equal(t, "New Room", res.Name)
				assert.Equal(t, intPtr(10), res.Capacity)
			},
		},
		{
			name: "success nil capacity",
			req:  dto.RoomCreateRequest{Name: "Room"},
			setupRepo: func(r *mockRoomRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(&model.Room{
					ID: uuid.New().String(), Name: "Room",
				}, nil)
			},
			checkResult: func(t *testing.T, res *dto.RoomResponse) {
				assert.Nil(t, res.Capacity)
			},
		},
		{
			name:    "empty name",
			req:     dto.RoomCreateRequest{Name: ""},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "zero capacity",
			req:     dto.RoomCreateRequest{Name: "Room", Capacity: intPtr(0)},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name:    "negative capacity",
			req:     dto.RoomCreateRequest{Name: "Room", Capacity: intPtr(-5)},
			wantErr: service.ErrInvalidRequest,
		},
		{
			name: "repo error",
			req:  dto.RoomCreateRequest{Name: "Room"},
			setupRepo: func(r *mockRoomRepo) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrInternalError)
			},
			wantErr: service.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var repo *mockRoomRepo
			if tt.setupRepo != nil {
				repo = newMockRoomRepo(t)
				tt.setupRepo(repo)
			}

			svc := NewRoomService(repo)
			result, err := svc.Create(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}

			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
