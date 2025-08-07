package service

import (
	"fmt"
	"skh_app/internal/model"
	"strings"
	"time"
	"sync"
)

// SuratRepositoryInterface mendefinisikan fungsi-fungsi database yang dibutuhkan oleh service ini.
type SuratRepositoryInterface interface {
	GetPengaturan() (*model.Pengaturan, error)
	CreateSurat(surat *model.SuratKeteranganHilang, nomorBaru int, nomorSuratLengkap string, tahunSekarang int) (int64, error)
	ResetNomorCounterIfEmpty() error

	// --- TAMBAHKAN 4 METHOD DI BAWAH INI ---
	GetTotalSurat() (int, error)
	GetTotalSuratBulanIni() (int, error)
	GetBarangHilangStats() ([]model.BarangStat, error)
	GetSuratHarianStats() (map[string]int, error)
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

// GetDashboardData mengambil semua data yang diperlukan untuk dashboard dan memprosesnya.
func (s *SuratService) GetDashboardData() (*model.DashboardData, error) {
	// 1. Panggil semua repository yang dibutuhkan.
	// Kita bisa gunakan goroutine agar pemanggilan ke DB berjalan bersamaan untuk efisiensi.
	var totalSurat, totalBulanIni int
	var barangStats []model.BarangStat
	var harianStats map[string]int
	var errTotal, errBulan, errBarang, errHarian error
	
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		totalSurat, errTotal = s.repo.GetTotalSurat()
	}()
	go func() {
		defer wg.Done()
		totalBulanIni, errBulan = s.repo.GetTotalSuratBulanIni()
	}()
	go func() {
		defer wg.Done()
		barangStats, errBarang = s.repo.GetBarangHilangStats()
	}()
	go func() {
		defer wg.Done()
		harianStats, errHarian = s.repo.GetSuratHarianStats()
	}()

	wg.Wait() // Tunggu semua panggilan ke database selesai

	// Cek jika ada error dari salah satu panggilan
	if errTotal != nil || errBulan != nil || errBarang != nil || errHarian != nil {
		// Di sini Anda bisa mencatat error spesifiknya jika perlu
		return nil, fmt.Errorf("gagal mengambil data statistik untuk dashboard")
	}

	// 2. Proses data statistik (ini adalah logika yang kita pindahkan dari handler)
	var statLabels []string
	var statData []int
	for _, stat := range barangStats {
		statLabels = append(statLabels, stat.JenisBarang)
		statData = append(statData, stat.Total)
	}

	var harianLabels []string
	var harianData []int
	hariIni := time.Now().In(s.loc)
	for i := 6; i >= 0; i-- {
		hari := hariIni.AddDate(0, 0, -i)
		tanggalStr := hari.Format("2006-01-02")
		harianLabels = append(harianLabels, hari.Format("02 Jan"))

		if total, ok := harianStats[tanggalStr]; ok {
			harianData = append(harianData, total)
		} else {
			harianData = append(harianData, 0)
		}
	}

	// 3. Kembalikan data dalam satu struct yang rapi dan siap pakai
	dashboardData := &model.DashboardData{
		TotalSurat:    totalSurat,
		TotalBulanIni: totalBulanIni,
		StatLabels:    statLabels,
		StatData:      statData,
		HarianLabels:  harianLabels,
		HarianData:    harianData,
	}

	return dashboardData, nil
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
