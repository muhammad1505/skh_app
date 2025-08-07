package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"skh_app/internal/model"
	"time"
)

// PengaturanRepositoryInterface mendefinisikan fungsi yang dibutuhkan dari database
type PengaturanRepositoryInterface interface {
	GetPengaturan() (*model.Pengaturan, error)
	UpdatePengaturan(*model.Pengaturan) error
}

// PengaturanService menangani logika bisnis untuk pengaturan
type PengaturanService struct {
	repo PengaturanRepositoryInterface
}

// NewPengaturanService adalah constructor untuk service pengaturan
func NewPengaturanService(repo PengaturanRepositoryInterface) *PengaturanService {
	return &PengaturanService{repo: repo}
}

// UpdatePengaturan berisi logika untuk update data dan menyimpan file logo
func (s *PengaturanService) UpdatePengaturan(p *model.Pengaturan, logoFile multipart.File, logoHandler *multipart.FileHeader) (*model.Pengaturan, error) {
	// 1. Logika penyimpanan file
	if logoFile != nil {
		defer logoFile.Close()
		
		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), logoHandler.Filename)
		uploadPath := filepath.Join("web", "static", "uploads", newFileName)

		// Buat file di server
		dst, err := os.Create(uploadPath)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat file di server: %w", err)
		}
		defer dst.Close()

		// Salin file yang diupload
		if _, err := io.Copy(dst, logoFile); err != nil {
			return nil, fmt.Errorf("gagal menyalin file: %w", err)
		}
		
		// Set path logo baru untuk disimpan ke database
		p.LogoPath = "/static/uploads/" + newFileName
	}
	
	// Jika tahun belum diset, set ke tahun sekarang
	if p.LastNomorYear == 0 {
		p.LastNomorYear = time.Now().Year()
	}
	
	// 2. Panggil repository untuk menyimpan semua perubahan ke database
	if err := s.repo.UpdatePengaturan(p); err != nil {
		return nil, fmt.Errorf("gagal menyimpan pengaturan ke database: %w", err)
	}

	return p, nil
}
