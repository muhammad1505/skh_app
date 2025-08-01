package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"skh_app/internal/repository"
	"skh_app/web"
	"strings"
	"time"
)

// Handler struct menampung semua dependensi, seperti repository
type Handler struct {
	Repo      *repository.SuratRepository
	Templates map[string]*template.Template
}

// NewHandler membuat instance Handler baru
func NewHandler(repo *repository.SuratRepository) *Handler {
	h := &Handler{
		Repo: repo,
	}
	h.loadTemplates()
	return h
}

// loadTemplates memuat semua file HTML dari embed.FS
func (h *Handler) loadTemplates() {
	if h.Templates == nil {
		h.Templates = make(map[string]*template.Template)
	}

	// Daftarkan semua fungsi kustom di sini
	funcMap := template.FuncMap{
		"split":   strings.Split,
		"ToUpper": strings.ToUpper,
		"FormatTanggalIndo": func(t time.Time) string {
			bulan := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
			loc, _ := time.LoadLocation("Asia/Makassar")
			t = t.In(loc)
			return fmt.Sprintf("%d %s %d", t.Day(), bulan[t.Month()], t.Year())
		},
		"UnmarshalJson": func(jsonString string) (map[string]interface{}, error) {
			var result map[string]interface{}
			err := json.Unmarshal([]byte(jsonString), &result)
			return result, err
		},
		// FUNGSI BARU UNTUK KONVERSI KE JSON
		"ToJson": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	}

	// Muat template print terlebih dahulu dengan FuncMap
	printTmpl, err := template.New("surat_print.html").Funcs(funcMap).ParseFS(web.Files, "templates/surat_print.html")
	if err != nil {
		log.Fatalf("Gagal memuat template print: %v", err)
	}
	h.Templates["surat_print.html"] = printTmpl

	// Muat layout utama dengan FuncMap
	layout, err := template.New("layout.html").Funcs(funcMap).ParseFS(web.Files, "templates/layout.html")
	if err != nil {
		log.Fatalf("Gagal memuat layout: %v", err)
	}

	pages, err := web.Files.ReadDir("templates")
	if err != nil {
		log.Fatalf("Gagal membaca folder templates: %v", err)
	}

	for _, page := range pages {
		name := page.Name()
		if name == "layout.html" || name == "surat_print.html" {
			continue
		}

		clone, _ := layout.Clone()
		pt, err := clone.ParseFS(web.Files, "templates/"+name)
		if err != nil {
			log.Fatalf("Gagal memuat halaman %s: %v", name, err)
		}
		h.Templates[name] = pt
	}
}

// render adalah helper untuk merender template
func (h *Handler) render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := h.Templates[name]
	if !ok {
		http.Error(w, "Template tidak ditemukan: "+name, http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Error executing template %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// renderPrint adalah helper khusus untuk halaman print
func (h *Handler) renderPrint(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := h.Templates[name]
	if !ok {
		http.Error(w, "Template tidak ditemukan: "+name, http.StatusInternalServerError)
		return
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing print template %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
