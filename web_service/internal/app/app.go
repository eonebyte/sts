package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sts/web_service/internal/auth"
	"sts/web_service/internal/handover"
	"sts/web_service/internal/shared"
	"sts/web_service/internal/shared/config"
	"sts/web_service/internal/shared/db"
	"sts/web_service/internal/shipment"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type App struct {
	Router   *chi.Mux
	Config   *config.Config
	DB       *sqlx.DB
	Logger   *slog.Logger
	WAClient *whatsmeow.Client
}

func NewApp(cfg *config.Config, logger *slog.Logger) (*App, error) {
	var conn *sqlx.DB
	var err error

	conn, err = db.NewOracleDB(cfg.DBDriver, cfg.DBUrl)
	if err != nil {
		return nil, err
	}

	tokenAuth := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	r := chi.NewRouter()
	r.Use(middleware.RequestID) // Generate ID unik tiap request
	r.Use(middleware.RealIP)    // Dapatkan IP asli user
	r.Use(middleware.Recoverer) // Penyelamat kalau ada panic

	// CORS untuk React frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// --- SETUP WHATSMEOW ---
	dbLogger := waLog.Stdout("Database", "DEBUG", true)

	// PERBAIKAN ERROR 1: Tambahkan context.Background() di argumen pertama
	dsn := "file:whatsapp_session.db?_pragma=foreign_keys(1)"
	container, err := sqlstore.New(context.Background(), "sqlite", dsn, dbLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create wa-sqlstore: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get wa-device: %w", err)
	}

	clientLog := waLog.Stdout("WhatsApp", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// --- LOGIKA LOGIN / CONNECT ---
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, err
		}

		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					fmt.Println("\n--- SCAN QR INI DENGAN HP ANDA ---")
					// PERBAIKAN ERROR 2: Gunakan os.Stdout, bukan fmt.Stdout
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					fmt.Println("----------------------------------\n")
				} else {
					fmt.Println("WA Login Event:", evt.Event)
				}
			}
		}()
	} else {
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}

	// Testing Group = "120363407477018375@g.us"
	// STS Group = "120363421649034694@g.us"
	waGroupID := "120363407477018375@g.us"
	notifSvc := handover.NewWANotificationService(client, waGroupID)

	// REPO
	authRepo := auth.NewOraRepository(conn)
	shipmentRepo := shipment.NewOraRepository(conn)
	handoverRepo := handover.NewOraRepository(conn)

	// SERVICE & HANDLER
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService, tokenAuth)

	shipmentService := shipment.NewService(shipmentRepo)
	shipmentHandler := shipment.NewHandler(shipmentService)

	handoverService := handover.NewService(handoverRepo, notifSvc)
	handoverHandler := handover.NewHandler(handoverService)

	workDir, _ := os.Getwd()
	// Pastikan folder ini mengarah ke root "uploads"
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))

	// Register route /uploads/* agar bisa diakses browser
	setupFileServer(r, "/uploads", filesDir)

	// Public Routes
	r.Group(func(r chi.Router) {
		// Auth Routes
		authHandler.RegisterPublicRoutes(r)
	})

	// Private Route
	r.Group(func(r chi.Router) {

		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		// Auth Routes
		authHandler.RegisterProtectedRoutes(r)

		shipmentHandler.RegisterProtectedRoutes(r)
		handoverHandler.RegisterProtectedRoutes(r)
	})

	return &App{
		Router:   r,
		Config:   cfg,
		DB:       conn,
		Logger:   logger,
		WAClient: client,
	}, nil
}

func (a *App) Run() error {
	srv := &http.Server{
		Addr:    a.Config.Port,
		Handler: a.Router,
	}

	a.listRoutes()

	// Panggil Utils untuk menjalankan server + graceful shutdown
	return shared.RunWithGracefulShutdown(srv, 5*time.Second, func() error {
		if a.DB != nil {
			return a.DB.Close()
		}
		if a.WAClient != nil {
			a.WAClient.Disconnect()
			fmt.Println("WhatsApp disconnected.")
		}
		return nil
	})
}

func (a *App) listRoutes() {
	a.Logger.Info("Registered routes:")
	chi.Walk(a.Router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("- %s %s", method, route)
		return nil
	})
}

// Helper function untuk Static File Server
func setupFileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
