package handover

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	BeginTx(ctx context.Context) (*sqlx.Tx, error) // Tambahkan ini
	CreateBundle(ctx context.Context, tx *sqlx.Tx, bundle ADWBundle, stsIDs []int64) error
	CreateBatch(ctx context.Context, tx *sqlx.Tx, entities []TrackingSJ, eventType string, notes string) error
	UpdateBatch(ctx context.Context, tx *sqlx.Tx, entities []TrackingSJ, eventType string, notes string) error

	GetByMInOutIDs(ctx context.Context, tx *sqlx.Tx, ids []int64) (map[int64]TrackingSJ, error)
	GetByCustomerIDDriverID(ctx context.Context, customerID, driverID int64) ([]TrackingSJ, error)
	GetNotificationDetails(ctx context.Context, mInOutIDs []int64, driverID int64) ([]HandoverNotifyDTO, error)
	LogActivityOnly(ctx context.Context, tx *sqlx.Tx, req HandoverRequest) error
	GetNotifLogActivityOnlyDetail(ctx context.Context, customerID, driverID int64) ([]HandoverNotifyDTO, error)

	GetBundleActors(ctx context.Context, bundleDocNo string) (*BundleActorDTO, error)

	UpdateBundleAttachment(ctx context.Context, documentNo, filePath string) error
}

type oraRepo struct {
	db *sqlx.DB
}

func NewOraRepository(db *sqlx.DB) Repository {
	return &oraRepo{db: db}
}

func (r *oraRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *oraRepo) CreateBundle(ctx context.Context, tx *sqlx.Tx, bundle ADWBundle, stsIDs []int64) error {
	// A. Logika Penanganan TX: Jika nil, buat TX baru. Jika tidak, pakai yang ada.
	useExternalTx := tx != nil
	if !useExternalTx {
		var err error
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		// Pastikan rollback jika terjadi error di tengah jalan (hanya jika kita yang buka tx)
		defer tx.Rollback()
	}

	// 1. Ambil ID Header dari Sequence
	var nextBundleID int64
	querySeq := "SELECT plastik.ADW_BUNDLE_SQ.NEXTVAL FROM DUAL"
	if err := tx.GetContext(ctx, &nextBundleID, querySeq); err != nil {
		return fmt.Errorf("gagal ambil sequence bundle: %w", err)
	}

	// 2. Insert Header dengan ATTACHMENT
	queryHeader := `
        INSERT INTO plastik.ADW_STS_BUNDLE (
            ADW_STS_BUNDLE_ID, AD_CLIENT_ID, AD_ORG_ID, DOCUMENTNO, 
            BUNDLE_TYPE, MOVEMENTDATE, DESCRIPTION, ATTACHMENT,
            CREATED, CREATEDBY, UPDATED, UPDATEDBY, ISACTIVE
        ) VALUES (:1, :2, :3, :4, :5, SYSDATE, :6, :7, SYSDATE, :8, SYSDATE, :9, 'Y')`

	_, err := tx.ExecContext(ctx, queryHeader,
		nextBundleID, AdClientID, AdOrgID, bundle.DocumentNo,
		bundle.BundleType, bundle.Description, bundle.Attachment, // ← Tambahkan attachment
		bundle.CreatedBy, bundle.CreatedBy)
	if err != nil {
		return fmt.Errorf("gagal insert bundle header: %w", err)
	}

	// 3. Insert Lines
	queryLine := `
        INSERT INTO plastik.ADW_STS_BUNDLE_LINE (
            ADW_STS_BUNDLE_LINE_ID, AD_CLIENT_ID, AD_ORG_ID, 
            ADW_STS_BUNDLE_ID, ADW_STS_ID, LINE, 
            CREATED, CREATEDBY, UPDATED, UPDATEDBY, ISACTIVE
        ) VALUES (plastik.ADW_BUNDLE_LINE_SQ.NEXTVAL, :1, :2, :3, :4, :5, SYSDATE, :6, SYSDATE, :7, 'Y')`

	for i, stsID := range stsIDs {
		_, err = tx.ExecContext(ctx, queryLine,
			AdClientID, AdOrgID, nextBundleID, stsID, (i+1)*10,
			bundle.CreatedBy, bundle.CreatedBy)
		if err != nil {
			return fmt.Errorf("gagal insert bundle line ke-%d: %w", i, err)
		}
	}

	// B. Commit jika kita yang membuka transaksi di awal fungsi ini
	if !useExternalTx {
		return tx.Commit()
	}

	return nil
}

func (r *oraRepo) UpdateBundleAttachment(ctx context.Context, documentNo, filePath string) error {
	query := `
		UPDATE plastik.ADW_STS_BUNDLE 
		SET ATTACHMENT = :1, UPDATED = SYSDATE 
		WHERE DOCUMENTNO = :2`

	_, err := r.db.ExecContext(ctx, query, filePath, documentNo)
	if err != nil {
		return fmt.Errorf("gagal update attachment: %w", err)
	}

	return nil
}

