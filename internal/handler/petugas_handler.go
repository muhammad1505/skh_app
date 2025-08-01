package handler

import (
	"net/http"
	"skh_app/internal/model"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) PetugasList(w http.ResponseWriter, r *http.Request) {
	petugas, err := h.Repo.GetAllPetugas()
	if err != nil {
		http.Error(w, "Gagal mengambil data petugas", http.StatusInternalServerError)
		return
	}
	h.render(w, "petugas_list.html", petugas)
}

func (h *Handler) PetugasFormNew(w http.ResponseWriter, r *http.Request) {
	h.render(w, "petugas_form.html", nil)
}

func (h *Handler) PetugasCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}
	p := &model.Petugas{
		Nama:    r.FormValue("nama"),
		Pangkat: r.FormValue("pangkat"),
		NRP:     r.FormValue("nrp"),
		Jabatan: r.FormValue("jabatan"),
		Tipe:    r.FormValue("tipe"),
	}
	if err := h.Repo.CreatePetugas(p); err != nil {
		http.Error(w, "Gagal menyimpan data petugas", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/petugas?status=success_update", http.StatusSeeOther)
}

func (h *Handler) PetugasFormEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	p, err := h.Repo.GetPetugasByID(id)
	if err != nil {
		http.Error(w, "Data petugas tidak ditemukan", http.StatusNotFound)
		return
	}
	h.render(w, "petugas_form.html", p)
}

func (h *Handler) PetugasUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Gagal parsing form", http.StatusBadRequest)
		return
	}
	p := &model.Petugas{
		ID:      id,
		Nama:    r.FormValue("nama"),
		Pangkat: r.FormValue("pangkat"),
		NRP:     r.FormValue("nrp"),
		Jabatan: r.FormValue("jabatan"),
		Tipe:    r.FormValue("tipe"),
	}
	if err := h.Repo.UpdatePetugas(p); err != nil {
		http.Error(w, "Gagal mengupdate data petugas", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/petugas?status=success_update", http.StatusSeeOther)
}

func (h *Handler) PetugasDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.Repo.DeletePetugas(id); err != nil {
		http.Error(w, "Gagal menghapus data petugas", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/petugas?status=success_delete", http.StatusSeeOther)
}
