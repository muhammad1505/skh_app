package repository

import (
	"database/sql"
	"skh_app/internal/model"
	"strings" // <-- PERBAIKAN DI SINI
	"time"
)

// Struct utama untuk semua interaksi database
type SuratRepository struct {
	DB *sql.DB
}

// Constructor untuk membuat instance repository baru
func NewSuratRepository(db *sql.DB) *SuratRepository {
	return &SuratRepository{DB: db}
}

func (r *SuratRepository) ResetNomorCounterIfEmpty() error {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(id) FROM surat").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		loc, _ := time.LoadLocation("Asia/Makassar")
		tahunSekarang := time.Now().In(loc).Year()

		tx, err := r.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Reset nomor urut di pengaturan
		_, err = tx.Exec("UPDATE pengaturan SET last_nomor_surat = 0, last_nomor_year = ? WHERE id = 1", tahunSekarang)
		if err != nil {
			return err
		}

		// Reset ID counter di tabel surat
		_, err = tx.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name = 'surat'")
		if err != nil {
			if !strings.Contains(err.Error(), "no such table") {
				return err
			}
		}

		return tx.Commit()
	}
	return nil
}

// --- FUNGSI PENGATURAN ---

func (r *SuratRepository) GetPengaturan() (*model.Pengaturan, error) {
	pengaturan := &model.Pengaturan{
		PejabatDetails:  &model.Petugas{},
		PenerimaDetails: &model.Petugas{},
	}
	query := `
		SELECT
			p.id, p.kop_surat_1, p.kop_surat_2, p.kop_surat_3, p.logo_path,
			p.format_nomor_surat, p.last_nomor_surat, p.last_nomor_year,
			p.pejabat_id, p.penerima_id, p.wilayah, p.nama_kantor,
			pejabat.nama, pejabat.pangkat, pejabat.nrp, pejabat.jabatan,
			penerima.nama, penerima.pangkat, penerima.nrp, penerima.jabatan
		FROM pengaturan p
		LEFT JOIN petugas AS pejabat ON p.pejabat_id = pejabat.id
		LEFT JOIN petugas AS penerima ON p.penerima_id = penerima.id
		WHERE p.id = 1
	`
	var kop1, kop2, kop3, logo, format, wilayah, kantor sql.NullString
	var lastNomor, lastNomorYear, pejabatID, penerimaID sql.NullInt64
	var pejNama, pejPangkat, pejNRP, pejJabatan sql.NullString
	var penNama, penPangkat, penNRP, penJabatan sql.NullString

	err := r.DB.QueryRow(query).Scan(
		&pengaturan.ID, &kop1, &kop2, &kop3, &logo, &format, &lastNomor, &lastNomorYear,
		&pejabatID, &penerimaID, &wilayah, &kantor,
		&pejNama, &pejPangkat, &pejNRP, &pejJabatan,
		&penNama, &penPangkat, &penNRP, &penJabatan,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	pengaturan.KopSurat1 = kop1.String
	pengaturan.KopSurat2 = kop2.String
	pengaturan.KopSurat3 = kop3.String
	pengaturan.LogoPath = logo.String
	pengaturan.FormatNomorSurat = format.String
	pengaturan.Wilayah = wilayah.String
	pengaturan.NamaKantor = kantor.String
	pengaturan.LastNomorSurat = int(lastNomor.Int64)
	pengaturan.LastNomorYear = int(lastNomorYear.Int64)
	pengaturan.PejabatID = int(pejabatID.Int64)
	pengaturan.PenerimaID = int(penerimaID.Int64)
	pengaturan.PejabatDetails.ID = int(pejabatID.Int64)
	pengaturan.PejabatDetails.Nama = pejNama.String
	pengaturan.PejabatDetails.Pangkat = pejPangkat.String
	pengaturan.PejabatDetails.NRP = pejNRP.String
	pengaturan.PejabatDetails.Jabatan = pejJabatan.String
	pengaturan.PenerimaDetails.ID = int(penerimaID.Int64)
	pengaturan.PenerimaDetails.Nama = penNama.String
	pengaturan.PenerimaDetails.Pangkat = penPangkat.String
	pengaturan.PenerimaDetails.NRP = penNRP.String
	pengaturan.PenerimaDetails.Jabatan = penJabatan.String

	return pengaturan, nil
}

func (r *SuratRepository) UpdatePengaturan(p *model.Pengaturan) error {
	var query string
	var args []interface{}
	baseQuery := `
		UPDATE pengaturan SET 
			kop_surat_1 = ?, kop_surat_2 = ?, kop_surat_3 = ?, 
			format_nomor_surat = ?, pejabat_id = ?, penerima_id = ?,
			wilayah = ?, nama_kantor = ?, last_nomor_surat = ?, last_nomor_year = ?`
	args = []interface{}{
		p.KopSurat1, p.KopSurat2, p.KopSurat3, p.FormatNomorSurat,
		p.PejabatID, p.PenerimaID, p.Wilayah, p.NamaKantor, p.LastNomorSurat, p.LastNomorYear,
	}

	if p.LogoPath != "" {
		query = baseQuery + ", logo_path = ? WHERE id = 1"
		args = append(args, p.LogoPath)
	} else {
		query = baseQuery + " WHERE id = 1"
	}
	_, err := r.DB.Exec(query, args...)
	return err
}

// --- FUNGSI SURAT ---

func (r *SuratRepository) CreateSurat(surat *model.SuratKeteranganHilang, nomorBaru int, nomorSuratLengkap string, tahunSekarang int) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE pengaturan SET last_nomor_surat = ?, last_nomor_year = ? WHERE id = 1", nomorBaru, tahunSekarang)
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(`
		INSERT INTO surat (nomor_surat, tanggal_surat, pelapor_nama, pelapor_ttl, pelapor_agama, pelapor_kelamin, pelapor_pekerjaan, pelapor_alamat, lokasi_hilang)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		nomorSuratLengkap, surat.TanggalSurat, surat.PelaporNama, surat.PelaporTTL, surat.PelaporAgama, surat.PelaporKelamin, surat.PelaporPekerjaan, surat.PelaporAlamat, surat.LokasiHilang,
	)
	if err != nil {
		return 0, err
	}

	suratID, _ := res.LastInsertId()

	for _, barang := range surat.BarangHilang {
		_, err = tx.Exec("INSERT INTO barang (surat_id, jenis_barang, data) VALUES (?, ?, ?)", suratID, barang.JenisBarang, barang.Data)
		if err != nil {
			return 0, err
		}
	}

	return suratID, tx.Commit()
}

func (r *SuratRepository) UpdateSurat(surat *model.SuratKeteranganHilang) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE surat SET 
		tanggal_surat = ?, pelapor_nama = ?, pelapor_ttl = ?, pelapor_agama = ?, 
		pelapor_kelamin = ?, pelapor_pekerjaan = ?, pelapor_alamat = ?, lokasi_hilang = ? 
		WHERE id = ?`,
		surat.TanggalSurat, surat.PelaporNama, surat.PelaporTTL, surat.PelaporAgama,
		surat.PelaporKelamin, surat.PelaporPekerjaan, surat.PelaporAlamat, surat.LokasiHilang, surat.ID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM barang WHERE surat_id = ?", surat.ID)
	if err != nil {
		return err
	}

	for _, barang := range surat.BarangHilang {
		_, err = tx.Exec("INSERT INTO barang (surat_id, jenis_barang, data) VALUES (?, ?, ?)", surat.ID, barang.JenisBarang, barang.Data)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SuratRepository) DeleteSurat(id int) error {
	_, err := r.DB.Exec("DELETE FROM surat WHERE id = ?", id)
	return err
}

func (r *SuratRepository) GetSuratByID(id int) (*model.SuratKeteranganHilang, error) {
	s := &model.SuratKeteranganHilang{}
	querySurat := `SELECT id, nomor_surat, tanggal_surat, pelapor_nama, pelapor_ttl, pelapor_agama, pelapor_kelamin, pelapor_pekerjaan, pelapor_alamat, lokasi_hilang FROM surat WHERE id = ?`

	err := r.DB.QueryRow(querySurat, id).Scan(
		&s.ID, &s.NomorSurat, &s.TanggalSurat, &s.PelaporNama, &s.PelaporTTL, &s.PelaporAgama,
		&s.PelaporKelamin, &s.PelaporPekerjaan, &s.PelaporAlamat, &s.LokasiHilang,
	)
	if err != nil {
		return nil, err
	}

	queryBarang := `SELECT id, jenis_barang, data FROM barang WHERE surat_id = ?`
	rows, err := r.DB.Query(queryBarang, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b model.Barang
		if err := rows.Scan(&b.ID, &b.JenisBarang, &b.Data); err != nil {
			return nil, err
		}
		s.BarangHilang = append(s.BarangHilang, b)
	}

	return s, nil
}

func (r *SuratRepository) GetAllSurat(searchTerm string) ([]model.SuratKeteranganHilang, error) {
	var surats []model.SuratKeteranganHilang
	query := `SELECT id, nomor_surat, tanggal_surat, pelapor_nama FROM surat`
	args := []interface{}{}
	if searchTerm != "" {
		query += " WHERE pelapor_nama LIKE ? OR nomor_surat LIKE ?"
		likeTerm := "%" + searchTerm + "%"
		args = append(args, likeTerm, likeTerm)
	}
	query += " ORDER BY id DESC"
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s model.SuratKeteranganHilang
		if err := rows.Scan(&s.ID, &s.NomorSurat, &s.TanggalSurat, &s.PelaporNama); err != nil {
			return nil, err
		}
		surats = append(surats, s)
	}
	return surats, nil
}

func (r *SuratRepository) GetTotalSurat() (int, error) {
	var count int
	err := r.DB.QueryRow("SELECT COUNT(id) FROM surat").Scan(&count)
	return count, err
}

func (r *SuratRepository) GetTotalSuratBulanIni() (int, error) {
	var count int
	loc, _ := time.LoadLocation("Asia/Makassar")
	now := time.Now().In(loc)
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	firstDayOfNextMonth := firstDayOfMonth.AddDate(0, 1, 0)
	query := "SELECT COUNT(id) FROM surat WHERE tanggal_surat >= ? AND tanggal_surat < ?"
	err := r.DB.QueryRow(query, firstDayOfMonth, firstDayOfNextMonth).Scan(&count)
	return count, err
}

func (r *SuratRepository) GetBarangHilangStats() ([]model.BarangStat, error) {
	var stats []model.BarangStat
	query := `SELECT jenis_barang, COUNT(*) as total FROM barang GROUP BY jenis_barang ORDER BY total DESC`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.BarangStat
		if err := rows.Scan(&stat.JenisBarang, &stat.Total); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// Ganti fungsi lama dengan yang ini
func (r *SuratRepository) GetSuratHarianStats() (map[string]int, error) {
	stats := make(map[string]int)
	// PERBAIKAN: Tambahkan modifier timezone di SELECT strftime
	query := `
		SELECT strftime('%Y-%m-%d', tanggal_surat, 'localtime', '+8 hours') as tanggal, COUNT(id) as total 
		FROM surat 
		WHERE tanggal_surat >= date('now', '-6 days', 'localtime', '+8 hours') 
		GROUP BY tanggal
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tanggal string
		var total int
		if err := rows.Scan(&tanggal, &total); err != nil {
			return nil, err
		}
		stats[tanggal] = total
	}
	return stats, nil
}


// --- FUNGSI MANAJEMEN PETUGAS ---

func (r *SuratRepository) CreatePetugas(p *model.Petugas) error {
	query := `INSERT INTO petugas (nama, pangkat, nrp, jabatan, tipe) VALUES (?, ?, ?, ?, ?)`
	_, err := r.DB.Exec(query, p.Nama, p.Pangkat, p.NRP, p.Jabatan, p.Tipe)
	return err
}

func (r *SuratRepository) GetPetugasByID(id int) (*model.Petugas, error) {
	p := &model.Petugas{}
	query := `SELECT id, nama, pangkat, nrp, jabatan, tipe FROM petugas WHERE id = ?`
	err := r.DB.QueryRow(query, id).Scan(&p.ID, &p.Nama, &p.Pangkat, &p.NRP, &p.Jabatan, &p.Tipe)
	return p, err
}

func (r *SuratRepository) GetAllPetugas() ([]model.Petugas, error) {
	var allPetugas []model.Petugas
	query := `SELECT id, nama, pangkat, nrp, jabatan, tipe FROM petugas ORDER BY nama ASC`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p model.Petugas
		if err := rows.Scan(&p.ID, &p.Nama, &p.Pangkat, &p.NRP, &p.Jabatan, &p.Tipe); err != nil {
			return nil, err
		}
		allPetugas = append(allPetugas, p)
	}
	return allPetugas, nil
}

func (r *SuratRepository) GetPetugasByTipe(tipe string) ([]model.Petugas, error) {
	var allPetugas []model.Petugas
	query := `SELECT id, nama, pangkat, nrp, jabatan, tipe FROM petugas WHERE tipe = ? ORDER BY nama ASC`
	rows, err := r.DB.Query(query, tipe)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p model.Petugas
		if err := rows.Scan(&p.ID, &p.Nama, &p.Pangkat, &p.NRP, &p.Jabatan, &p.Tipe); err != nil {
			return nil, err
		}
		allPetugas = append(allPetugas, p)
	}
	return allPetugas, nil
}

func (r *SuratRepository) UpdatePetugas(p *model.Petugas) error {
	query := `UPDATE petugas SET nama = ?, pangkat = ?, nrp = ?, jabatan = ?, tipe = ? WHERE id = ?`
	_, err := r.DB.Exec(query, p.Nama, p.Pangkat, p.NRP, p.Jabatan, p.Tipe, p.ID)
	return err
}

func (r *SuratRepository) DeletePetugas(id int) error {
	_, err := r.DB.Exec("DELETE FROM petugas WHERE id = ?", id)
	return err
}
