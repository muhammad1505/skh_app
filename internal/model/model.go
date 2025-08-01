package model

import "time"

// Petugas menyimpan data petugas kepolisian
type Petugas struct {
	ID      int    `db:"id"`
	Nama    string `db:"nama"`
	Pangkat string `db:"pangkat"`
	NRP     string `db:"nrp"`
	Jabatan string `db:"jabatan"`
	Tipe    string `db:"tipe"` // Tipe: "Pejabat" atau "Penerima"
}

// Pengaturan menyimpan semua konfigurasi aplikasi
type Pengaturan struct {
	ID               int    `db:"id"`
	KopSurat1        string `db:"kop_surat_1"`
	KopSurat2        string `db:"kop_surat_2"`
	KopSurat3        string `db:"kop_surat_3"`
	LogoPath         string `db:"logo_path"`
	FormatNomorSurat string `db:"format_nomor_surat"`
	LastNomorSurat   int    `db:"last_nomor_surat"`
	LastNomorYear    int    `db:"last_nomor_year"`
	PejabatID        int    `db:"pejabat_id"`
	PenerimaID       int    `db:"penerima_id"`
	Wilayah          string `db:"wilayah"`
	NamaKantor       string `db:"nama_kantor"`

	PejabatDetails  *Petugas
	PenerimaDetails *Petugas
}

// SuratKeteranganHilang adalah data utama surat
type SuratKeteranganHilang struct {
	ID               int       `db:"id"`
	NomorSurat       string    `db:"nomor_surat"`
	TanggalSurat     time.Time `db:"tanggal_surat"`
	PelaporNama      string    `db:"pelapor_nama"`
	PelaporTTL       string    `db:"pelapor_ttl"`
	PelaporAgama     string    `db:"pelapor_agama"`
	PelaporKelamin   string    `db:"pelapor_kelamin"`
	PelaporPekerjaan string    `db:"pelapor_pekerjaan"`
	PelaporAlamat    string    `db:"pelapor_alamat"`
	LokasiHilang     string    `db:"lokasi_hilang"`
	BarangHilang     []Barang
	CreatedAt        time.Time `db:"created_at"`
}

// Barang yang hilang dalam satu surat
type Barang struct {
	ID          int    `db:"id"`
	SuratID     int    `db:"surat_id"`
	JenisBarang string `db:"jenis_barang"`
	Data        string `db:"data"`
}

// PageData adalah struct untuk mengirim data ke template
type PageData struct {
	Surat *SuratKeteranganHilang
	Error string
}

// BarangStat untuk menampung hasil statistik
type BarangStat struct {
	JenisBarang string
	Total       int
}

// DashboardData untuk statistik di dashboard
type DashboardData struct {
	TotalSurat       int
	TotalBulanIni    int
	PesanPenyambutan string
	StatLabels       []string // Untuk Pie Chart Kategori
	StatData         []int
	HarianLabels     []string // Untuk Bar Chart Harian
	HarianData       []int
}