func (r *oraRepo) GetByMInOutIDs(ctx context.Context, tx *sqlx.Tx, ids []int64) (map[int64]TrackingSJ, error) {
	if len(ids) == 0 {
		return make(map[int64]TrackingSJ), nil
	}

	var executor sqlx.QueryerContext = r.db
	if tx != nil {
		executor = tx
	}

	// 1. Buat placeholder manual (:1, :2, :3...) karena sqlx Rebind gagal
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = ":" + strconv.Itoa(i+1) // Menghasilkan :1, :2, dst
		args[i] = id
	}

	// 2. Gabungkan ke dalam Query (Pastikan tidak ada backtick atau newline)
	queryStr := "SELECT ADW_STS_ID, M_INOUT_ID, TNKB_ID, DRIVERBY, UPDATEDBY, CURRENTCUSTOMER, CREATED " +
		"FROM ADW_STS " +
		"WHERE M_INOUT_ID IN (" + strings.Join(placeholders, ",") + ") " +
		"AND ISACTIVE = 'Y'"

	var list []TrackingSJ

	// 3. Eksekusi langsung tanpa Rebind lagi
	err := sqlx.SelectContext(ctx, executor, &list, queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("error database: %w", err)
	}

	result := make(map[int64]TrackingSJ)
	for _, row := range list {
		result[row.MInOutID] = row
	}
	return result, nil
}

func (r *oraRepo) GetByCustomerIDDriverID(ctx context.Context, customerID, driverID int64) ([]TrackingSJ, error) {
	var list []TrackingSJ

	// Mencari data yang aktif berdasarkan DRIVERBY
	queryStr := `SELECT 
				sts.ADW_STS_ID, sts.M_INOUT_ID, sts.TNKB_ID, sts.DRIVERBY, sts.UPDATEDBY
				FROM ADW_STS sts
				JOIN M_INOUT mi ON sts.M_INOUT_ID  = mi.M_INOUT_ID 
				LEFT JOIN ADW_TMS tms ON mi.ADW_TMS_ID  = tms.ADW_TMS_ID  
				WHERE 
					(tms.DRIVER = :1 OR sts.DRIVERBY = :2)
					AND mi.C_BPARTNER_ID = :3
					AND sts.STATUS = 'HO: DPK_TO_DRIVER'`

	err := r.db.SelectContext(ctx, &list, queryStr, driverID, driverID, customerID)
	if err != nil {
		return nil, fmt.Errorf("error database: %w", err)
	}
	return list, nil
}

func (r *oraRepo) GetNotificationDetails(ctx context.Context, ids []int64, driverID int64) ([]HandoverNotifyDTO, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = ":" + strconv.Itoa(i+1)
		args[i] = id
	}

	query := `
		SELECT 
			mi.DOCUMENTNO, 
			bp.VALUE AS CUSTOMER_NAME,
			NVL(au.NAME, 'N/A') AS DRIVER_NAME,
			TO_CHAR(SYSDATE, 'DD-MM-YYYY HH24:MI') AS TIME,
			TO_CHAR(mi.MovementDate, 'DD-MM-YYYY') AS MOVEMENTDATE,
			NVL(att.NAME, 'N/A') AS TNKB
		FROM M_INOUT mi
		LEFT JOIN C_BPARTNER bp ON mi.C_BPARTNER_ID = bp.C_BPARTNER_ID
		LEFT JOIN ADW_TMS t ON mi.ADW_TMS_ID = t.ADW_TMS_ID 
		LEFT JOIN ADW_STS sts ON mi.M_INOUT_ID = sts.M_INOUT_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON sts.TNKB_ID = att.ADW_TMS_TNKB_ID 
		WHERE mi.M_INOUT_ID IN (` + strings.Join(placeholders, ",") + `)`

	if driverID > 0 {
		lastIdx := len(ids) + 1
		query += ` AND sts.DRIVERBY = :` + strconv.Itoa(lastIdx)
		args = append(args, driverID)
	}

	var details []HandoverNotifyDTO
	err := r.db.SelectContext(ctx, &details, query, args...)
	return details, err
}

func (r *oraRepo) GetNotifLogActivityOnlyDetail(ctx context.Context, customerID, driverID int64) ([]HandoverNotifyDTO, error) {
	// Query untuk mengambil nama Driver dan nama Customer berdasarkan ID
	// SYSDATE digunakan untuk generate waktu saat ini langsung dari DB
	query := `
        SELECT 
            au.NAME AS DRIVER_NAME,
            bp.VALUE AS CUSTOMER_NAME,
            TO_CHAR(SYSDATE, 'DD-MM-YYYY HH24:MI') AS TIME
        FROM AD_USER au
        CROSS JOIN C_BPARTNER bp
        WHERE au.AD_USER_ID = :1
        AND bp.C_BPARTNER_ID = :2
    `

	var details []HandoverNotifyDTO
	err := r.db.SelectContext(ctx, &details, query, driverID, customerID)
	if err != nil {
		return nil, fmt.Errorf("error ambil detail log activity: %w", err)
	}

	return details, nil
}

