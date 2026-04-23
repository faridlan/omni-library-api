# ==========================================
# STAGE 1: BUILDER (Pabrik Perakitan)
# ==========================================
# Gunakan image Golang versi alpine agar ringan untuk build
FROM golang:1.25-alpine AS builder

# Set working directory di dalam container
WORKDIR /app

# Copy go.mod dan go.sum lebih dulu
# Trik ini agar Docker me-nyimpan cache download library, 
# sehingga build berikutnya jauh lebih cepat.
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build aplikasinya!
# CGO_ENABLED=0 membuat binary yang dihasilkan benar-benar mandiri (statically linked)
# -o omni-api adalah nama output file binary-nya
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o omni-api ./cmd/api/main.go

# ==========================================
# STAGE 2: RUNNER (Etalase / Mesin Eksekusi)
# ==========================================
# Gunakan Alpine Linux super kecil (sekitar 5MB)
FROM alpine:latest

WORKDIR /app

# Install tzdata agar aplikasi Golang membaca zona waktu dengan benar
RUN apk --no-cache add tzdata

# Copy HANYA file binary dari STAGE 1
COPY --from=builder /app/omni-api .

# (Opsional) Copy file .env jika ada konfigurasi default 
# Biasanya saat produksi, variabel di-inject via docker-compose/server
COPY .env .

# Beritahu Docker port berapa yang digunakan aplikasimu (Misal: 8080)
EXPOSE 8080

# Perintah untuk menjalankan aplikasi saat container menyala
CMD ["./omni-api"]