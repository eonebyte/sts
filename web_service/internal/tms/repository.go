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
            AND sts.STATUS = 'HO: DPK_TO_DRIVER'
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
