package auth

import (
	"errors"
	"log"
	"net/http"
	"sts/web_service/internal/shared"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

type handler struct {
	service   Service
	tokenAuth *jwtauth.JWTAuth
}

func NewHandler(s Service, tokenAuth *jwtauth.JWTAuth) *handler {
	return &handler{service: s, tokenAuth: tokenAuth}
}

func (h *handler) RegisterPublicRoutes(r chi.Router) {

	r.Route("/auth", func(r chi.Router) {
		// Rute publik, tidak memerlukan token
		r.Post("/register", h.register)
		r.Post("/login", h.login)

	})
}

func (h *handler) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/me", h.me)
	r.Post("/refresh", h.refresh)
}

func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := shared.BindAndValidate(r, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	err := h.service.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, APIResponse{
				Success: false,
				Message: "Username already exists",
			})
			return
		}

		log.Printf("[ERROR] Register failure: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
	})
}

// login menangani permintaan login dan mengembalikan token pair.
func (h *handler) login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest
	if err := shared.BindAndValidate(r, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	tokens, err := h.service.Login(r.Context(), req.Username, req.Password, h.tokenAuth)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, APIResponse{
				Success: false,
				Message: "Invalid username or password",
			})
			return
		}

		log.Printf("[ERROR] Login failure: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Data:    tokens,
	})
}

// me adalah endpoint yang dilindungi untuk mendapatkan info user dari token.
func (h *handler) me(w http.ResponseWriter, r *http.Request) {
	// Middleware sudah memvalidasi token, kita bisa langsung ambil claims dari context.
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	// Gunakan 'sub' sebagai standar user ID
	userIDStr, _ := claims["sub"].(string)
	userTitleStr, _ := claims["title"].(string)
	userNameStr, _ := claims["username"].(string)

	data := map[string]interface{}{
		"message":  "welcome",
		"user_id":  userIDStr,
		"title":    userTitleStr,
		"username": userNameStr,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Data:    data,
	})
}

// refresh menangani pembuatan access token baru menggunakan refresh token.
func (h *handler) refresh(w http.ResponseWriter, r *http.Request) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	newAccessToken, err := h.service.RefreshToken(r.Context(), claims, h.tokenAuth)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, APIResponse{
				Success: false,
				Message: "Invalid token",
			})
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, APIResponse{
			Success: false,
			Message: "Internal server error",
		})
		return
	}

	data := map[string]string{
		"access_token": newAccessToken,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, APIResponse{
		Success: true,
		Message: "OK",
		Data:    data,
	})
}
