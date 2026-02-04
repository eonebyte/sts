package handover

import "time"

const (
	AdClientID = 1000000
	AdOrgID    = 1000000
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HandoverNotifyDTO struct {
	DocumentNo   string `db:"DOCUMENTNO"`
	CustomerName string `db:"CUSTOMER_NAME"`
	DriverName   string `db:"DRIVER_NAME"`
	Time         string `db:"TIME"`
	MovementDate string `db:"MOVEMENTDATE"`
	TNKB         string `db:"TNKB"`
}

type ADWBundle struct {
	ID           int64     `db:"ADW_STS_BUNDLE_ID"`
	ClientID     int64     `db:"AD_CLIENT_ID"`
	OrgID        int64     `db:"AD_ORG_ID"`
	DocumentNo   string    `db:"DOCUMENTNO"`
	BundleType   string    `db:"BUNDLE_TYPE"`
	MovementDate time.Time `db:"MOVEMENTDATE"`
	Description  string    `db:"DESCRIPTION"`
	CreatedBy    int64     `db:"CREATEDBY"`
	Attachment   string    `db:"ATTACHMENT"`
}

type TrackingSJ struct {
	ID              int64     `db:"ADW_STS_ID"`
	ClientID        int64     `db:"AD_CLIENT_ID"`
	OrgID           int64     `db:"AD_ORG_ID"`
	MInOutID        int64     `db:"M_INOUT_ID"`
	TNKBID          *int64    `db:"TNKB_ID"`
	Status          string    `db:"STATUS"`
	DriverBy        *int64    `db:"DRIVERBY"`
	CurrentCustomer *int64    `db:"CURRENTCUSTOMER"`
	IsActive        string    `db:"ISACTIVE"` // Y / N
	CreatedAt       time.Time `db:"CREATED"`
	CreatedBy       int64     `db:"CREATEDBY"`
	UpdatedAt       time.Time `db:"UPDATED"`
	UpdatedBy       int64     `db:"UPDATEDBY"`
	PrevActorID     int64     `db:"PREVACTOR"`
	CurrentActor    int64     `db:"CURRENTACTOR"`
}

// Request dari client
type HandoverRequest struct {
	MInOutIDs       []int64 `json:"m_inout_ids" binding:"required"`
	TNKBID          int64   `json:"tnkb_id" binding:"required"`
	Status          string  `json:"status" binding:"required"`
	DriverBy        int64   `json:"driver_by,omitempty"`
	CurrentCustomer int64   `json:"customer_id,omitempty"`
	UserID          int64   `json:"user_id"` // Untuk CreatedBy
	Notes           string  `json:"notes"`
}

type NotificationDetail struct {
	DocumentNo   string
	CustomerName string
	Time         string
	DriverName   string
	TNKB         string
}

type BundleActorDTO struct {
	PrevActorName    string `db:"PREV_ACTOR_NAME"`
	CurrentActorName string `db:"CURRENT_ACTOR_NAME"`
	EventType        string `db:"EVENTTYPE"`
	HandoverTime     string `db:"HANDOVERTIME"`
	ReceiveTime      string `db:"RECEIVETIME"`
}

// Response ke client
// type HandoverResponse struct {
// 	ID         int64  `json:"id"`
// 	MInOutID   int64  `json:"m_inout_id"`
// 	DocumentNo string `json:"document_no,omitempty"` // hasil JOIN
// 	TNKBID     int64  `json:"tnkb_id"`
// 	Status     string `json:"status"`
// 	IsActive   bool   `json:"is_active"`
// }