func (r *oraRepo) CreateBatch(ctx context.Context, tx *sqlx.Tx, entities []TrackingSJ, eventType string, notes string) error {
	// 1. Logika Penanganan TX
	useExternalTx := tx != nil
	if !useExternalTx {
		var err error
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		// Hanya defer rollback jika kita yang membuat transaksi di sini
		defer tx.Rollback()
	}

	querySts := `
        INSERT INTO ADW_STS (
            ADW_STS_ID, AD_CLIENT_ID, AD_ORG_ID, M_INOUT_ID, 
            STATUS, ISACTIVE,
            CREATED, CREATEDBY, UPDATED, UPDATEDBY
        ) VALUES (:1, :2, :3, :4, :5, 'Y', SYSDATE, :6, SYSDATE, :7)`

	queryEvent := `
        INSERT INTO ADW_STS_EVENT (
            ADW_STS_EVENT_ID, ADW_STS_ID, AD_CLIENT_ID, AD_ORG_ID, 
            EVENTTYPE, CURRENTACTOR, ISACTIVE, NOTES,
            CREATED, CREATEDBY, UPDATED, UPDATEDBY
        ) VALUES (:1, :2, :3, :4, :5, :6, 'Y', :7, SYSDATE, :8, SYSDATE, :9)`

	queryUpdateInOut := `
        UPDATE M_INOUT 
        SET INSTS = 'Y', UPDATED = SYSDATE 
        WHERE M_INOUT_ID = :1`

	for _, e := range entities {
		// --- STEP A & B: Ambil ID Sequence ---
		var nextStsID, nextEventID int64

		if err := tx.GetContext(ctx, &nextStsID, "SELECT ADW_TRACKINGSJ_SQ.NEXTVAL FROM DUAL"); err != nil {
			return fmt.Errorf("gagal ambil sequence STS: %w", err)
		}
		if err := tx.GetContext(ctx, &nextEventID, "SELECT ADW_STS_EVENT_SQ.NEXTVAL FROM DUAL"); err != nil {
			return fmt.Errorf("gagal ambil sequence Event: %w", err)
		}

		// --- STEP C: Insert ADW_STS ---
		_, err := tx.ExecContext(ctx, querySts,
			nextStsID, e.ClientID, e.OrgID, e.MInOutID, e.Status, e.CreatedBy, e.CreatedBy)
		if err != nil {
			return fmt.Errorf("gagal insert STS (M_INOUT_ID %d): %w", e.MInOutID, err)
		}

		// --- STEP D: Insert ADW_STS_EVENT ---
		_, err = tx.ExecContext(ctx, queryEvent,
			nextEventID, nextStsID, e.ClientID, e.OrgID, eventType, e.CreatedBy, notes, e.CreatedBy, e.CreatedBy)
		if err != nil {
			return fmt.Errorf("gagal insert Event (STS_ID %d): %w", nextStsID, err)
		}

		// --- STEP E: Update flag di M_INOUT ---
		_, err = tx.ExecContext(ctx, queryUpdateInOut, e.MInOutID)
		if err != nil {
			return fmt.Errorf("gagal update M_INOUT flag (ID %d): %w", e.MInOutID, err)
		}
	}

	// 2. Commit hanya jika transaksi dibuat secara internal di fungsi ini
	if !useExternalTx {
		return tx.Commit()
	}

	return nil
}

