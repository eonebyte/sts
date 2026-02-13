package shipment

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetAllCustomers() ([]Customer, error)
	GetDriver() ([]Driver, error)
	GetTnkb() ([]TNKB, error)
	GetPending(from, to time.Time) ([]Shipment, error)
	GetPrepare(from, to time.Time) ([]Shipment, error)
	GetPrepareToLeave(from, to time.Time) ([]Shipment, error)
	GetInTransitCustomer(from, to time.Time, driverID int64) ([]Shipment, error)
	GetOnCustomer(from, to time.Time, customerID, driverID int64) ([]Shipment, error)

	GetComeback(from, to time.Time) ([]Shipment, error)
	GetComebackToDelivery(from, to time.Time) ([]Shipment, error)
	GetReceiptComebackToDelivery(from, to time.Time) ([]Shipment, error)

	GetComebackToMarketing(from, to time.Time) ([]Shipment, error)
	GetReceiptComebackToMarketing(from, to time.Time) ([]Shipment, error)

	GetComebackToFat(from, to time.Time) ([]Shipment, error)
	GetReceiptComebackToFat(from, to time.Time) ([]Shipment, error)

	//Progress Shipment
	GetDailyProgress(ctx context.Context, from, to time.Time) ([]ShipmentProgress, error)

	GetHistory(from, to time.Time) ([]ShipmentHistory, error)

	GetOutstandingDPK(from, to time.Time) ([]Shipment, error)
	GetOutstandingDelivery(from, to time.Time) ([]Shipment, error)

	UpdateDriverTnkb(ctx context.Context, inoutID int64, driverID int64, tnkbID int64) error

	ExecuteOutstandingCancel(ctx context.Context, inoutID int64, nextStatus string, isHardDelete bool) error
}

type oraRepo struct {
	db *sqlx.DB
}

func NewOraRepository(db *sqlx.DB) Repository {
	return &oraRepo{db: db}
}

func (r *oraRepo) UpdateDriverTnkb(ctx context.Context, inoutID int64, driverID int64, tnkbID int64) error {
	// Mulai transaksi
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	// 1. Update tabel utama ADW_STS
	querySts := `
        UPDATE ADW_STS 
        SET DRIVERBY = :1, 
            TNKB_ID = :2, 
            UPDATED = SYSDATE 
        WHERE M_INOUT_ID = :3`

	_, err = tx.ExecContext(ctx, querySts, driverID, tnkbID, inoutID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update ADW_STS: %w", err)
	}

	// 2. Update tabel ADW_STS_EVENT
	// Filter berdasarkan ADW_STS_ID (lewat join ke M_INOUT_ID)
	// dan EVENTTYPE 'HO: DPK_TO_DRIVER'
	queryEvent := `
        UPDATE ADW_STS_EVENT ase
        SET ase.DRIVERBY = :1,
            ase.TNKB_ID = :2,
            ase.UPDATED = SYSDATE,
            ase.NOTES = ase.NOTES || 'HO:DPK_TO_DRIVER (Edited)'
        WHERE ase.EVENTTYPE = 'HO: DPK_TO_DRIVER'
        AND ase.ADW_STS_ID = (
            SELECT ADW_STS_ID FROM ADW_STS WHERE M_INOUT_ID = :3
        )`

	_, err = tx.ExecContext(ctx, queryEvent, driverID, tnkbID, inoutID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update ADW_STS_EVENT: %w", err)
	}

	// Selesaikan transaksi
	return tx.Commit()
}

