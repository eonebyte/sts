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

	CancelOutstanding(ctx context.Context, id int64, currentStatus string) error
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
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetPending(dateFrom, dateTo)
}

func (s *service) GetHistory(fromStr, toStr string) ([]ShipmentHistory, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetHistory(dateFrom, dateTo)
}

func (s *service) GetPrepare(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetPrepare(dateFrom, dateTo)
}

func (s *service) GetPrepareToLeave(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetPrepareToLeave(dateFrom, dateTo)
}

func (s *service) GetInTransit(driverID int64) ([]Shipment, error) {
	now := time.Now()
	loc := now.Location()

	// Ambil awal bulan ini jam 00:00:00
	dateFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	// Ambil sampai besok (agar data hari ini jam berapapun masuk)
	dateTo := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)

	return s.repo.GetInTransitCustomer(dateFrom, dateTo, driverID)
}

func (s *service) GetOnCustomer(customerID, driverID int64) ([]Shipment, error) {
	now := time.Now()
	loc := now.Location()

	// Ambil awal bulan ini jam 00:00:00
	dateFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	// Ambil sampai besok (agar data hari ini jam berapapun masuk)
	dateTo := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)

	return s.repo.GetOnCustomer(dateFrom, dateTo, customerID, driverID)
}

func (s *service) GetComeback(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetComeback(dateFrom, dateTo)
}

func (s *service) GetComebackToDelivery(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetComebackToDelivery(dateFrom, dateTo)
}

func (s *service) GetReceiptComebackToDelivery(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetReceiptComebackToDelivery(dateFrom, dateTo)
}

func (s *service) GetComebackToMarketing(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetComebackToMarketing(dateFrom, dateTo)
}

func (s *service) GetReceiptComebackToMarketing(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetReceiptComebackToMarketing(dateFrom, dateTo)
}

func (s *service) GetComebackToFat(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetComebackToFat(dateFrom, dateTo)
}

func (s *service) GetReceiptComebackToFat(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetReceiptComebackToFat(dateFrom, dateTo)
}

func (s *service) GetOutstandingDPK(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetOutstandingDPK(dateFrom, dateTo)
}

func (s *service) GetOutstandingDelivery(fromStr, toStr string) ([]Shipment, error) {
	var dateFrom, dateTo time.Time
	now := time.Now()
	loc := now.Location()

	// 1. Logika untuk FROM (Default: Tanggal 1 bulan berjalan)
	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err == nil {
			dateFrom = time.Date(parsedFrom.Year(), parsedFrom.Month(), parsedFrom.Day(), 0, 0, 0, 0, loc)
		} else {
			// Jika parsing gagal, fallback ke awal bulan
			dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		}
	} else {
		// DEFAULT: Tanggal 1 bulan ini jam 00:00:00
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	// 2. Logika untuk TO (Default: Tanggal 1 bulan depan / Akhir bulan ini + 1 hari)
	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err == nil {
			// Ditambah 1 hari agar mencakup data sampai akhir hari yang dipilih
			dateTo = time.Date(parsedTo.Year(), parsedTo.Month(), parsedTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
		} else {
			dateTo = dateFrom.AddDate(0, 1, 0) // Jika salah format, tambah 1 bulan dari dateFrom
		}
	} else {
		if fromStr == "" {
			// Jika dua-duanya kosong, maka range adalah "Bulan Ini"
			dateTo = dateFrom.AddDate(0, 1, 0)
		} else {
			// Jika hanya toStr yang kosong, ambil range 1 hari dari dateFrom
			dateTo = dateFrom.AddDate(0, 0, 1)
		}
	}

	return s.repo.GetOutstandingDelivery(dateFrom, dateTo)
}

func (s *service) CancelOutstanding(ctx context.Context, id int64, currentStatus string) error {
	var nextStatus string
	var isHardDelete bool

	// Business Logic: State Machine Pembatalan
	switch currentStatus {
	case "HO: DEL_TO_DPK":
		// Jika masih di tahap awal, hapus record agar bisa diulang dari nol
		isHardDelete = true

	case "HO: DPK_TO_DRIVER":
		nextStatus = "RE: DPK_FROM_DEL"
		isHardDelete = false

	case "HO: DPK_TO_DEL":
		nextStatus = "RE: DPK_FROM_DRIVER"
		isHardDelete = false

	case "HO: DEL_TO_MKT":
		nextStatus = "RE: DEL_FROM_DPK"
		isHardDelete = false

	case "HO: MKT_TO_FAT":
		nextStatus = "RE: MKT_FROM_DEL"
		isHardDelete = false

	default:
		return fmt.Errorf("status '%s' tidak valid untuk dibatalkan melalui sistem ini", currentStatus)
	}

	// Eksekusi ke Repository
	return s.repo.ExecuteOutstandingCancel(ctx, id, nextStatus, isHardDelete)
}
