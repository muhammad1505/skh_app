package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"skh_app/internal/model"
	// "skh_app/internal/service" <-- service sudah dipanggil via h.Service
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Dashboard menampilkan halaman utama dengan statistik
// (Fungsi ini akan kita refactor berikutnya)
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	total, _ := h.Repo.GetTotalSurat()
	bulanIni, _ := h.Repo.GetTotalSuratBulanIni()
	stats, _ := h.Repo.GetBarangHilangStats()
	harianStats, _ := h.Repo.GetSuratHarianStats()

	// Proses data kategori untuk Pie Chart
	var statLabels []string
	var statData []int
	for _, s := range stats {
		statLabels = append(statLabels, s.JenisBarang)
		statData = append(statData, s.Total)
	}

	// Proses data harian untuk Bar Chart
	var harianLabels []string
	var harianData []int
	loc, _ := time.LoadLocation("Asia/Makassar")
	hariIni := time.Now().In(loc)
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

	data := model.DashboardData{
		TotalSurat:    total,
		TotalBulanIni: bulanIni,
		StatLabels:    statLabels,
		StatData:      statData,
		HarianLabels:  harianLabels,
		HarianData:    harianData,
	}
	h.render(w, "dashboard.html", data)
}

// SuratList menampilkan semua surat yang telah dibuat, dengan fitur pencarian
func (h *Handler) SuratList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	surats, err := h.Repo.GetAllSurat(query)
	if err != nil {
		http.Error(w, "Gagal mengambil daftar surat", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Surats": surats,
		"Query":  query,
	}
	h.render(w, "surat_list.html", data)
}

// SuratFormNew menampilkan formulir untuk membuat surat baru
func (h *Handler) SuratFormNew(w http.ResponseWriter, r *http.Request) {
	h.render(w, "surat_form.html", model.PageData{})
}

// ====================================================================
// FUNGSI INI SUDAH DI-REFACTOR
// ====================================================================
// SuratCreate memproses data dari form dan membuat surat baru
func (h *Handler) SuratCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal mem-parsing form", http.StatusBadRequest)
		return
	}

	// 1. Kumpulkan data dari form ke dalam struct model
	tempatLahir := r.FormValue("tempat_lahir")
	tanggalLahir := r.FormValue("tanggal_lahir")
	ttl := fmt.Sprintf("%s, %s", tempatLahir, tanggalLahir)

	surat := &model.SuratKeteranganHilang{
		PelaporNama:      r.FormValue("pelapor_nama"),
		PelaporTTL:       ttl,
		PelaporAgama:     r.FormValue("pelapor_agama"),
		PelaporKelamin:   r.FormValue("pelapor_kelamin"),
		PelaporPekerjaan: r.FormValue("pelapor_pekerjaan"),
		PelaporAlamat:    r.FormValue("pelapor_alamat"),
		LokasiHilang:     r.FormValue("lokasi_hilang"),
	}

	jenisBarangList := r.Form["barang_jenis[]"]
	dataBarangList := r.Form["barang_data[]"]
	for i, jenis := range jenisBarangList {
		if jenis != "" && i < len(dataBarangList) {
			surat.BarangHilang = append(surat.BarangHilang, model.Barang{
				JenisBarang: jenis,
				Data:        dataBarangList[i],
			})
		}
	}

	// 2. Panggil Service untuk menjalankan SEMUA logika bisnis
	createdSurat, err := h.Service.CreateNewSurat(surat)
	if err != nil {
		// Jika ada error dari service, tampilkan di form agar pengguna bisa memperbaiki
		data := model.PageData{
			Surat: surat, // Kirim kembali data yang sudah diisi pengguna
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest) // Set status code yang sesuai
		h.render(w, "surat_form.html", data)
		return
	}

	// 3. Jika berhasil, redirect seperti biasa
	http.Redirect(w, r, fmt.Sprintf("/surat?status=success_create&new_id=%d", createdSurat.ID), http.StatusSeeOther)
}

// SuratFormEdit menampilkan form yang sudah terisi data untuk diubah
func (h *Handler) SuratFormEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	surat, err := h.Repo.GetSuratByID(id)
	if err != nil {
		http.Error(w, "Surat tidak ditemukan", http.StatusNotFound)
		return
	}
	data := model.PageData{Surat: surat}
	h.render(w, "surat_form.html", data)
}

