# Usage: make migration:new name=add_users_table
migration\:new:
	@migrate create -ext sql -dir ./database/migrations ${name}

# Usage: make migration:up
migration\:up:
	./migration-tool up

# Usage: make migration:down steps=1
migration\:down:
	./migration-tool down --steps ${steps}

# Usage: make migration:force version=20250113101930
migration\:force:
	./migration-tool force ${version}

# Usage: make wire:generate
wire\:generate:
	wire ./cmd/server