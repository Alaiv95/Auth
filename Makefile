migration-up:
	go run ./cmd/migrator --storage-dsn=postgres://postgres:postgres@localhost:5434/auth?sslmode=disable --migrations-path="./migrations"