// SuratUpdate memproses data dari form edit
func (h *Handler) SuratUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal mem-parsing form", http.StatusBadRequest)
		return
	}

	existingSurat, err := h.Repo.GetSuratByID(id)
	if err != nil {
		http.Error(w, "Surat yang akan diupdate tidak ditemukan", http.StatusNotFound)
		return
	}

	if r.FormValue("pelapor_nama") == "" || r.FormValue("lokasi_hilang") == "" {
		data := model.PageData{
			Surat: existingSurat,
			Error: "Nama Pelapor dan Lokasi Hilang wajib diisi.",
		}
		h.render(w, "surat_form.html", data)
		return
	}

	tempatLahir := r.FormValue("tempat_lahir")
	tanggalLahir := r.FormValue("tanggal_lahir")
	ttl := fmt.Sprintf("%s, %s", tempatLahir, tanggalLahir)

	surat := model.SuratKeteranganHilang{
		ID:               id,
		TanggalSurat:     existingSurat.TanggalSurat, // Pertahankan tanggal surat asli
		PelaporNama:      r.FormValue("pelapor_nama"),
		PelaporTTL:       ttl,
		PelaporAgama:     r.FormValue("pelapor_agama"),
		PelaporKelamin:   r.FormValue("pelapor_kelamin"),
		PelaporPekerjaan: r.FormValue("pelapor_pekerjaan"),
		PelaporAlamat:    r.FormValue("pelapor_alamat"),
		LokasiHilang:     r.FormValue("lokasi_hilang"),
	}

	jenisBarangList := r.Form["barang_jenis[]"]
	dataBarangList := r.Form["barang_data[]"]
	for i, jenis := range jenisBarangList {
		if jenis != "" && i < len(dataBarangList) {
			surat.BarangHilang = append(surat.BarangHilang, model.Barang{
				JenisBarang: jenis,
				Data:        dataBarangList[i],
			})
		}
	}

	if err := h.Repo.UpdateSurat(&surat); err != nil {
		http.Error(w, "Gagal mengupdate surat", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/surat?status=success_update", http.StatusSeeOther)
}

// SuratDelete menghapus surat
func (h *Handler) SuratDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.Repo.DeleteSurat(id); err != nil {
		http.Error(w, "Gagal menghapus surat", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/surat?status=success_delete", http.StatusSeeOther)
}

// SuratPrint menampilkan halaman siap cetak
func (h *Handler) SuratPrint(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	surat, err := h.Repo.GetSuratByID(id)
	if err != nil {
		http.Error(w, "Surat tidak ditemukan", http.StatusNotFound)
		return
	}
	pengaturan, err := h.Repo.GetPengaturan()
	if err != nil {
		http.Error(w, "Gagal mengambil pengaturan untuk cetak", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Surat":      surat,
		"Pengaturan": pengaturan,
	}

	h.renderPrint(w, "surat_print.html", data)
}

// PengaturanForm menampilkan halaman pengaturan
func (h *Handler) PengaturanForm(w http.ResponseWriter, r *http.Request) {
	pengaturan, err := h.Repo.GetPengaturan()
	if err != nil {
		http.Error(w, "Gagal mengambil data pengaturan", http.StatusInternalServerError)
		return
	}

	totalSurat, _ := h.Repo.GetTotalSurat()
	if totalSurat == 0 {
		pengaturan.LastNomorSurat = 0
	}

	pejabatList, _ := h.Repo.GetPetugasByTipe("Pejabat")
	penerimaList, _ := h.Repo.GetPetugasByTipe("Penerima")
	data := map[string]interface{}{
		"Pengaturan":   pengaturan,
		"PejabatList":  pejabatList,
		"PenerimaList": penerimaList,
		"Timestamp":    time.Now().Unix(),
	}
	h.render(w, "pengaturan.html", data)
}

// PengaturanUpdate menyimpan perubahan dari form pengaturan
func (h *Handler) PengaturanUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Gagal mem-parsing form", http.StatusBadRequest)
		return
	}

	p, err := h.Repo.GetPengaturan()
	if err != nil {
		http.Error(w, "Gagal mengambil pengaturan yang ada", http.StatusInternalServerError)
		return
	}

	pejabatID, _ := strconv.Atoi(r.FormValue("pejabat_id"))
	penerimaID, _ := strconv.Atoi(r.FormValue("penerima_id"))
	lastNomor, _ := strconv.Atoi(r.FormValue("last_nomor_surat"))

	p.KopSurat1 = r.FormValue("kop_surat_1")
	p.KopSurat2 = r.FormValue("kop_surat_2")
	p.KopSurat3 = r.FormValue("kop_surat_3")
	p.FormatNomorSurat = r.FormValue("format_nomor_surat")
	p.Wilayah = r.FormValue("wilayah")
	p.NamaKantor = r.FormValue("nama_kantor")
	p.LastNomorSurat = lastNomor
	p.PejabatID = pejabatID
	p.PenerimaID = penerimaID

	if p.LastNomorYear == 0 {
		p.LastNomorYear = time.Now().Year()
	}

	file, handler, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		newFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
		filePath := filepath.Join("web", "static", "uploads", newFileName)

		dst, createErr := os.Create(filePath)
		if createErr != nil {
			http.Error(w, "Gagal membuat file di server", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, copyErr := io.Copy(dst, file); copyErr != nil {
			http.Error(w, "Gagal menyalin file", http.StatusInternalServerError)
			return
		}
		p.LogoPath = "/static/uploads/" + newFileName
	} else if err != http.ErrMissingFile {
		http.Error(w, "Error saat upload file", http.StatusInternalServerError)
		return
	}

	if err := h.Repo.UpdatePengaturan(p); err != nil {
		http.Error(w, "Gagal menyimpan pengaturan", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/pengaturan?status=success_update", http.StatusSeeOther)
}
