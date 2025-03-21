## migration Up
migrate -path ./internal/infrastructure/database/sql_migrations/ -database "postgres://postgres:password@localhost:5432/petdoc?sslmode=disable" up
## cloudflare
cloudflared tunnel --url http://localhost:8080