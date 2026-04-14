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

## 🚀 Setup & Menjalankan Aplikasi

### 1. Jalankan Container PostgreSQL

Perintah berikut akan mengunduh image PostgreSQL dan menjalankannya di
background melalui Docker (port `5432`):

``` bash
make postgres
```

------------------------------------------------------------------------

### 2. Buat Database Kosong

Membuat database bernama `omnilibrary` di dalam container yang sedang
berjalan:

``` bash
make createdb
```

------------------------------------------------------------------------

### 3. Jalankan Migrasi Skema

Menjalankan file SQL untuk membuat struktur tabel secara otomatis:

``` bash
make migrateup
```

------------------------------------------------------------------------

### 4. Jalankan Aplikasi Golang

``` bash
go run cmd/api/main.go
```

Server akan berjalan dan siap menerima request di:

http://localhost:8080

------------------------------------------------------------------------

## 📡 API Endpoints Documentation

### 1. Fetch & Save Book Metadata

Endpoint ini digunakan untuk mencari metadata buku berdasarkan ISBN
melalui Google Books API, lalu otomatis menyimpannya ke database
PostgreSQL.

Fitur: - Mengecek database lokal terlebih dahulu (efisiensi & hemat
bandwidth) - Menghindari duplikasi data (DRY principle)

------------------------------------------------------------------------

#### 🔹 Request

-   **URL**: `/api/books/fetch`\
-   **Method**: `POST`\
-   **Content-Type**: `application/json`

**Request Body:**

``` json
{
  "isbn": "9786020633176"
}
```

------------------------------------------------------------------------

#### ✅ Success Response (200 OK)

``` json
{
  "id": "e6a218d6-7a8e-4b92-8c1a-2d9f8e7a6b5c",
  "isbn": "9786020633176",
  "title": "Clean Architecture",
  "authors": ["Robert C. Martin"],
  "published_date": "2017-09-10T00:00:00Z",
  "description": "Practical Software Architecture Solutions...",
  "page_count": 432,
  "cover_url": "http://books.google.com/books/content?id=...",
  "created_at": "2026-04-14T23:50:00.000Z",
  "updated_at": "2026-04-14T23:50:00.000Z"
}
```

------------------------------------------------------------------------

#### ❌ Error Responses

-   **400 Bad Request**\
    Jika ISBN kosong atau format JSON tidak valid.

-   **500 Internal Server Error**\
    Jika terjadi kegagalan koneksi ke database atau API eksternal.

------------------------------------------------------------------------

## ❤️ Notes

Built with ❤️ for a productive late-night coding session.


