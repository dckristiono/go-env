# Cara Merilis Versi Baru

Proyek ini menggunakan GitHub Actions untuk membantu proses pembuatan tag dan rilis. Ada dua workflow yang relevan dalam proses ini:

## 1. Tag dan Release Otomatis

Untuk membuat rilis baru:

1. Buka repositori di GitHub
2. Pilih tab **Actions**
3. Pilih workflow "Bump version and create release"
4. Klik tombol **Run workflow**
5. Masukkan informasi yang diminta:
    - **Version**: Versi baru (misalnya: 1.0.1, 1.1.0, 2.0.0)
    - **Type**: Jenis rilis (patch, minor, atau major)
6. Klik **Run workflow**

Workflow ini akan:
- Mengupdate versi di `version.go`
- Membuat commit perubahan versi
- Membuat tag dengan format `v{version}` (misalnya: v1.0.1)
- Push tag dan commit ke repositori

## 2. Pembuatan Release

Setelah tag dibuat, workflow `go.yml` akan otomatis:
- Mendeteksi tag baru dengan format `v*`
- Menjalankan build dan test
- Membuat GitHub Release yang sesuai dengan tag
- Menghasilkan changelog secara otomatis dari commit sejak rilis terakhir

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

## Changelog

Changelog dibuat otomatis berdasarkan commit messages. Untuk memudahkan pembacaan changelog, disarankan untuk menggunakan format commit yang baik, misalnya:

- `fix: perbaikan bug X`
- `feat: tambah fitur Y`
- `docs: update dokumentasi Z`
- `refactor: refactor fungsi A`