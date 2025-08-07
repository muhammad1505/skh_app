package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"skh_app/internal/handler"
	"skh_app/internal/repository"
	"skh_app/internal/service" // <-- Pastikan import ini ada
	"skh_app/web"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/browser"
)

func main() {
	db, err := repository.ConnectDatabase()
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}
	defer db.Close()

	// --- BAGIAN INISIALISASI FINAL ---
	suratRepo := repository.NewSuratRepository(db)
	
	// Inisialisasi kedua service dengan repository yang sama
	suratService := service.NewSuratService(suratRepo)
	pengaturanService := service.NewPengaturanService(suratRepo)

	// Suntikkan semua dependensi ke Handler
	h := handler.NewHandler(suratRepo, suratService, pengaturanService)
	// --- AKHIR BAGIAN INISIALISASI FINAL ---

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Gagal mendapatkan working directory: %v", err)
	}
	uploadsDir := http.Dir(filepath.Join(wd, "web", "static", "uploads"))

	r.Handle("/static/uploads/*", http.StripPrefix("/static/uploads/", http.FileServer(uploadsDir)))
	r.Handle("/static/*", http.FileServer(http.FS(web.Files)))

	r.Get("/", h.Dashboard)

	r.Route("/surat", func(r chi.Router) {
		r.Get("/", h.SuratList)
		r.Get("/baru", h.SuratFormNew)
		r.Post("/baru", h.SuratCreate)
		r.Get("/print/{id}", h.SuratPrint)
		r.Get("/edit/{id}", h.SuratFormEdit)
		r.Post("/edit/{id}", h.SuratUpdate)
		r.Get("/hapus/{id}", h.SuratDelete)
	})

	r.Route("/petugas", func(r chi.Router) {
		r.Get("/", h.PetugasList)
		r.Get("/baru", h.PetugasFormNew)
		r.Post("/baru", h.PetugasCreate)
		r.Get("/edit/{id}", h.PetugasFormEdit)
		r.Post("/edit/{id}", h.PetugasUpdate)
		r.Get("/hapus/{id}", h.PetugasDelete)
	})

	r.Route("/pengaturan", func(r chi.Router) {
		r.Get("/", h.PengaturanForm)
		r.Post("/", h.PengaturanUpdate)
	})

	go func() {
		fmt.Println("Server berjalan di http://localhost:8080")
		if err := http.ListenAndServe(":8080", r); err != nil {
			log.Fatalf("Gagal menjalankan server: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	fmt.Println("Membuka browser...")
	err = browser.OpenURL("http://localhost:8080")
	if err != nil {
		log.Printf("Gagal membuka browser secara otomatis: %v", err)
		fmt.Println("Silakan buka http://localhost:8080 secara manual di browser Anda.")
	}

	select {}
}
