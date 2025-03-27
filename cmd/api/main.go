package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"petdoc/apps/auth/login"
	"petdoc/apps/auth/register"
	"petdoc/apps/consultation"
	"petdoc/apps/doctor"
	"petdoc/apps/user"
	"petdoc/internal/config"
	"petdoc/internal/infrastructure/cloudinary"
	"petdoc/internal/infrastructure/database"
	"petdoc/internal/infrastructure/utils/jwt"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables dari .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Gunakan semua core CPU yang tersedia untuk multi-threading
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setup logger dengan format JSON
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(logHandler))

	// Load konfigurasi dari file YAML
	err = config.LoadConfig("./config.yml")
	if err != nil {
		log.Fatalf("Gagal membaca konfigurasi: %v", err)
		panic(err)
	}
	fmt.Println("Berhasil Load Config")
	slog.Info("Konfigurasi berhasil dimuat")

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
	slog.Info("Terhubung ke database")

	// Inisialisasi Cloudinary
	cloudinaryService, err := cloudinary.NewService(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		slog.Error("Gagal inisialisasi Cloudinary", "error", err)
		os.Exit(1)
	}
	slog.Info("Cloudinary berhasil diinisialisasi")

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

	// Inisialisasi JWT service
	jwtService := jwt.NewTokenService(cfg.App.SecretKey)

	// Ambil token expiry dari konfigurasi
	tokenExpiry := time.Duration(cfg.App.ExpireTime) * time.Second

	// Inisialisasi modul register (autentikasi)
	register.InitModule(router, db)
	// Inisialisasi modul login (autentikasi)
	login.InitModule(router, db, jwtService, tokenExpiry)
	// Panggil InitRoutes dengan tokenService(doctor)
	doctor.InitRoutes(router, db, jwtService)
	// Inisialisasi modul user
	user.InitUserModule(router, db)

	// Inisialisasi modul konsultasi dengan middleware auth
	consultation.InitRoutes(router, db, cloudinaryService, jwtService) // Tambahkan ini

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
