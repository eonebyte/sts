package tms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type handler struct {
	service Service
}

func NewHandler(s Service) *handler {
	return &handler{service: s}
}
func (h *handler) RegisterPublicRoutes(r chi.Router) {
	r.Route("/tms", func(r chi.Router) {
		r.Get("/drivers/capital", h.GetDrivers)
		r.Get("/list/sj/bydriver", h.ShipmentByDriver)
		r.Get("/customer/logs", h.GetCustomerLogs)
		r.Post("/customer/logs/update", h.UpdateCustomerLog)
	})
}

func (h *handler) GetDrivers(w http.ResponseWriter, r *http.Request) {
	// Mengambil parameter dari URL: /shipments/drivers/search?name=budi
	searchKey := r.URL.Query().Get("searchTerm")

	if searchKey == "" {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Wajib memasukan nama driver",
		})
		return
	}

	drivers, err := h.service.SearchDriver(r.Context(), searchKey)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Jika data tidak ditemukan, kirim array kosong bukan null
	if drivers == nil {
		drivers = []SearchDriver{}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: false,
		Message: "Success fetch drivers",
		Data:    drivers,
	})
}

func (h *handler) ShipmentByDriver(w http.ResponseWriter, r *http.Request) {
	// Mengambil parameter dari URL: /shipments/drivers/search?name=budi
	driverIDRaw := r.URL.Query().Get("driver_id")

	if driverIDRaw == "" {
		render.Status(r, http.StatusBadRequest) // Gunakan 400 untuk kesalahan input client
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "ID Driver kosong",
		})
		return
	}

	driverID, err := strconv.ParseInt(driverIDRaw, 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "ID Driver harus berupa angka valid",
		})
		return
	}

	shipments, err := h.service.ShipmentByDriver(r.Context(), driverID)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	tnkbNo := ""
	if len(shipments) > 0 {
		tnkbNo = shipments[0].TNKBNo
	} else {
		shipments = []ShipmentByDriver{} // Pastikan kirim [] bukan null
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"message": "fetch successfully",
		"data": map[string]interface{}{
			"success": true,
			"tnkb_no": tnkbNo,
			"data":    shipments,
		},
	})
}

func (h *handler) GetCustomerLogs(w http.ResponseWriter, r *http.Request) {
	tmsIDStr := r.URL.Query().Get("tms_id")
	tmsID, _ := strconv.Atoi(tmsIDStr)

	list, err := h.service.GetCustomerLogs(r.Context(), int64(tmsID))
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Data:    list,
	})
}

func (h *handler) UpdateCustomerLog(w http.ResponseWriter, r *http.Request) {
	var req UpdateLogRequest

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]interface{}{
			"success": false,
			"message": "Payload tidak valid",
		})
		return
	}

	fmt.Printf("test EventID : ", req.EventID)
	fmt.Printf("test EventTime : ", req.EventTime)
	fmt.Printf("test Notes : ", req.Notes)

	// Panggil Service
	err := h.service.UpdateLog(r.Context(), req.EventID, req.EventTime, req.Notes)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]interface{}{
			"success": false,
			"message": "Gagal update database: " + err.Error(),
		})
		return
	}

	// Response Sukses
	render.JSON(w, r, map[string]interface{}{
		"success": true,
		"message": "Data berhasil diperbarui",
	})
}
