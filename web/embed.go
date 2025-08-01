package web

import "embed"

// PERBAIKAN: Gunakan static/* untuk membungkus semua subfolder di dalamnya
//go:embed templates/* static/*
var Files embed.FS
