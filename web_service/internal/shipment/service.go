package shipment

import (
	"context"
	"fmt"
	"time"
)

type Service interface {
	GetAllCustomers() ([]Customer, error)
	GetDriver() ([]Driver, error)
	GetTnkb() ([]TNKB, error)
	GetPending(fromStr, toStr string) ([]Shipment, error)
	GetPrepare(fromStr, toStr string) ([]Shipment, error)
	GetPrepareToLeave(fromStr, toStr string) ([]Shipment, error)
	GetInTransit(driverID int64) ([]Shipment, error)
	GetOnCustomer(customerID, driverID int64) ([]Shipment, error)

	GetComeback(fromStr, toStr string) ([]Shipment, error)
	GetComebackToDelivery(fromStr, toStr string) ([]Shipment, error)
	GetReceiptComebackToDelivery(fromStr, toStr string) ([]Shipment, error)

	GetComebackToMarketing(fromStr, toStr string) ([]Shipment, error)
	GetReceiptComebackToMarketing(fromStr, toStr string) ([]Shipment, error)

	GetComebackToFat(fromStr, toStr string) ([]Shipment, error)
	GetReceiptComebackToFat(fromStr, toStr string) ([]Shipment, error)

	FetchProgress(ctx context.Context) ([]ShipmentProgress, error)
	GetHistory(fromStr, toStr string) ([]ShipmentHistory, error)

	GetOutstandingDPK(fromStr, toStr string) ([]Shipment, error)
	GetOutstandingDelivery(fromStr, toStr string) ([]Shipment, error)

	UpdateDriverTnkb(ctx context.Context, inoutID int64, driverID int64, tnkbID int64) error
}

type service struct {
	repo Repository
}

func NewService(r Repository) Service {
	return &service{repo: r}
}

func (s *service) UpdateDriverTnkb(ctx context.Context, inoutID int64, driverID int64, tnkbID int64) error {
	// Validasi bisnis tambahan (opsional)
	if inoutID <= 0 {
		return fmt.Errorf("invalid M_INOUT_ID")
	}

	// Meneruskan semua parameter ke repository
	return s.repo.UpdateDriverTnkb(ctx, inoutID, driverID, tnkbID)
}

func (s *service) FetchProgress(ctx context.Context) ([]ShipmentProgress, error) {
	return s.repo.GetDailyProgress(ctx)
}

func (s *service) GetHistory(fromStr, toStr string) ([]ShipmentHistory, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()

	// Default 7 hari terakhir jika tanggal kosong
	if fromStr == "" {
		dateFrom = now.AddDate(0, 0, -7)
	} else {
		parsedFrom, _ := time.Parse("2006-01-02", fromStr)
		dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
	}

	if toStr == "" {
		dateTo = now.AddDate(0, 0, 1)
	} else {
		parsedTo, _ := time.Parse("2006-01-02", toStr)
		dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	}

	return s.repo.GetHistory(dateFrom, dateTo)
}

func (s *service) GetAllCustomers() ([]Customer, error) {
	return s.repo.GetAllCustomers()
}

func (s *service) GetDriver() ([]Driver, error) {
	return s.repo.GetDriver()
}

func (s *service) GetTnkb() ([]TNKB, error) {
	return s.repo.GetTnkb()
}

func (s *service) GetPending(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetPending(dateFrom, dateTo)
}

func (s *service) GetPrepare(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetPrepare(dateFrom, dateTo)
}

func (s *service) GetPrepareToLeave(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetPrepareToLeave(dateFrom, dateTo)
}

func (s *service) GetInTransit(driverID int64) ([]Shipment, error) {
	// Gunakan range waktu default (misal: 3 hari kebelakang sampai hari ini)
	now := time.Now()
	from := now.AddDate(0, 0, -3)
	to := now.AddDate(0, 0, 1)

	return s.repo.GetInTransitCustomer(from, to, driverID)
}

func (s *service) GetOnCustomer(customerID, driverID int64) ([]Shipment, error) {
	// Gunakan range waktu default (misal: 3 hari kebelakang sampai hari ini)
	now := time.Now()
	from := now.AddDate(0, 0, -3)
	to := now.AddDate(0, 0, 1)

	return s.repo.GetOnCustomer(from, to, customerID, driverID)
}

func (s *service) GetComeback(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetComeback(dateFrom, dateTo)
}

func (s *service) GetComebackToDelivery(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetComebackToDelivery(dateFrom, dateTo)
}

func (s *service) GetReceiptComebackToDelivery(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetReceiptComebackToDelivery(dateFrom, dateTo)
}

func (s *service) GetComebackToMarketing(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetComebackToMarketing(dateFrom, dateTo)
}

func (s *service) GetReceiptComebackToMarketing(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetReceiptComebackToMarketing(dateFrom, dateTo)
}

func (s *service) GetComebackToFat(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetComebackToFat(dateFrom, dateTo)
}

func (s *service) GetReceiptComebackToFat(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetReceiptComebackToFat(dateFrom, dateTo)
}

func (s *service) GetOutstandingDPK(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetOutstandingDPK(dateFrom, dateTo)
}

func (s *service) GetOutstandingDelivery(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	// Hari ini jam 00:00:00 sebagai default awal
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, now.Location())
		} else {
			dateFrom = todayStart
		}
	} else {
		dateFrom = todayStart
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup jam 23:59:59 pada hari tersebut
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
		} else {
			// Jika format salah, ambil sampai besok (asumsi filter 1 hari)
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	} else {
		// Jika dateTo kosong, default-nya adalah 1 hari setelah dateFrom
		dateTo = dateFrom.AddDate(0, 0, 1)
	}

	return s.repo.GetOutstandingDelivery(dateFrom, dateTo)
}
