# ==============================================================================
# Environment Variables (Bisa disesuaikan nanti)
# ==============================================================================
DB_USER=root
DB_PASSWORD=secret
DB_HOST=localhost
DB_PORT=5432
DB_NAME=omnilibrary
DB_URL="postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

# ==============================================================================
# Database Commands
# ==============================================================================

# 1. Menjalankan container PostgreSQL di background
postgres:
	docker run --name omnilibrary-db -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASSWORD) -p $(DB_PORT):5432 -d postgres:15-alpine

# 2. Membuat database baru di dalam container
createdb:
	docker exec -it psqldb createdb --username=$(DB_USER) --owner=$(DB_USER) $(DB_NAME)

# 3. Menghapus database (Hati-hati!)
dropdb:
	docker exec -it psqldb dropdb $(DB_NAME)

# ==============================================================================
# Migration Commands
# ==============================================================================

# 4. Menjalankan migrasi UP (membuat tabel)
migrateup:
	migrate -path db/migrations -database $(DB_URL) -verbose up

# 5. Menjalankan migrasi DOWN (menghapus tabel)
migratedown:
	migrate -path db/migrations -database $(DB_URL) -verbose down

# ==============================================================================
# .PHONY memastikan make tidak bentrok dengan nama folder/file yang kebetulan sama
# ==============================================================================
.PHONY: postgres createdb dropdb migrateup migratedown