func (r *oraRepo) ExecuteOutstandingCancel(ctx context.Context, inoutID int64, nextStatus string, isHardDelete bool) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if isHardDelete {
		// --- KONDISI HARD DELETE (HO: DEL_TO_DPK) ---
		queryDelEvent := `DELETE FROM ADW_STS_EVENT WHERE ADW_STS_ID IN (SELECT ADW_STS_ID FROM ADW_STS WHERE M_InOut_ID = :1)`
		if _, err := tx.ExecContext(ctx, queryDelEvent, inoutID); err != nil {
			return fmt.Errorf("failed hard delete event: %w", err)
		}

		queryDelSts := `DELETE FROM ADW_STS WHERE M_InOut_ID = :1`
		if _, err := tx.ExecContext(ctx, queryDelSts, inoutID); err != nil {
			return fmt.Errorf("failed hard delete status: %w", err)
		}

		queryUpdateInSts := `
			UPDATE M_InOut 
			SET INSTS = :1
			WHERE M_InOut_ID = :2`

		if _, err := tx.ExecContext(ctx, queryUpdateInSts, "N", inoutID); err != nil {
			return fmt.Errorf("failed to update m_inout insts: %w", err)
		}
	} else {
		// --- KONDISI REVERSE STATUS ---

		// 1. Ambil data STS ID dan Status saat ini
		var currentStatus string
		var stsID int64
		// Gunakan QueryRowContext untuk mengambil data tunggal
		err := tx.QueryRowContext(ctx, `SELECT ADW_STS_ID, STATUS FROM ADW_STS WHERE M_InOut_ID = :1`, inoutID).Scan(&stsID, &currentStatus)
		if err != nil {
			return fmt.Errorf("status record not found: %w", err)
		}

		// 2. HAPUS event yang status sts-nya sama dengan event type
		queryDelEvent := `DELETE FROM ADW_STS_EVENT WHERE ADW_STS_ID = :1 AND EVENTTYPE = :2`
		if _, err := tx.ExecContext(ctx, queryDelEvent, stsID, currentStatus); err != nil {
			return fmt.Errorf("failed to delete current event: %w", err)
		}

		// 3. AMBIL CURRENTACTOR dari event terakhir yang tersisa
		var prevActorID int64
		var prevTimeUpdated time.Time

		queryGetLastActor := `
			SELECT CREATED, CURRENTACTOR FROM (
				SELECT CREATED, CURRENTACTOR FROM ADW_STS_EVENT 
				WHERE ADW_STS_ID = :1 
				ORDER BY ADW_STS_EVENT_ID DESC
			) WHERE ROWNUM = 1`

		// Perbaikan: Gunakan QueryRowContext + Scan karena mengambil > 1 kolom
		err = tx.QueryRowContext(ctx, queryGetLastActor, stsID).Scan(&prevTimeUpdated, &prevActorID)
		if err != nil {
			return fmt.Errorf("failed to find previous actor from remaining events: %w", err)
		}

		// 4. UPDATE ADW_STS: balikkan status dan isi UPDATEDBY dengan aktor sebelumnya
		queryUpdateSts := `
			UPDATE ADW_STS 
			SET STATUS = :1, 
			    UPDATEDBY = :2,
			    UPDATED = :3 
			WHERE ADW_STS_ID = :4`

		if _, err := tx.ExecContext(ctx, queryUpdateSts, nextStatus, prevActorID, prevTimeUpdated, stsID); err != nil {
			return fmt.Errorf("failed to restore status and updatedby: %w", err)
		}
	}

	return tx.Commit()
}

func (r *oraRepo) GetDailyProgress(ctx context.Context, from, to time.Time) ([]ShipmentProgress, error) {
	query := `
    SELECT 
        io.DOCUMENTNO, 
		CASE
			WHEN io.ADW_TMS_ID IS NOT NULL THEN 'Y'
			ELSE 'N'
		END MATCHTMS,
		cb.VALUE CUSTOMER, 
		io.MOVEMENTDATE, 
        NVL(NVL(au.NAME, au2.NAME), '-') AS DRIVER, 
		NVL(NVL(att.NAME, t.TNKB), '-') AS TNKB,
        MAX(CASE WHEN io.INSTS = 'Y' OR io.ADW_TMS_ID IS NOT NULL AND io.SPPNO IS NOT NULL THEN 1 ELSE 0 END) AS DELIVERY,
	    MAX(CASE WHEN ase.EVENTTYPE = 'HO: DPK_TO_DRIVER' OR io.ADW_TMS_ID IS NOT NULL AND io.SPPNO IS NOT NULL THEN 1 ELSE 0 END) AS ONDPK,
	    MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKIN' OR io.ADW_TMS_ID IS NOT NULL AND io.SPPNO IS NOT NULL THEN 1 ELSE 0 END) AS ONDRIVER,
	    MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKOUT' OR io.ADW_TMS_ID IS NOT NULL AND io.SPPNO IS NOT NULL THEN 1 ELSE 0 END) AS ONCUSTOMER,
        MAX(CASE WHEN ase.EVENTTYPE = 'RE: DPK_FROM_DRIVER' THEN 1 ELSE 0 END) AS OUTCUSTOMER,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DPK_TO_DEL' THEN 1 ELSE 0 END) AS COMEBACKDPK,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DEL_TO_MKT' THEN 1 ELSE 0 END) AS COMEBACKDEL,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: MKT_TO_FAT' THEN 1 ELSE 0 END) AS COMEBACKMKT,
        MAX(CASE WHEN ase.EVENTTYPE = 'RE: FAT_FROM_MKT' THEN 1 ELSE 0 END) AS COMEBACKFAT
    FROM M_INOUT io 
    LEFT JOIN ADW_STS sts ON io.M_INOUT_ID = sts.M_INOUT_ID 
	LEFT JOIN ADW_TMS t ON io.ADW_TMS_ID = t.ADW_TMS_ID
    JOIN C_BPARTNER cb ON io.C_BPARTNER_ID = cb.C_BPARTNER_ID
    LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
	lEFT JOIN AD_USER au2 ON t.DRIVER = au2.AD_USER_ID 
    LEFT JOIN ADW_TMS_TNKB att ON sts.TNKB_ID = att.ADW_TMS_TNKB_ID 
    LEFT JOIN ADW_STS_EVENT ase ON sts.ADW_STS_ID = ase.ADW_STS_ID
    WHERE  io.movementdate >= :1
        AND io.movementdate < :2
		AND io.AD_Client_ID = 1000000
		AND io.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		AND io.ISSOTRX = 'Y'
    GROUP BY io.DOCUMENTNO, io.ADW_TMS_ID, cb.VALUE, io.MOVEMENTDATE, au.NAME, au2.NAME, att.NAME, t.TNKB 
    ORDER BY (DELIVERY + ONDPK + ONDRIVER + ONCUSTOMER + OUTCUSTOMER + 
              COMEBACKDPK + COMEBACKDEL + COMEBACKMKT + COMEBACKFAT) DESC, DOCUMENTNO ASC`

	var results []ShipmentProgress
	err := r.db.SelectContext(ctx, &results, query, from, to)
	return results, err
}

