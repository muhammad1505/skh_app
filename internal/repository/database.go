package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func ConnectDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./skh.db?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Ganti createTables dengan runMigrations
	if err = runMigrations(db); err != nil {
		return nil, fmt.Errorf("gagal menjalankan migrasi: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	// 1. Buat tabel untuk mencatat versi migrasi
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER NOT NULL PRIMARY KEY);`)
	if err != nil {
		return err
	}

	// 2. Dapatkan versi migrasi saat ini dari database
	var currentVersion int
	err = db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&currentVersion)
	if err != nil {
		return err
	}
	
	log.Printf("Versi database saat ini: %d", currentVersion)

	// 3. Baca semua file migrasi dari folder
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files) // Urutkan file berdasarkan nama

	// 4. Jalankan migrasi yang versinya lebih tinggi dari versi saat ini
	for _, file := range files {
		fileName := filepath.Base(file)
		parts := strings.Split(fileName, "_")
		if len(parts) == 0 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // Abaikan file dengan format nama yang salah
		}

		if version > currentVersion {
			log.Printf("Menjalankan migrasi: %s (versi %d)", fileName, version)

			// Baca isi file SQL
			content, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			// Jalankan SQL di dalam transaksi agar aman
			tx, err := db.Begin()
			if err != nil {
				return err
			}

			if _, err := tx.Exec(string(content)); err != nil {
				tx.Rollback()
				return fmt.Errorf("error di file %s: %w", fileName, err)
			}

			// Catat versi baru ke tabel schema_migrations
			if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
				tx.Rollback()
				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}
			
			log.Printf("Migrasi %s berhasil.", fileName)
		}
	}

	log.Println("Migrasi database selesai.")
	return nil
}