func (r *oraRepo) UpdateBatch(ctx context.Context, tx *sqlx.Tx, entities []TrackingSJ, eventType string, notes string) error {
	useExternalTx := tx != nil
	if !useExternalTx {
		var err error
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	queryUpdate := `
		UPDATE ADW_STS 
			SET STATUS = :1, TNKB_ID = :2, DRIVERBY = :3, 
			UPDATED = SYSDATE, UPDATEDBY = :4, CURRENTCUSTOMER = :5 WHERE ADW_STS_ID = :6
		`
	queryEvent := `
		INSERT INTO ADW_STS_EVENT 
			(ADW_STS_EVENT_ID, ADW_STS_ID, AD_CLIENT_ID, AD_ORG_ID, 
			EVENTTYPE, PREVACTOR, ISACTIVE, CURRENTACTOR, 
			NOTES, CREATED, CREATEDBY, UPDATED, UPDATEDBY, DRIVERBY, TNKB_ID, CURRENTCUSTOMER, PREVCREATED) 
			VALUES 
			(:1, :2, :3, :4, :5, :6, 'Y', :7, :8, SYSDATE, :9, SYSDATE, :10, :11, :12, :13, :14)`

	for _, e := range entities {
		// 1. Update Tabel Utama
		if _, err := tx.ExecContext(ctx, queryUpdate, e.Status, e.TNKBID, e.DriverBy, e.UpdatedBy, e.CurrentCustomer, e.ID); err != nil {
			tx.Rollback()
			return err
		}

		// 2. Ambil Sequence Event secara manual (Cara yang Anda suka/berhasil)
		var nextEventID int64
		if err := tx.GetContext(ctx, &nextEventID, "SELECT ADW_STS_EVENT_SQ.NEXTVAL FROM DUAL"); err != nil {
			tx.Rollback()
			return err
		}

		// 3. Insert Log (Logic PrevActorID sudah dihitung Service)
		if _, err := tx.ExecContext(ctx, queryEvent, nextEventID, e.ID, e.ClientID, e.OrgID, eventType, e.PrevActorID, e.UpdatedBy, notes, e.UpdatedBy, e.UpdatedBy, e.DriverBy, e.TNKBID, e.CurrentCustomer, e.CreatedAt); err != nil {
			tx.Rollback()
			return err
		}
	}

	if !useExternalTx {
		return tx.Commit()
	}

	return nil
}

// Tambahkan di Interface
// LogActivityOnly(ctx context.Context, req HandoverRequest) error

func (r *oraRepo) LogActivityOnly(ctx context.Context, tx *sqlx.Tx, req HandoverRequest) error {
	useExternalTx := tx != nil
	if !useExternalTx {
		var err error
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	var nextEventID int64
	if err := tx.GetContext(ctx, &nextEventID, "SELECT ADW_STS_EVENT_SQ.NEXTVAL FROM DUAL"); err != nil {
		return fmt.Errorf("gagal ambil sequence: %w", err)
	}

	queryEvent := `
        INSERT INTO ADW_STS_EVENT 
            (ADW_STS_EVENT_ID, ADW_STS_ID, AD_CLIENT_ID, AD_ORG_ID, 
            EVENTTYPE, ISACTIVE, CURRENTACTOR, 
            NOTES, CREATED, CREATEDBY, UPDATED, UPDATEDBY, DRIVERBY, TNKB_ID, CURRENTCUSTOMER) 
        VALUES 
            (:1, NULL, :2, :3, :4, 'Y', :5, :6, SYSDATE, :7, SYSDATE, :8, :9, :10, :11)`

	_, err := tx.ExecContext(ctx, queryEvent,
		nextEventID,
		AdClientID, // Gunakan konstanta yang sama
		AdOrgID,    // Gunakan konstanta yang sama
		req.Status,
		req.UserID,
		req.Notes,
		req.UserID,
		req.UserID,
		req.DriverBy,
		req.TNKBID,
		req.CurrentCustomer,
	)

	if err != nil {
		return fmt.Errorf("gagal insert activity log: %w", err)
	}

	if !useExternalTx {
		return tx.Commit()
	}

	return nil
}

func (r *oraRepo) GetBundleActors(ctx context.Context, bundleDocNo string) (*BundleActorDTO, error) {
	query := `
        SELECT 
            NVL(prev_user.NAME, 'System') AS PREV_ACTOR_NAME,
            NVL(curr_user.NAME, 'System') AS CURRENT_ACTOR_NAME,
            evt.EVENTTYPE,
            TO_CHAR(evt.CREATED, 'DD-MM-YYYY HH24:MI') AS RECEIVETIME,
            TO_CHAR(evt.PREVCREATED, 'DD-MM-YYYY HH24:MI') AS HANDOVERTIME
        FROM plastik.ADW_STS_BUNDLE bnd
        JOIN ADW_STS_BUNDLE_LINE bln ON bnd.ADW_STS_BUNDLE_ID = bln.ADW_STS_BUNDLE_ID
        JOIN ADW_STS sts ON bln.ADW_STS_ID = sts.ADW_STS_ID
        JOIN ADW_STS_EVENT evt ON sts.ADW_STS_ID = evt.ADW_STS_ID AND evt.EVENTTYPE = bnd.BUNDLE_TYPE
        JOIN AD_USER prev_user ON evt.PREVACTOR = prev_user.AD_USER_ID
        JOIN AD_USER curr_user ON evt.CURRENTACTOR = curr_user.AD_USER_ID
        WHERE bnd.DOCUMENTNO = :1
        AND ROWNUM = 1
        ORDER BY evt.CREATED DESC
    `

	var actor BundleActorDTO
	err := r.db.GetContext(ctx, &actor, query, bundleDocNo)
	if err != nil {
		return nil, fmt.Errorf("gagal ambil actor info: %w", err)
	}

	return &actor, nil
}
