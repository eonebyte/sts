package tms

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Service interface {
	SearchDriver(ctx context.Context, searchKey string) ([]SearchDriver, error)
	ShipmentByDriver(ctx context.Context, driverID int64) ([]ShipmentByDriver, error)
	GetCustomerLogs(ctx context.Context, tmsID int64) ([]CustomerLog, error)
	UpdateLog(ctx context.Context, eventID int64, rawTime string, notes string) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) SearchDriver(ctx context.Context, searchKey string) ([]SearchDriver, error) {
	decodedKey, err := url.QueryUnescape(searchKey)

	if err != nil {
		decodedKey = searchKey
	}

	cleanKey := strings.TrimSpace(decodedKey)

	drivers, err := s.repo.GetDriverByName(ctx, cleanKey)
	if err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	return drivers, nil

}

func (s *service) ShipmentByDriver(ctx context.Context, driverID int64) ([]ShipmentByDriver, error) {
	shipments, err := s.repo.ShipmentByDriver(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	return shipments, nil

}

func (s *service) GetCustomerLogs(ctx context.Context, tmsID int64) ([]CustomerLog, error) {
	return s.repo.GetLogsByTMS(ctx, tmsID)
}

func (s *service) UpdateLog(ctx context.Context, eventID int64, rawTime string, notes string) error {
	// Parsing format datetime-local HTML (ISO 8601 tanpa detik)
	// Layout: 2006-01-02T15:04
	parsedTime, err := time.Parse("2006-01-02T15:04", rawTime)
	if err != nil {
		return err // Kirim error jika format waktu salah
	}

	timeStrForDB := parsedTime.Format("2006-01-02 15:04:05")

	// Teruskan ke repository
	return s.repo.UpdateEventLog(ctx, eventID, timeStrForDB, notes)
}
