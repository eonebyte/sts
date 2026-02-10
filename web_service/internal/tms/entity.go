package tms

import "time"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Count   int         `json:"count"`
	Data    interface{} `json:"data,omitempty"`
}

type SearchDriver struct {
	AD_USER_ID int64  `db:"AD_USER_ID", json:"AD_USER_ID"`
	Name       string `db:"NAME" json:"NAME"`
}

type ShipmentByDriver struct {
	MInoutID   int64  `db:"M_INOUT_ID" json:"M_INOUT_ID"`
	TNKBNo     string `db:"TNKB_NO" json:"-"`
	SuratJalan string `db:"SURATJALAN" json:"SURATJALAN"`
}

type CustomerLog struct {
	BPartnerID    int64      `db:"C_BPARTNER_ID" json:"c_bpartner_id"`
	CustomerName  string     `db:"CUSTOMER" json:"customer_name"` // Pastikan ini CUSTOMER (huruf besar)
	CheckInID     *int64     `db:"CHECKIN_ID" json:"checkin_id"`
	CheckIn       *time.Time `db:"CHECKIN" json:"checkin"`
	CheckInNotes  string     `db:"CHECKIN_NOTES" json:"checkin_notes"`
	CheckOutID    *int64     `db:"CHECKOUT_ID" json:"checkout_id"`
	CheckOut      *time.Time `db:"CHECKOUT" json:"checkout"`
	CheckOutNotes string     `db:"CHECKOUT_NOTES" json:"checkout_notes"`
}

type UpdateLogRequest struct {
	EventID   int64  `json:"event_id"`
	EventTime string `json:"event_time"` // Format: 2026-02-09T14:30
	Notes     string `json:"notes"`
}
