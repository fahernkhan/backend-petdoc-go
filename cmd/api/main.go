package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"petdoc/apps/auth/register"
	"petdoc/internal/config"
	"petdoc/internal/infrastructure/database"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// Gunakan semua core CPU yang tersedia untuk multi-threading
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup logger dengan format JSON
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(logHandler))

	// Load konfigurasi dari file YAML
	err := config.LoadConfig("./config.yml")
	if err != nil {
		log.Fatalf("Gagal membaca konfigurasi: %v", err)
		panic(err)
	}
	fmt.Println("Berhasil Load Config")

	// Ambil konfigurasi yang sudah dimuat
	cfg := config.GetConfig()

	// Setup koneksi database PostgreSQL
	db, err := database.ConnectPostgres(cfg.DB)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Connected to database successfully")

	// Inisialisasi router Gin
	router := gin.Default()

	// Tambahkan middleware CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://frontend-petdoc.vercel.app"}, // Ganti dengan origin frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Inisialisasi modul register (autentikasi)
	register.InitModule(router, db)

	// Start server dengan port dari konfigurasi
	appPort := config.GetConfig().App.Port
	slog.Info("Starting server", slog.String("port", appPort))
	if err := router.Run(":" + appPort); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// Cek apakah nilai-nilai sudah terbaca dengan benar
	fmt.Println("Nama Aplikasi:", cfg.App.Name)
	fmt.Println("Port Aplikasi:", cfg.App.Port)
	fmt.Println("Database Host:", cfg.DB.Host)
	fmt.Println("Database Port:", cfg.DB.Port)
	fmt.Println("Username Database:", cfg.DB.Username)
	fmt.Println("Max Open Conns:", cfg.DB.MaxOpenConns)

}
