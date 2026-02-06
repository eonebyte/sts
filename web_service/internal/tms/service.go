package tms

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type Service interface {
	SearchDriver(ctx context.Context, searchKey string) ([]SearchDriver, error)
	ShipmentByDriver(ctx context.Context, driverID int64) ([]ShipmentByDriver, error)
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
