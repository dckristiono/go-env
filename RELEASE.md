# Panduan Rilis

Proyek ini menggunakan GitHub Actions dan GoReleaser untuk proses rilis otomatis setelah tag dibuat secara manual.

## Cara Melakukan Rilis Baru

### 1. Membuat Tag Secara Manual

Untuk membuat tag versi baru:

```bash
# Update versi di version.go (opsional)
# Commit perubahan jika diperlukan
git commit -am "Bump version to x.y.z"

# Buat tag dengan pesan
git tag -a vx.y.z -m "Release vx.y.z" 

# Push tag ke remote
git push origin vx.y.z
```

Contoh:
```bash
git tag -a v0.1.0 -m "Initial release" 
git push origin v0.1.0
```

### 2. Proses Rilis Otomatis

Setelah tag di-push, GitHub Actions akan otomatis:
- Mendeteksi tag baru dengan format `v*`
- Menjalankan workflow `release.yml`
- Menggunakan GoReleaser untuk membuat GitHub Release
- Menghasilkan changelog berdasarkan commit sejak rilis terakhir

## Format Commit untuk Changelog yang Baik

GoReleaser menghasilkan changelog berdasarkan commit messages. Untuk changelog yang rapi, gunakan format commit conventional:

- `feat: tambah fitur X` - untuk fitur baru
- `fix: perbaiki bug Y` - untuk bug fixes
- `refactor: refaktor fungsi Z` - untuk refactoring
- `perf: optimalkan operasi A` - untuk peningkatan performa
- `docs: update dokumentasi B` - untuk perubahan dokumentasi
- `test: tambah test untuk C` - untuk penambahan test
- `chore: update dependency D` - untuk perubahan tooling/dependencies

Pesan dengan awalan `docs:`, `test:`, dan `chore:` tidak akan muncul di changelog default.

## Versioning

Proyek ini mengikuti [Semantic Versioning](https://semver.org/):

- **Patch version** (`1.0.x`): Untuk backward-compatible bug fixes
- **Minor version** (`1.x.0`): Untuk backward-compatible new features
- **Major version** (`x.0.0`): Untuk breaking changes

## Menggunakan Versi Tertentu

Pengguna dapat menginstall versi tertentu dengan:

```
go get github.com/dckristiono/go-env@v1.0.0
```

## Konfigurasi Rilis

Konfigurasi GoReleaser dapat ditemukan di file `.goreleaser.yml`. File ini mengatur:
- Skip build binary (karena ini adalah library)
- Format dan pengelompokan changelog
- Footer release dengan petunjuk instalasi