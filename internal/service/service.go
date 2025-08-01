package service

import (
	"fmt"
	"strings"
	"time"
)

// GenerateNomorSurat membuat nomor surat lengkap berdasarkan format
func GenerateNomorSurat(format string, nomor int, t time.Time) string {
	tahun := fmt.Sprintf("%d", t.Year())
	bulanRomawi := toRoman(int(t.Month()))

	r := strings.NewReplacer(
		"{NO}", fmt.Sprintf("%03d", nomor), // Diberi padding 0, misal: 001, 002
		"{THN}", tahun,
		"{BLN_ROMAWI}", bulanRomawi,
		// Anda bisa menambahkan placeholder lain seperti {KODE} jika diperlukan
	)

	return r.Replace(format)
}

// toRoman mengubah angka bulan menjadi romawi
func toRoman(num int) string {
	romans := map[int]string{
		1: "I", 2: "II", 3: "III", 4: "IV", 5: "V", 6: "VI",
		7: "VII", 8: "VIII", 9: "IX", 10: "X", 11: "XI", 12: "XII",
	}
	return romans[num]
}
