package handover

import (
	"log"
	"net/http"
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
	r.Route("/handover", func(r chi.Router) {
		r.Post("/init", h.Init)        // Untuk scan pertama kali (Create)
		r.Post("/process", h.Handover) // Untuk scan berikutnya (Update)
	})
}

func (h *handler) Init(w http.ResponseWriter, r *http.Request) {
	var req HandoverRequest
	if err := shared.BindAndValidate(r, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Panggil service yang menangani batch
	if err := h.service.ProcessInit(r.Context(), req); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Srv: failed ProcessInit()",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "Init Ok",
	})
}

func (h *handler) Handover(w http.ResponseWriter, r *http.Request) {
	var req HandoverRequest
	if err := shared.BindAndValidate(r, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Memanggil ProcessHandover (Update)
	if err := h.service.ProcessHandover(r.Context(), req); err != nil {
		log.Printf(
			"[SERVICE] path=%s method=%s error=%v",
			r.URL.Path,
			r.Method,
			err,
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		Message: "Hanover Process Ok",
	})
}
