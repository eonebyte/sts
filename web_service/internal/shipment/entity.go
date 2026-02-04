package shipment

import "time"

type Shipment struct {
	MInOutID     int64     `db:"M_INOUT_ID" json:"m_inout_id"`
	DocumentNo   string    `db:"DOCUMENTNO" json:"document_no"`
	MovementDate time.Time `db:"MOVEMENTDATE" json:"movement_date"`
	Customer     string    `db:"CUSTOMER" json:"customer_name"`
	Status       *string   `db:"STATUS" json:"status"`
	Driver       *string   `db:"DRIVER" json:"driver_name"`
	CustomerID   *string   `db:"CUSTOMERID" json:"customer_id"`
	TNKBID       *string   `db:"TNKBID" json:"tnkb_id"`
	TNKBNo       *string   `db:"TNKBNO" json:"tnkb_no"`
}

type ShipmentProgress struct {
	DocumentNo   string    `db:"DOCUMENTNO" json:"documentno"`
	Customer     string    `db:"CUSTOMER" json:"customer"`
	MovementDate time.Time `db:"MOVEMENTDATE" json:"movementdate"`
	Driver       string    `db:"DRIVER" json:"driver"`
	TNKB         string    `db:"TNKB" json:"tnkb"`
	Delivery     int       `db:"DELIVERY" json:"delivery"`
	OnDPK        int       `db:"ONDPK" json:"ondpk"`
	OnDriver     int       `db:"ONDRIVER" json:"ondriver"`
	OnCustomer   int       `db:"ONCUSTOMER" json:"oncustomer"`
	OutCustomer  int       `db:"OUTCUSTOMER" json:"outcustomer"`
	ComebackDPK  int       `db:"COMEBACKDPK" json:"comebackdpk"`
	ComebackDel  int       `db:"COMEBACKDEL" json:"comebackdel"`
	ComebackMkt  int       `db:"COMEBACKMKT" json:"comebackmkt"`
	ComebackFat  int       `db:"COMEBACKFAT" json:"comebackfat"`
	FinishFat    int       `db:"FINISHFAT" json:"finishfat"`
}

type ShipmentHistory struct {
	M_InOut_ID     int64     `db:"M_INOUT_ID" json:"m_inout_id"`
	DocumentNo     string    `db:"DOCUMENTNO" json:"document_no"`
	MovementDate   time.Time `db:"MOVEMENTDATE" json:"movement_date"`
	Customer       string    `db:"CUSTOMER" json:"customer_name"`
	Driver         string    `db:"DRIVER" json:"driver_name"`
	Status         string    `db:"STATUS" json:"status"`
	BundleNo       *string   `db:"BUNDLENO" json:"bundle_no"` // Gunakan pointer untuk null safety
	AttachmentPath *string   `db:"ATTACHMENT" json:"attachment_path"`
}

type Customer struct {
	CustomerID   int64  `db:"CUSTOMERID" json:"customer_id"`
	CustomerName string `db:"CUSTOMERNAME" json:"customer_name"`
}

type Driver struct {
	ID   int64  `db:"AD_USER_ID" json:"driver_by"`
	Name string `db:"NAME", json:"driver_name"`
}

type TNKB struct {
	ID   int64  `db:"ADW_TMS_TNKB_ID" json:"tnkb_id"`
	Name string `db:"NAME", json:"tnkb_no"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Count   int         `json:"count"`
	Data    interface{} `json:"data,omitempty"`
}

type UpdateShipmentRequest struct {
	MInOutID int64 `json:"m_inout_id"`
	DriverBy int64 `json:"driver_by"`
	TnkbID   int64 `json:"tnkb_id"`
	// UserID   int64 `json:"user_id"`
}
