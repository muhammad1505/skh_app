-- Tabel surat keterangan hilang
CREATE TABLE IF NOT EXISTS surat (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nomor_surat TEXT NOT NULL UNIQUE,
    tanggal_surat DATETIME NOT NULL,
    pelapor_nama TEXT,
    pelapor_ttl TEXT,
    pelapor_agama TEXT,
    pelapor_kelamin TEXT,
    pelapor_pekerjaan TEXT,
    pelapor_alamat TEXT,
    lokasi_hilang TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tabel barang yang hilang
CREATE TABLE IF NOT EXISTS barang (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    surat_id INTEGER,
    deskripsi TEXT,
    FOREIGN KEY(surat_id) REFERENCES surat(id) ON DELETE CASCADE
);

-- Tabel pengaturan versi lama
CREATE TABLE IF NOT EXISTS pengaturan (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    kop_surat_1 TEXT,
    kop_surat_2 TEXT,
    kop_surat_3 TEXT,
    logo_path TEXT,
    format_nomor_surat TEXT,
    last_nomor_surat INTEGER DEFAULT 0,
    pejabat_nama TEXT, pejabat_pangkat TEXT, pejabat_nrp TEXT, pejabat_jabatan TEXT,
    penerima_nama TEXT, penerima_pangkat TEXT, penerima_nrp TEXT, penerima_jabatan TEXT
);

-- Insert data awal pengaturan
INSERT OR IGNORE INTO pengaturan (id) VALUES (1);
