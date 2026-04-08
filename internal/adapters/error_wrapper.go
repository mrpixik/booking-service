package adapters

import (
	"encoding/json"
	"net/http"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/repository"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/server"
	"github.com/avito-internships/test-backend-1-mrpixik/internal/domain/dto/errors/service"
)

// Написал этот функционал, чтобы не передавать ошибки с уровня репозитория наверх к уровню контроллеров
// и чтобы избежать огромных структур if/else в сервисном слое при проверке соответствия ошибок.
// Я понимаю, что в целом, передавать можно и так делают в проде, и в высоконагруженных сервисах это оправданное решение,
// но так как это тестовое задание, решил сделать все максимально правильно с точки чистой архитектуры
// P.S. Возможно это можно сделать покрасивее или получше если использовать фабрику или что-то такое, но пока решил сделать так.
var repoToServiceMap = map[error]error{
	repository.ErrEmailExists:     service.ErrEmailTaken,
	repository.ErrScheduleExists:  service.ErrScheduleExists,
	repository.ErrRoomNotFound:    service.ErrRoomNotFound,
	repository.ErrSlotNotFound:    service.ErrSlotNotFound,
	repository.ErrBookingNotFound: service.ErrBookingNotFound,
	repository.ErrBookingExists:   service.ErrSlotAlreadyBooked,
	repository.ErrNotFound:        service.ErrNotFound,
	repository.ErrInternalError:   service.ErrInternalError,
}

func ErrUnwrapRepoToService(err error) error {
	if err == nil {
		return nil
	}

	if mapped, ok := repoToServiceMap[err]; ok {
		return mapped
	}

	return service.ErrInternalError
}

var serviceErrorMap = map[error]server.ErrorResponse{
	// Auth
	service.ErrEmptyCredentials: {
		Error:  server.ErrorDetail{Code: server.CodeInvalidRequest, Message: server.EmptyCredentialsMsg},
		Status: http.StatusBadRequest,
	},
	service.ErrEmailTaken: {
		Error:  server.ErrorDetail{Code: server.CodeInvalidRequest, Message: server.EmailTakenMsg},
		Status: http.StatusBadRequest,
	},
	service.ErrInvalidCredentials: {
		Error:  server.ErrorDetail{Code: server.CodeUnauthorized, Message: server.InvalidCredentialsMsg},
		Status: http.StatusUnauthorized,
	},

	// Room
	service.ErrRoomNotFound: {
		Error:  server.ErrorDetail{Code: server.CodeRoomNotFound, Message: server.RoomNotFoundMsg},
		Status: http.StatusNotFound,
	},

	// Schedule
	service.ErrScheduleExists: {
		Error:  server.ErrorDetail{Code: server.CodeScheduleExists, Message: server.ScheduleExistsMsg},
		Status: http.StatusConflict,
	},

	// Slot
	service.ErrSlotNotFound: {
		Error:  server.ErrorDetail{Code: server.CodeSlotNotFound, Message: server.SlotNotFoundMsg},
		Status: http.StatusNotFound,
	},

	// Booking
	service.ErrSlotAlreadyBooked: {
		Error:  server.ErrorDetail{Code: server.CodeSlotAlreadyBooked, Message: server.SlotAlreadyBookedMsg},
		Status: http.StatusConflict,
	},
	service.ErrBookingNotFound: {
		Error:  server.ErrorDetail{Code: server.CodeBookingNotFound, Message: server.BookingNotFoundMsg},
		Status: http.StatusNotFound,
	},
	service.ErrCancelBookFromAnotherUser: {
		Error:  server.ErrorDetail{Code: server.CodeForbidden, Message: server.CancelBookFromAnotherUserMsg},
		Status: http.StatusForbidden,
	},

	// Default
	service.ErrInvalidRequest: {
		Error:  server.ErrorDetail{Code: server.CodeInvalidRequest, Message: server.InvalidRequestMsg},
		Status: http.StatusBadRequest,
	},
	service.ErrNotFound: {
		Error:  server.ErrorDetail{Code: server.CodeNotFound, Message: server.NotFoundMsg},
		Status: http.StatusNotFound,
	},
	service.ErrInternalError: {
		Error:  server.ErrorDetail{Code: server.CodeInternalError, Message: server.InternalErrorMsg},
		Status: http.StatusInternalServerError,
	},
}

// WriteServiceError принимает ошибку уровня service и пишет ошибку уровня контроллера в ResponseWriter
func WriteServiceError(w http.ResponseWriter, err error) {
	resp, ok := serviceErrorMap[err]
	if !ok {
		resp = serviceErrorMap[service.ErrInternalError]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.Status)
	_ = json.NewEncoder(w).Encode(server.ErrorResponse{
		Error: resp.Error,
	})
}

func WriteError(w http.ResponseWriter, code server.ErrorCode, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(server.ErrorResponse{
		Error: server.ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
