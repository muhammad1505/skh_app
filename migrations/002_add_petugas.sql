-- Tabel baru untuk manajemen petugas
CREATE TABLE IF NOT EXISTS petugas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nama TEXT NOT NULL,
    pangkat TEXT,
    nrp TEXT,
    jabatan TEXT,
    tipe TEXT NOT NULL -- 'Pejabat' atau 'Penerima'
);

-- Menambahkan kolom baru ke tabel pengaturan yang sudah ada
-- Perintah ini aman dan tidak akan menghapus data
ALTER TABLE pengaturan ADD COLUMN pejabat_id INTEGER;
ALTER TABLE pengaturan ADD COLUMN penerima_id INTEGER;