func (r *oraRepo) GetHistory(from, to time.Time) ([]ShipmentHistory, error) {
	var list []ShipmentHistory

	// Query melakukan join ke bundle_line dan bundle untuk mendapatkan PDF
	query := `
        SELECT 
            mi.M_InOut_ID, 
            mi.DocumentNo, 
            mi.MovementDate,
            cb.Value Customer,
            NVL(au.NAME, '-') Driver,
            sts.STATUS,
            asb.DOCUMENTNO as BundleNo,
            asb.ATTACHMENT
        FROM ADW_STS sts
        JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
        JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
        LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
        LEFT JOIN ADW_STS_BUNDLE_LINE asbl ON sts.ADW_STS_ID = asbl.ADW_STS_ID
        LEFT JOIN ADW_STS_BUNDLE asb ON asbl.ADW_STS_BUNDLE_ID = asb.ADW_STS_BUNDLE_ID
        WHERE mi.movementdate >= :1
          AND mi.movementdate < :2
          AND mi.IsSoTrx = 'Y'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
        ORDER BY asb.CREATED DESC, mi.movementdate ASC
    `

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetAllCustomers() ([]Customer, error) {

	queryStr := `
		SELECT DISTINCT
			cb.C_BPARTNER_ID CUSTOMERID, cb.VALUE CUSTOMERNAME
		FROM 
			C_ORDER co 
		JOIN C_BPARTNER cb ON co.C_BPARTNER_ID = cb.C_BPARTNER_ID
		WHERE cb.ISACTIVE = 'Y' 
            AND cb.ISCUSTOMER = 'Y'
            AND cb.ISSUBCONTRACT = 'N'
            AND cb.AD_CLIENT_ID = 1000000
            AND EXISTS (
                SELECT 1 
                FROM M_INOUT mi
                WHERE mi.C_BPARTNER_ID = cb.C_BPARTNER_ID
                  AND mi.ISSOTRX = 'Y'
                  AND mi.MOVEMENTDATE >= (
                      SELECT DATE_VALUE 
                      FROM ADW_STS_SETTING 
                      WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
                  )
            )
	`

	var list []Customer

	err := r.db.Select(&list, queryStr)
	if err != nil {
		return nil, fmt.Errorf("error database: %w", err)
	}

	return list, nil
}

func (r *oraRepo) GetDriver() ([]Driver, error) {
	var list []Driver

	query := `
		SELECT au.AD_USER_ID, au.NAME FROM AD_USER au WHERE au.TITLE = 'driver' AND au.ISACTIVE = 'Y'
	`

	err := r.db.Select(&list, query)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetTnkb() ([]TNKB, error) {
	var list []TNKB

	query := `
		SELECT DISTINCT att.ADW_TMS_TNKB_ID, att.NAME FROM ADW_TMS_TNKB att WHERE att.ISACTIVE = 'Y'
	`

	err := r.db.Select(&list, query)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// For Delivery
func (r *oraRepo) GetPending(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO,
			mi.ADW_TMS_ID
		FROM M_InOut mi
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN ADW_STS sts ON mi.M_INOUT_ID = sts.M_INOUT_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
		  -- AND mi.ADW_TMS_ID IS NULL
  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.AD_Client_ID = 1000000
		  AND mi.IsSoTrx = 'Y'
		  -- AND mi.INSTS = 'N'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			mi.ADW_TMS_ID ASC NULLS FIRST,
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetPrepare(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	// 	query := `
	// 		SELECT
	// 			mi.M_InOut_ID,
	// 			mi.DocumentNo,
	// 			mi.MovementDate,
	// 			cb.Value Customer,
	// 			au.NAME Driver,
	// 			att.NAME TNKBNO
	// 		FROM ADW_STS sts
	// 		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID
	// 		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID
	// 		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID
	// 		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID
	// 		WHERE mi.movementdate >= :1
	// 		  AND mi.movementdate < :2
	// --		  AND mi.ADW_TMS_ID IS NULL
	// --  		  AND mi.C_INVOICE_ID IS NULL
	// 		  AND mi.IsSoTrx = 'Y'
	// 		  AND mi.INSTS = 'Y'
	// 		  AND sts.STATUS = 'HO: DEL_TO_DPK'
	// 		  AND mi.MOVEMENTDATE >= (
	// 				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD'))
	// 				FROM ADW_STS_SETTING
	// 				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
	// 			)
	// 		ORDER BY
	// 			movementdate ASC
	// 	`

	query := `
	SELECT 
    M_InOut_ID, 
    DocumentNo, 
    MovementDate, 
    Customer, 
    Driver, 
    TNKBNO
FROM (
    SELECT 
        mi.M_InOut_ID, 
        mi.DocumentNo, 
        mi.MovementDate,
        cb.Value AS Customer,
        au.NAME AS Driver,
        att.NAME AS TNKBNO,
        -- Memberi nomor baris per DocumentNo, diurutkan dari ID status terbaru
        ROW_NUMBER() OVER (PARTITION BY mi.M_InOut_ID ORDER BY sts.ADW_STS_ID DESC) as rn
    FROM ADW_STS sts
    JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
    JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
    LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
    LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
    WHERE mi.movementdate >= :1
	  AND mi.movementdate < :2
      AND mi.IsSoTrx = 'Y'
      AND mi.INSTS = 'Y'
      AND sts.STATUS = 'HO: DEL_TO_DPK'
      AND mi.MOVEMENTDATE >= (
            SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
            FROM ADW_STS_SETTING 
            WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
        )
) 
WHERE rn = 1 -- Hanya ambil baris pertama (terbaru) untuk setiap M_InOut_ID
ORDER BY MovementDate ASC
`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetPrepareToLeave(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'RE: DPK_FROM_DEL'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetInTransitCustomer(from, to time.Time, driverID int64) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			cb.C_BPartner_ID CustomerID,
			au.NAME Driver,
			att.ADW_TMS_TNKB_ID TNKBID,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'HO: DPK_TO_DRIVER'
		  AND sts.DRIVERBY = :3
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to, driverID)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetOnCustomer(from, to time.Time, customerID, driverID int64) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			sts.Status,
			cb.C_BPartner_ID CustomerID,
			au.NAME Driver,
			att.ADW_TMS_TNKB_ID TNKBID,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'HO: DRIVER_CHECKIN'
		  AND sts.CURRENTCUSTOMER = :3 OR sts.DRIVERBY = :4
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			-- mi.DocumentNo DESC
			mi.movementdate ASC
	`

	err := r.db.Select(&list, query, from, to, customerID, driverID)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetComeback(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS IN ('HO: DPK_TO_DRIVER', 'HO: DRIVER_CHECKIN', 'HO: DRIVER_CHECKOUT')
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetComebackToDelivery(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'RE: DPK_FROM_DRIVER'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetReceiptComebackToDelivery(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'HO: DPK_TO_DEL'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetComebackToMarketing(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'RE: DEL_FROM_DPK'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetReceiptComebackToMarketing(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			mi.SPPNO,
			sts.Status
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'HO: DEL_TO_MKT'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY 
			-- 1. Prioritaskan yang SPPNO tidak null dan tidak kosong
			CASE 
				WHEN mi.SPPNO IS NOT NULL AND mi.SPPNO <> ' ' THEN 0 
				ELSE 1 
			END ASC,
			-- 2. Baru kemudian urutkan berdasarkan tanggal
			mi.MovementDate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetComebackToFat(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			mi.SPPNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'RE: MKT_FROM_DEL'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		  AND mi.SPPNO IS NOT NULL
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetReceiptComebackToFat(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'HO: MKT_TO_FAT'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetOutstandingDPK(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO,
			sts.Status
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS = 'HO: DPK_TO_DRIVER'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) GetOutstandingDelivery(from, to time.Time) ([]Shipment, error) {
	var list []Shipment

	query := `
		SELECT 
			mi.M_InOut_ID, 
			mi.DocumentNo, 
			mi.MovementDate,
			cb.Value Customer,
			au.NAME Driver,
			att.NAME TNKBNO,
			sts.Status
		FROM ADW_STS sts
		JOIN M_InOut mi ON sts.M_INOUT_ID = mi.M_INOUT_ID  
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
--		  AND mi.ADW_TMS_ID IS NULL
--  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'Y'
		  AND sts.STATUS <> 'RE: DEL_FROM_DPK'
		  AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}
