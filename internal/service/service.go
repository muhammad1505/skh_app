package service

import (
	"fmt"
	"skh_app/internal/model"
	"strings"
	"time"
)

// SuratRepositoryInterface mendefinisikan fungsi-fungsi database yang dibutuhkan oleh service ini.
type SuratRepositoryInterface interface {
	GetPengaturan() (*model.Pengaturan, error)
	CreateSurat(surat *model.SuratKeteranganHilang, nomorBaru int, nomorSuratLengkap string, tahunSekarang int) (int64, error)
	ResetNomorCounterIfEmpty() error
}

// SuratService adalah service layer yang berisi logika bisnis.
type SuratService struct {
	repo SuratRepositoryInterface
	loc  *time.Location
}

// NewSuratService adalah constructor untuk SuratService.
func NewSuratService(repo SuratRepositoryInterface) *SuratService {
	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		loc = time.Local
	}
	return &SuratService{
		repo: repo,
		loc:  loc,
	}
}

// CreateNewSurat berisi semua logika bisnis untuk membuat surat.
func (s *SuratService) CreateNewSurat(suratData *model.SuratKeteranganHilang) (*model.SuratKeteranganHilang, error) {
	// 1. Validasi awal
	if suratData.PelaporNama == "" || suratData.LokasiHilang == "" {
		return nil, fmt.Errorf("nama pelapor dan lokasi hilang wajib diisi")
	}

	// 2. Periksa dan reset counter
	if err := s.repo.ResetNomorCounterIfEmpty(); err != nil {
		return nil, fmt.Errorf("gagal memeriksa dan mereset nomor surat: %w", err)
	}

	// 3. Ambil pengaturan
	pengaturan, err := s.repo.GetPengaturan()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil pengaturan: %w", err)
	}

	// 4. Logika Bisnis Penomoran
	tanggalSurat := time.Now().In(s.loc)
	nomorBaru := 0
	tahunSekarang := tanggalSurat.Year()

	if tahunSekarang > pengaturan.LastNomorYear {
		nomorBaru = 1
	} else {
		nomorBaru = pengaturan.LastNomorSurat + 1
	}

	// 5. Buat string nomor surat lengkap
	nomorLengkap := generateNomorSurat(pengaturan.FormatNomorSurat, nomorBaru, tanggalSurat)

	// 6. Lengkapi data surat yang akan disimpan
	suratData.NomorSurat = nomorLengkap
	suratData.TanggalSurat = tanggalSurat

	// 7. Panggil repository untuk menyimpan
	suratID, err := s.repo.CreateSurat(suratData, nomorBaru, nomorLengkap, tahunSekarang)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan surat ke database: %w", err)
	}

	suratData.ID = int(suratID)
	return suratData, nil
}


// --- Fungsi Helper (tetap sama) ---
func generateNomorSurat(format string, nomor int, t time.Time) string {
	tahun := fmt.Sprintf("%d", t.Year())
	bulanRomawi := toRoman(int(t.Month()))
	r := strings.NewReplacer(
		"{NO}", fmt.Sprintf("%03d", nomor),
		"{THN}", tahun,
		"{BLN_ROMAWI}", bulanRomawi,
	)
	return r.Replace(format)
}

func toRoman(num int) string {
	romans := map[int]string{
		1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI",
		7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII",
	}
	return romans[num]
}
