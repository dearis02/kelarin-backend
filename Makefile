# Usage: make migration:new name=add_users_table
migration\:new:
	@migrate create -ext sql -dir ./db/migrations ${name}

# Usage: make migration:up database=postgres://user:password@127.0.0.1:5432/db_name?sslmode=disable
migration\:up:
	migration-tool up

# Usage: make migration:down database=postgres://user:password@127.0.0.1:5432/db_name?sslmode=disable
migration\:down:
	@migrate -database ${database} -path ./db/migrations down

# Usage: make wire:generate
wire\:generate:
	wire ./cmd/server