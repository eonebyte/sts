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
	GetDailyProgress(ctx context.Context) ([]ShipmentProgress, error)

	GetHistory(from, to time.Time) ([]ShipmentHistory, error)

	GetOutstandingDPK(from, to time.Time) ([]Shipment, error)
	GetOutstandingDelivery(from, to time.Time) ([]Shipment, error)

	UpdateDriverTnkb(ctx context.Context, inoutID int64, driverID int64, tnkbID int64) error
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

func (r *oraRepo) GetDailyProgress(ctx context.Context) ([]ShipmentProgress, error) {
	query := `
    SELECT 
        io.DOCUMENTNO, cb.VALUE CUSTOMER, io.MOVEMENTDATE, 
        NVL(au.NAME, '-') AS DRIVER, NVL(att.NAME, '-') AS TNKB,
        MAX(CASE WHEN io.INSTS = 'Y' THEN 1 ELSE 0 END) AS DELIVERY,
        MAX(CASE WHEN ase.EVENTTYPE = 'RE: DPK_FROM_DEL' THEN 1 ELSE 0 END) AS ONDPK,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DPK_TO_DRIVER' THEN 1 ELSE 0 END) AS ONDRIVER,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKIN' THEN 1 ELSE 0 END) AS ONCUSTOMER,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKOUT' THEN 1 ELSE 0 END) AS OUTCUSTOMER,
        MAX(CASE WHEN ase.EVENTTYPE = 'RE: DPK_FROM_DRIVER' THEN 1 ELSE 0 END) AS COMEBACKDPK,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DPK_TO_DEL' THEN 1 ELSE 0 END) AS COMEBACKDEL,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: DEL_TO_MKT' THEN 1 ELSE 0 END) AS COMEBACKMKT,
        MAX(CASE WHEN ase.EVENTTYPE = 'HO: MKT_TO_FAT' THEN 1 ELSE 0 END) AS COMEBACKFAT,
		MAX(CASE WHEN ase.EVENTTYPE = 'RE: FAT_FROM_MKT' THEN 1 ELSE 0 END) AS FINISHFAT
    FROM M_INOUT io 
    LEFT JOIN ADW_STS sts ON io.M_INOUT_ID = sts.M_INOUT_ID 
    JOIN C_BPARTNER cb ON io.C_BPARTNER_ID = cb.C_BPARTNER_ID
    LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
    LEFT JOIN ADW_TMS_TNKB att ON sts.TNKB_ID = att.ADW_TMS_TNKB_ID 
    LEFT JOIN ADW_STS_EVENT ase ON sts.ADW_STS_ID = ase.ADW_STS_ID
    WHERE io.MOVEMENTDATE >= TRUNC(SYSDATE) AND io.MOVEMENTDATE < TRUNC(SYSDATE) + 1
    GROUP BY io.DOCUMENTNO, cb.VALUE, io.MOVEMENTDATE, au.NAME, att.NAME 
    ORDER BY (DELIVERY + ONDPK + ONDRIVER + ONCUSTOMER + OUTCUSTOMER + 
              COMEBACKDPK + COMEBACKDEL + COMEBACKMKT + COMEBACKFAT + FINISHFAT) DESC, DOCUMENTNO ASC`

	var results []ShipmentProgress
	err := r.db.SelectContext(ctx, &results, query)
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
		WHERE 
			co.DOCSTATUS = 'CO'
			AND co.ISSOTRX = 'Y'
			AND co.C_DOCTYPETARGET_ID IN (1000054, 1000053) --Delivery Order, Schedule Order
			AND co.DATEORDERED >= ADD_MONTHS(TRUNC(SYSDATE), -12)
			AND co.DATEORDERED <  TRUNC(SYSDATE) + 1
			AND cb.ISACTIVE = 'Y' 
			AND cb.ISCUSTOMER = 'Y'
			AND cb.ISSUBCONTRACT = 'N'
			AND cb.AD_CLIENT_ID = 1000000
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
			att.NAME TNKBNO
		FROM M_InOut mi
		JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID 
		LEFT JOIN ADW_STS sts ON mi.M_INOUT_ID = sts.M_INOUT_ID 
		LEFT JOIN AD_USER au ON sts.DRIVERBY = au.AD_USER_ID 
		LEFT JOIN ADW_TMS_TNKB att ON att.ADW_TMS_TNKB_ID = sts.TNKB_ID 
		WHERE mi.movementdate >= :1
		  AND mi.movementdate < :2
		  AND mi.ADW_TMS_ID IS NULL
  		  AND mi.C_INVOICE_ID IS NULL
		  AND mi.IsSoTrx = 'Y'
		  AND mi.INSTS = 'N'
		ORDER BY
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
		  AND sts.STATUS = 'HO: DEL_TO_DPK'
		ORDER BY
			movementdate ASC
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
		ORDER BY
			movementdate ASC
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
		  AND sts.STATUS = 'HO: DEL_TO_MKT'
		ORDER BY
			movementdate ASC
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
		  AND sts.STATUS = 'RE: MKT_FROM_DEL'
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
		  AND sts.STATUS <> 'RE: DEL_FROM_DPK'
		ORDER BY
			movementdate ASC
	`

	err := r.db.Select(&list, query, from, to)
	if err != nil {
		return nil, err
	}

	return list, nil
}
