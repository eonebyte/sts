package handover

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
)

type Service interface {
	ProcessInit(ctx context.Context, req HandoverRequest) error
	ProcessHandover(ctx context.Context, req HandoverRequest) error
}

type service struct {
	repo         Repository
	notifService NotificationService
}

func NewService(r Repository, n NotificationService) Service {
	return &service{repo: r, notifService: n}
}

func (s *service) generateHandoverPdf(bundleNo string, req HandoverRequest, details []HandoverNotifyDTO, actors *BundleActorDTO) (string, error) {
	// 1. Path Setup
	uploadDir := "uploads/handover"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create dir: %v", err)
	}

	uniqueId := uuid.New().String()
	fileName := fmt.Sprintf("handover_%s.pdf", uniqueId)
	filePath := filepath.Join(uploadDir, fileName)

	// 2. Init PDF (A4, Unit: mm)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Header - Title
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "LIST HANDOVER", "", 1, "C", false, 0, "")

	// Sub-Header
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Status: %s", req.Status), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Bundle No: %s", bundleNo), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// 3. Draw Table Header
	pdf.SetFillColor(230, 230, 230)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(10, 8, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 8, "Customer", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 8, "Shipment No", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, "Movement Date", "1", 1, "L", true, 0, "")

	// 4. Draw Table Body
	pdf.SetFont("Arial", "", 8)
	for i, item := range details {
		// Auto add page if content exceeds
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// Redraw header on new page
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(10, 8, "No", "1", 0, "C", true, 0, "")
			pdf.CellFormat(60, 8, "Customer", "1", 0, "L", true, 0, "")
			pdf.CellFormat(70, 8, "Shipment No", "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 8, "Movement Date", "1", 1, "L", true, 0, "")
			pdf.SetFont("Arial", "", 8)
		}

		pdf.CellFormat(10, 7, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(60, 7, item.CustomerName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(70, 7, item.DocumentNo, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 7, item.MovementDate, "1", 1, "L", false, 0, "")
	}

	pdf.Ln(8)

	// 5. Layout: [Penyerah] [QR Code] [Penerima] dalam satu baris
	currentY := pdf.GetY()
	pageWidth := 210.0 // A4 width in mm
	margin := 10.0
	contentWidth := pageWidth - (2 * margin) // 190mm

	// QR Code Generation (ukuran lebih kecil)
	qrFileName := fmt.Sprintf("temp_qr_%s.png", uniqueId)
	qrPath := filepath.Join(uploadDir, qrFileName)
	qrContent := fmt.Sprintf("https://yourdomain.com/handover/%s", fileName)
	qrSize := 20.0 // Ukuran QR code 20mm (lebih kecil dari 30mm sebelumnya)

	if errQr := qrcode.WriteFile(qrContent, qrcode.Medium, 256, qrPath); errQr == nil {
		// QR code di tengah
		qrX := margin + (contentWidth-qrSize)/2
		pdf.ImageOptions(qrPath, qrX, currentY, qrSize, 0, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		defer os.Remove(qrPath)
	}

	// Kolom Penyerah (Kiri)
	colWidth := (contentWidth - qrSize) / 2 // Lebar kolom kiri/kanan

	pdf.SetXY(margin, currentY)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(colWidth, 5, "Diserahkan Oleh:", "", 1, "C", false, 0, "")

	pdf.SetX(margin)
	pdf.Ln(10)
	pdf.SetFont("Arial", "U", 9)
	pdf.SetX(margin)
	pdf.CellFormat(colWidth, 5, actors.PrevActorName, "", 1, "C", false, 0, "")

	pdf.SetX(margin)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(colWidth, 5, actors.HandoverTime+" WIB", "", 0, "C", false, 0, "")

	// Kolom Penerima (Kanan)
	rightColX := margin + colWidth + qrSize

	pdf.SetXY(rightColX, currentY)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(colWidth, 5, "Diterima Oleh:", "", 1, "C", false, 0, "")

	pdf.SetX(rightColX)
	pdf.Ln(10)
	pdf.SetFont("Arial", "U", 9)
	pdf.SetX(rightColX)
	pdf.CellFormat(colWidth, 5, actors.CurrentActorName, "", 1, "C", false, 0, "")

	pdf.SetX(rightColX)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(colWidth, 5, actors.ReceiveTime+" WIB", "", 0, "C", false, 0, "")

	// 6. Save File
	err := pdf.OutputFileAndClose(filePath)
	return filePath, err
}

func (s *service) ProcessInit(ctx context.Context, req HandoverRequest) error {
	if len(req.MInOutIDs) == 0 {
		return errors.New("minimal satu Surat Jalan harus dipilih")
	}
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var entities []TrackingSJ
	for _, id := range req.MInOutIDs {
		entities = append(entities, TrackingSJ{
			ClientID:  AdClientID,
			OrgID:     AdOrgID,
			MInOutID:  id,
			Status:    req.Status,
			CreatedBy: req.UserID,
		})
	}

	// 1. Buat STS dan Log Event (Logic Batch yang sudah ada)
	if err := s.repo.CreateBatch(ctx, tx, entities, req.Status, req.Notes); err != nil {
		return err // Otomatis trigger rollback lewat defer
	}

	// 2. Karena Init adalah proses batch, kita buatkan Bundle-nya
	// Kita perlu ambil data STS yang baru saja di-insert untuk mendapatkan ADW_STS_ID-nya
	newSTSMap, err := s.repo.GetByMInOutIDs(ctx, tx, req.MInOutIDs)
	if err != nil {
		return fmt.Errorf("bundle failed: cannot fetch new STS IDs: %w", err)
	}

	var stsIDs []int64
	for _, sts := range newSTSMap {
		stsIDs = append(stsIDs, sts.ID)
	}

	// Simpan Bundle (Header & Lines)
	return tx.Commit()
}

func (s *service) ProcessHandover(ctx context.Context, req HandoverRequest) error {
	var oldDataList []TrackingSJ
	var err error

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if req.Status == "HO: DRIVER_CHECKOUT" && len(req.MInOutIDs) == 0 {
		// 1. Catat log aktivitas ke event (tanpa update table ADW_STS)
		err = s.repo.LogActivityOnly(ctx, tx, req)
		if err != nil {
			return fmt.Errorf("gagal mencatat aktivitas checkout: %w", err)
		}

		// 2. Kirim Notifikasi Sederhana secara Async
		go s.sendSimpleCheckoutNotification(req)

		return nil // Berhenti di sini karena tidak ada SJ yang diproses
	}

	// 1. Tentukan sumber data
	if req.Status == "HO: DRIVER_CHECKIN" {
		// Tarik dari DB berdasarkan Driver ID
		oldDataList, err = s.repo.GetByCustomerIDDriverID(ctx, req.CurrentCustomer, req.DriverBy)
		if err != nil {
			return err
		}
		if len(oldDataList) == 0 {
			return fmt.Errorf("tidak ada Surat Jalan aktif untuk Driver ID %d", req.DriverBy)
		}
	} else {
		// Skenario normal: Tarik berdasarkan m_inout_ids dari request
		if len(req.MInOutIDs) == 0 {
			return errors.New("minimal satu Surat Jalan harus dipilih")
		}

		oldDataMap, err := s.repo.GetByMInOutIDs(ctx, tx, req.MInOutIDs)
		if err != nil {
			fmt.Printf("[GetByMInOutIDs]: error: %v\n", err)
			return err
		}

		for _, id := range req.MInOutIDs {
			data, exists := oldDataMap[id]
			if !exists {
				return fmt.Errorf("SJ ID %d belum di-proses INIT", id)
			}
			oldDataList = append(oldDataList, data)
		}
	}

	// 2. Proses data yang sudah terkumpul (baik dari ID maupun dari Driver)
	var entities []TrackingSJ
	var stsIDs []int64
	var mInOutIDs []int64

	for _, oldData := range oldDataList {
		var finalTNKB, finalDriver, finalCurrentCustomer *int64
		var prevActor, currentActor int64
		var prevCreated time.Time

		switch req.Status {
		case "HO: DEL_TO_DPK",
			"RE: DPK_FROM_DEL",
			"RE: DPK_FROM_DRIVER",
			"HO: DPK_TO_DEL",
			"RE: DEL_FROM_DPK",
			"HO: DEL_TO_MKT",
			"RE: MKT_FROM_DEL",
			"HO: MKT_TO_FAT",
			"RE: FAT_FROM_MKT":
			finalTNKB = oldData.TNKBID
			finalDriver = oldData.DriverBy
			prevActor = oldData.UpdatedBy
			prevCreated = oldData.CreatedAt
			currentActor = req.UserID

		case "HO: DPK_TO_DRIVER":
			// Untuk Check-in, gunakan TNKB dan Driver dari Request
			newTNKB := req.TNKBID
			newDriver := req.DriverBy

			finalTNKB = &newTNKB
			finalDriver = &newDriver
			prevActor = oldData.UpdatedBy
			prevCreated = oldData.CreatedAt
			currentActor = req.UserID // Aktornya adalah Driver itu sendiri

		case "HO: DRIVER_CHECKIN":
			// Untuk Check-in, gunakan TNKB dan Driver dari Request
			newTNKB := req.TNKBID
			newDriver := req.DriverBy
			newCurrentCustomer := req.CurrentCustomer

			finalTNKB = &newTNKB
			finalDriver = &newDriver
			finalCurrentCustomer = &newCurrentCustomer
			prevActor = oldData.UpdatedBy
			prevCreated = oldData.CreatedAt
			currentActor = req.UserID // Aktornya adalah Driver itu sendiri

		case "HO: DRIVER_CHECKOUT":
			// Untuk Check-in, gunakan TNKB dan Driver dari Request
			newTNKB := req.TNKBID
			newDriver := req.DriverBy
			newCurrentCustomer := oldData.CurrentCustomer

			finalTNKB = &newTNKB
			finalDriver = &newDriver
			finalCurrentCustomer = newCurrentCustomer
			prevActor = oldData.UpdatedBy
			prevCreated = oldData.CreatedAt
			currentActor = req.UserID // Aktornya adalah Driver itu sendiri

		default:
			newTNKB := req.TNKBID
			newDriver := req.DriverBy
			finalTNKB = &newTNKB
			finalDriver = &newDriver
			prevActor = oldData.UpdatedBy
			prevCreated = oldData.CreatedAt
			currentActor = req.UserID
		}

		entities = append(entities, TrackingSJ{
			ID:              oldData.ID,
			ClientID:        AdClientID,
			OrgID:           AdOrgID,
			MInOutID:        oldData.MInOutID,
			TNKBID:          finalTNKB,
			Status:          req.Status,
			DriverBy:        finalDriver,
			CurrentCustomer: finalCurrentCustomer,
			UpdatedBy:       currentActor,
			PrevActorID:     prevActor,
			CreatedAt:       prevCreated,
		})
		stsIDs = append(stsIDs, oldData.ID)
		mInOutIDs = append(mInOutIDs, oldData.MInOutID)
	}

	err = s.repo.UpdateBatch(ctx, tx, entities, req.Status, req.Notes)
	if err != nil {
		return err
	}

	// 3. Logika Pembuatan Bundle & Persiapan PDF
	var bundleDocNo string // ← Simpan DocumentNo di sini
	var shouldGeneratePDF bool

	if req.Status == "RE: DPK_FROM_DEL" ||
		req.Status == "RE: DEL_FROM_DPK" ||
		req.Status == "RE: MKT_FROM_DEL" ||
		req.Status == "RE: FAT_FROM_MKT" {

		shouldGeneratePDF = true

		prefix := getPrefixForStatus(req.Status)
		bundleDocNo = fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()/1e6)

		bundleHeader := ADWBundle{
			DocumentNo:  bundleDocNo, // ← Gunakan variable yang sama
			BundleType:  req.Status,
			Description: req.Notes,
			CreatedBy:   req.UserID,
			Attachment:  "",
		}

		if errB := s.repo.CreateBundle(ctx, tx, bundleHeader, stsIDs); errB != nil {
			return fmt.Errorf("gagal membuat bundle penerimaan: %w", errB)
		}
	}

	// 4. Commit Transaksi
	if err := tx.Commit(); err != nil {
		return err
	}

	// 5. G5. Generate PDF & Update Attachment
	if shouldGeneratePDF {
		capturedIDs := make([]int64, len(mInOutIDs))
		copy(capturedIDs, mInOutIDs)
		capturedReq := req
		capturedDocNo := bundleDocNo

		go func() {
			fmt.Printf("[PDF-DEBUG]: Querying IDs=%v\n", capturedIDs)

			// A. Ambil detail SJ
			details, errD := s.repo.GetNotificationDetails(context.Background(), capturedIDs, 0)
			if errD != nil {
				fmt.Printf("[PDF-ERR]: Gagal ambil detail: %v\n", errD)
				return
			}

			if len(details) == 0 {
				fmt.Printf("[PDF-WARN]: Tidak ada detail untuk IDs: %v\n", capturedIDs)
				return
			}

			// B. Ambil info Penyerah & Penerima
			actors, errActor := s.repo.GetBundleActors(context.Background(), capturedDocNo)
			if errActor != nil {
				fmt.Printf("[PDF-WARN]: Gagal ambil actor info: %v\n", errActor)
				// Set default jika gagal
				actors = &BundleActorDTO{
					PrevActorName:    "N/A",
					CurrentActorName: "N/A",
					EventType:        capturedReq.Status,
					HandoverTime:     "N/A",
					ReceiveTime:      time.Now().Format("02-01-2006 15:04"),
				}
			}

			// C. Generate PDF dengan info actors
			filePath, errPdf := s.generateHandoverPdf(capturedDocNo, capturedReq, details, actors)
			if errPdf != nil {
				fmt.Printf("[PDF-ERR]: Gagal generate file: %v\n", errPdf)
				return
			}

			fmt.Printf("[PDF-OK]: File tersimpan di %s\n", filePath)

			// D. Update attachment di database
			errUpdate := s.repo.UpdateBundleAttachment(context.Background(), capturedDocNo, filePath)
			if errUpdate != nil {
				fmt.Printf("[PDF-ERR]: Gagal update attachment: %v\n", errUpdate)
			} else {
				fmt.Printf("[PDF-OK]: Attachment updated untuk bundle %s\n", capturedDocNo)
			}
		}()
	}

	// 6. Kirim Notifikasi WA (Logic tetap sama)
	if req.Status == "HO: DRIVER_CHECKIN" || req.Status == "HO: DRIVER_CHECKOUT" {
		// Ambil semua M_INOUT_ID dari entities untuk di-query detailnya
		var ids []int64
		for _, e := range entities {
			ids = append(ids, e.MInOutID)
		}

		go func() {
			// A. Ambil Data Detail (Join) dari Database
			details, err := s.repo.GetNotificationDetails(context.Background(), ids, req.DriverBy)
			if err != nil {
				fmt.Printf("[WA-ERROR]: Gagal ambil detail: %v\n", err)
				return
			}

			if len(details) == 0 {
				return
			}

			var titleDisplay string
			if req.Status == "HO: DRIVER_CHECKIN" {
				titleDisplay = "Driver Check-In"
			} else {
				titleDisplay = "Driver Check-Out"
			}

			// B. Susun Pesan WA yang Cantik
			driverName := details[0].DriverName // Ambil dari baris pertama
			checkinTime := details[0].Time
			tnkbName := details[0].TNKB
			customerName := details[0].CustomerName

			msg := fmt.Sprintf("*%s*\n\n", titleDisplay)
			msg += fmt.Sprintf("Driver: *%s*\n", driverName)
			msg += fmt.Sprintf("TNKB: *%s*\n", tnkbName)
			msg += fmt.Sprintf("Customer: *%s*\n", customerName)
			msg += fmt.Sprintf("Waktu   : *%s*\n", checkinTime) // Menggunakan field Time
			msg += fmt.Sprintf("Catatan: %s\n\n", req.Notes)
			msg += "*Daftar Surat Jalan:*\n"

			for i, d := range details {
				msg += fmt.Sprintf("%d. *%s* - %s\n", i+1, d.DocumentNo, d.CustomerName)
			}
			msg += fmt.Sprintf("\n_Total: %d Surat Jalan_", len(details))

			// C. Kirim via NotifService (pastikan ada method SendCustomMessage)
			// Atau sesuaikan dengan method SendDriverCheckinNotification Anda
			errNotif := s.notifService.SendDriverCheckinNotification(context.Background(), msg)
			if errNotif != nil {
				fmt.Printf("[WA-ERROR]: %v\n", errNotif)
			}
		}()
	}

	return nil
}

func (s *service) sendSimpleCheckoutNotification(req HandoverRequest) {
	// 1. Ambil detail nama driver & customer dari DB
	details, err := s.repo.GetNotifLogActivityOnlyDetail(context.Background(), req.CurrentCustomer, req.DriverBy)

	var driverName, customerName, currentTime string
	if err != nil || len(details) == 0 {
		// Fallback jika query gagal
		driverName = fmt.Sprintf("ID: %d", req.DriverBy)
		customerName = fmt.Sprintf("ID: %d", req.CurrentCustomer)
		currentTime = "Baru saja"
	} else {
		driverName = details[0].DriverName
		customerName = details[0].CustomerName
		currentTime = details[0].Time
	}

	// 2. Susun pesan WA yang lebih informatif
	msg := "*Driver Check-Out (Tanpa SJ)*\n\n"
	msg += fmt.Sprintf("Driver: *%s*\n", driverName)
	msg += fmt.Sprintf("Lokasi: *%s*\n", customerName)
	msg += fmt.Sprintf("Waktu : *%s*\n", currentTime)
	msg += fmt.Sprintf("Catatan: %s\n\n", req.Notes)
	msg += "_Keterangan: Driver telah meninggalkan lokasi customer tanpa membawa kembali dokumen Surat Jalan._"

	// 3. Kirim
	errNotif := s.notifService.SendDriverCheckinNotification(context.Background(), msg)
	if errNotif != nil {
		fmt.Printf("[WA-ERROR]: %v\n", errNotif)
	}
}

// Helper function
func getPrefixForStatus(status string) string {
	switch status {
	case "RE: DPK_FROM_DEL":
		return "HOPT"
	case "RE: DEL_FROM_DPK":
		return "HITP"
	case "RE: MKT_FROM_DEL":
		return "HIPM"
	case "RE: FAT_FROM_MKT":
		return "HIMF"
	default:
		return "RECV"
	}
}
