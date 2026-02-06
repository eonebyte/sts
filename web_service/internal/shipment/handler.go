package shipment

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sts/web_service/internal/shared"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{service: s}
}

func (h *handler) RegisterProtectedRoutes(r chi.Router) {
	r.Route("/shipments", func(r chi.Router) {
		r.Get("/customers", h.GetCustomers)
		r.Get("/drivers", h.GetDrivers)
		r.Get("/tnkbs", h.GetTnkbs)
		r.Get("/pending", h.GetPendingShipments)
		r.Get("/prepare", h.GetPrepareShipments)
		r.Get("/preparetoleave", h.GetPrepareToLeaveShipments)
		r.Get("/in-transit", h.GetInTransitShipments)
		r.Get("/on-customer", h.GetOnCustomerShipments)
		r.Get("/comeback", h.GetComebackShipments)
		r.Get("/comebacktodelivery", h.GetComebackToDeliveryShipments)
		r.Get("/receiptcomebacktodelivery", h.GetReceiptComebackToDeliveryShipments)

		r.Get("/comebacktomarketing", h.GetComebackToMarketingShipments)
		r.Get("/receiptcomebacktomarketing", h.GetReceiptComebackToMarketingShipments)

		r.Get("/comebacktofat", h.GetComebackToFatShipments)
		r.Get("/receiptcomebacktofat", h.GetReceiptComebackToFatShipments)

		r.Get("/history", h.GetHistoryShipments)
		r.Get("/progress", h.GetShipmentProgress)

		r.Get("/outstanding/dpk", h.GetOutstandingDPKShipments)
		r.Get("/outstanding/delivery", h.GetOutstandingDeliveryShipments)
		r.Post("/edit/drivertnkb", h.HandleEditShipment)
		r.Post("/outstanding/cancel", h.CancelOutstanding)
	})
}

func (h *handler) GetHistoryShipments(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil parameter dari query URL
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	// 2. Panggil service
	list, err := h.service.GetHistory(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed to get shipment history",
		})
		return
	}

	// 3. Handle data kosong agar return array [] bukan null
	if list == nil {
		list = []ShipmentHistory{}
	}

	countData := len(list)

	// 4. Kirim response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetShipmentProgress(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.FetchProgress(r.Context())
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get progress shipment",
		})
		return
	}

	if data == nil {
		data = []ShipmentProgress{}
	}

	countData := len(data)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    data,
	})
}

func (h *handler) GetDrivers(w http.ResponseWriter, r *http.Request) {

	list, err := h.service.GetDriver()
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get drivers",
		})

		return
	}

	if list == nil {
		list = []Driver{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetCustomers(w http.ResponseWriter, r *http.Request) {

	list, err := h.service.GetAllCustomers()
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get tnkbs",
		})

		return
	}

	if list == nil {
		list = []Customer{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetTnkbs(w http.ResponseWriter, r *http.Request) {

	list, err := h.service.GetTnkb()
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get tnkbs",
		})

		return
	}

	if list == nil {
		list = []TNKB{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetPendingShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetPending(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetPrepareShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetPrepare(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetPrepareToLeaveShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetPrepareToLeave(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetInTransitShipments(w http.ResponseWriter, r *http.Request) {
	driverIDStr := r.URL.Query().Get("driverId")
	driverID, _ := strconv.Atoi(driverIDStr)

	list, err := h.service.GetInTransit(int64(driverID))
	if err != nil {
		log.Printf(
			"[SERVICESS] path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{Success: false, Message: err.Error()})
		return
	}

	render.JSON(w, r, APIResponse{
		Success: true,
		Data:    list,
	})
}

func (h *handler) GetOnCustomerShipments(w http.ResponseWriter, r *http.Request) {
	customerIDStr := r.URL.Query().Get("customerId")
	driverIDStr := r.URL.Query().Get("driverId")
	customerID, _ := strconv.Atoi(customerIDStr)
	driverID, _ := strconv.Atoi(driverIDStr)

	list, err := h.service.GetOnCustomer(int64(customerID), int64(driverID))
	if err != nil {
		log.Printf(
			"[SERVICESS] path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{Success: false, Message: err.Error()})
		return
	}

	render.JSON(w, r, APIResponse{
		Success: true,
		Data:    list,
	})
}

func (h *handler) GetComebackShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetComeback(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetComebackToDeliveryShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetComebackToDelivery(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetReceiptComebackToDeliveryShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetReceiptComebackToDelivery(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetComebackToMarketingShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetComebackToMarketing(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetReceiptComebackToMarketingShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetReceiptComebackToMarketing(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetComebackToFatShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetComebackToFat(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetReceiptComebackToFatShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetReceiptComebackToFat(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) HandleEditShipment(w http.ResponseWriter, r *http.Request) {
	var req UpdateShipmentRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Invalid JSON payload",
		})
		return
	}

	// Validasi input minimal
	if req.MInOutID == 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "m_inout_id is required",
		})
		return
	}

	// Pastikan urutan parameter sesuai: ctx, inoutID, driverID, tnkbID, userID
	err := h.service.UpdateDriverTnkb(r.Context(), req.MInOutID, req.DriverBy, req.TnkbID)

	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed to update driver and TNKB: " + err.Error(),
		})
		return
	}

	// Response Sukses
	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "Shipment and Event Log updated successfully",
	})
}

func (h *handler) GetOutstandingDPKShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetOutstandingDPK(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) GetOutstandingDeliveryShipments(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("dateFrom")
	to := r.URL.Query().Get("dateTo")

	list, err := h.service.GetOutstandingDelivery(from, to)
	if err != nil {
		log.Printf(
			"[SERVICE]: path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Failed get shipments",
		})

		return
	}

	if list == nil {
		list = []Shipment{}
	}

	countData := len(list)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Count:   countData,
		Data:    list,
	})
}

func (h *handler) CancelOutstanding(w http.ResponseWriter, r *http.Request) {
	data := &CancelOutstandingRequest{}

	// 1. Decode & Validate (Pastikan passing r *http.Request)
	err := shared.BindAndValidate(r, data) // Asumsi shared butuh r untuk decode JSON
	if err != nil {
		render.Status(r, http.StatusBadRequest) // Gunakan 400 untuk error input
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Format data tidak valid: " + err.Error(),
		})
		return
	}

	// 2. Panggil Service Layer
	// Gunakan '=' karena err sudah dideklarasikan di atas
	err = h.service.CancelOutstanding(r.Context(), data.MInOutID, data.Status)
	if err != nil {
		render.Status(r, http.StatusInternalServerError) // Gunakan 500 untuk error server/DB
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Gagal memproses pembatalan: " + err.Error(),
		})
		return
	}

	// 3. Response Sukses
	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "Pembatalan berhasil diproses",
		Data:    data,
	})
}
