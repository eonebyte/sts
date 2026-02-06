package tms

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
