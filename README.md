# OmniLibrary API 📚

OmniLibrary adalah sebuah *backend service* Minimum Viable Product (MVP) untuk mengelola perpustakaan pribadi dan *reading tracker*. Sistem ini dilengkapi dengan fitur *automated metadata fetcher* yang mengambil data buku lengkap berdasarkan ISBN.

Dibangun dengan memegang teguh prinsip **Clean Architecture** dan **DRY (Don't Repeat Yourself)**.

## 🚀 Tech Stack

* **Language:** Golang (Go)
* **Framework:** Fiber (v2)
* **Database:** PostgreSQL
* **ORM:** GORM
* **Migrations:** `golang-migrate`
* **External API:** Google Books API

## 🏗️ Architecture

Proyek ini mengadopsi pola *Clean Architecture* dengan struktur direktori sebagai berikut:

* `cmd/api/` - *Entry point* aplikasi dan *wiring* (*Dependency Injection*).
* `internal/domain/` - *Entities* murni dan *Interfaces* (Kontrak kerja). Tidak ada *dependency* eksternal di sini.
* `internal/delivery/http/` - *Layer* presentasi (Fiber Handlers) untuk menerima dan merespons HTTP *request*.
* `internal/usecase/` - Otak aplikasi yang berisi *Business Logic*.
* `internal/repository/postgres/` - Implementasi *Data Access Object* (DAO) menggunakan GORM.
* `internal/repository/external/` - Implementasi *HTTP Client* untuk berinteraksi dengan API pihak ketiga.

## 🛠️ Prerequisites

Sebelum menjalankan aplikasi, pastikan sistem kamu sudah ter-install:
1. [Go](https://golang.org/doc/install) (v1.20+)
2. [Docker](https://docs.docker.com/get-docker/) & Docker Compose
3. [Make](https://www.gnu.org/software/make/)
4. [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

## ⚙️ Setup & Installation

Proyek ini menggunakan `Makefile` untuk mempermudah operasional *database* lokal.

1. **Clone repository ini:**
   ```bash
   git clone [https://github.com/faridlan/omni-library-api.git](https://github.com/faridlan/omni-library-api.git)
   cd omnilibrary