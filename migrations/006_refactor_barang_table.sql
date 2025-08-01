-- Menghapus tabel barang yang lama
DROP TABLE IF EXISTS barang;

-- Membuat tabel barang baru dengan struktur yang lebih fleksibel
CREATE TABLE barang (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    surat_id INTEGER,
    jenis_barang TEXT NOT NULL, -- Untuk menyimpan tipe barang, misal: "KTP", "ATM"
    data TEXT NOT NULL,         -- Untuk menyimpan detail (NIK, No. Rek, dll) dalam format JSON
    FOREIGN KEY(surat_id) REFERENCES surat(id) ON DELETE CASCADE
);
