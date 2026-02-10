package tms

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetDriverByName(ctx context.Context, searchKey string) ([]SearchDriver, error)
	ShipmentByDriver(ctx context.Context, driverID int64) ([]ShipmentByDriver, error)
	GetLogsByTMS(ctx context.Context, tmsID int64) ([]CustomerLog, error)
	UpdateEventLog(ctx context.Context, eventID int64, eventTime string, notes string) error
}

type oraRepo struct {
	db *sqlx.DB
}

func NewOraRepository(db *sqlx.DB) Repository {
	return &oraRepo{db: db}
}

func (r *oraRepo) ShipmentByDriver(ctx context.Context, driverID int64) ([]ShipmentByDriver, error) {
	var shipments []ShipmentByDriver

	query := `
        SELECT 
            sts.M_INOUT_ID, 
            tnkb.NAME AS TNKB_NO,
            (TO_CHAR(COALESCE(mi.MovementDateRev, mi.MovementDate), 'DD-MON-YY') || ' / ' || mi.DOCUMENTNO || ' / ' || cb.NAME) AS SURATJALAN 
        FROM ADW_STS sts
        JOIN M_Inout mi ON sts.M_Inout_ID = mi.M_Inout_ID
        JOIN C_BPartner cb ON mi.C_BPartner_ID = cb.C_BPartner_ID
        LEFT JOIN ADW_TMS_TNKB tnkb ON sts.TNKB_ID = tnkb.ADW_TMS_TNKB_ID
        WHERE sts.DRIVERBY = :1
            AND sts.STATUS IN ('HO: DPK_TO_DRIVER', 'HO: DRIVER_CHECKIN', 'HO: DRIVER_CHECKOUT')
			AND mi.ADW_TMS_ID IS NULL
			AND mi.MOVEMENTDATE >= (
				SELECT NVL(MAX(DATE_VALUE), TO_DATE('2026-02-01', 'YYYY-MM-DD')) 
				FROM ADW_STS_SETTING 
				WHERE SETTING_KEY = 'GLOBAL_CUTOFF_DATE'
			)
        ORDER BY sts.CREATED DESC
    `

	err := r.db.SelectContext(ctx, &shipments, query, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to search driver: %w", err)
	}

	return shipments, nil
}

func (r *oraRepo) GetDriverByName(ctx context.Context, searchKey string) ([]SearchDriver, error) {
	var sDriver []SearchDriver

	cleanSearch := strings.ToUpper(strings.ReplaceAll(searchKey, "%20", ""))
	query := `
		SELECT AD_USER_ID, NAME
        FROM AD_USER au
        WHERE au.TITLE = 'driver'
            AND UPPER(au.NAME) LIKE '%' || :1 || '%'
		ORDER BY au.NAME ASC
	`

	err := r.db.SelectContext(ctx, &sDriver, query, cleanSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to search driver: %w", err)
	}

	return sDriver, nil
}

func (r *oraRepo) GetLogsByTMS(ctx context.Context, tmsID int64) ([]CustomerLog, error) {
	list := []CustomerLog{}
	query := `
		SELECT 
			mi.C_BPARTNER_ID,
			cbp.VALUE AS CUSTOMER,
			MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKIN' THEN ase.ADW_STS_EVENT_ID END) AS CHECKIN_ID,
			MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKIN' THEN ase.CREATED END) AS CHECKIN,
			MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKIN' THEN ase.NOTES END) AS CHECKIN_NOTES,
			MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKOUT' THEN ase.ADW_STS_EVENT_ID END) AS CHECKOUT_ID,
			MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKOUT' THEN ase.CREATED END) AS CHECKOUT,
			MAX(CASE WHEN ase.EVENTTYPE = 'HO: DRIVER_CHECKOUT' THEN ase.NOTES END) AS CHECKOUT_NOTES
		FROM ADW_STS_EVENT ase
		JOIN ADW_STS t ON t.ADW_STS_ID = ase.ADW_STS_ID
		JOIN M_INOUT mi ON mi.M_INOUT_ID = t.M_INOUT_ID
		JOIN C_BPARTNER cbp ON mi.C_BPARTNER_ID = cbp.C_BPARTNER_ID
		WHERE mi.ADW_TMS_ID = :1
		GROUP BY mi.C_BPARTNER_ID, cbp.VALUE
		ORDER BY CHECKIN ASC
	`
	err := r.db.SelectContext(ctx, &list, query, tmsID)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *oraRepo) UpdateEventLog(ctx context.Context, eventID int64, eventTime string, notes string) error {
	query := `UPDATE ADW_STS_EVENT 
          SET CREATED = TO_DATE(:1, 'YYYY-MM-DD HH24:MI:SS'), 
              NOTES = :2 
          WHERE ADW_STS_EVENT_ID = :3`

	result, err := r.db.ExecContext(ctx, query, eventTime, notes, eventID)
	if err != nil {
		return err
	}

	// CEK: Apakah ada baris yang diupdate?
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("gagal update: data dengan ID %d tidak ditemukan", eventID)
	}

	fmt.Printf("Berhasil update %d baris\n", rows)
	return nil
}
