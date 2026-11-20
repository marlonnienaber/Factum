-include .env
export

DATABASE_URL?=postgres://username:password@localhost:5435/factum_database?sslmode=disable
MIGRATIONS_FOLDER_PATH=./migrations

run:
	@go build -o .server-binary .
	@exec ./.server-binary

containers-up:
	@docker compose up -d
	@until [ "$$(docker inspect factum_database --format='{{.State.Health.Status}}')" = "healthy" ]; do sleep 0.1; done

containers-down:
	@docker compose down

containers-wipe-up:
	@docker compose down -v
	@docker compose up -d

containers-up-and-run:
	$(MAKE) containers-up
	$(MAKE) run

create-migration:
	@migrate create -ext sql -dir $(MIGRATIONS_FOLDER_PATH) $(name)

migrate-up:
	@if [ -z "$(n)" ]; then \
		migrate -path $(MIGRATIONS_FOLDER_PATH) -database "$(DATABASE_URL)" up 1; \
	elif [ "$(n)" = "all" ]; then \
		migrate -path $(MIGRATIONS_FOLDER_PATH) -database "$(DATABASE_URL)" up; \
	else \
		migrate -path $(MIGRATIONS_FOLDER_PATH) -database "$(DATABASE_URL)" up $(n); \
	fi

migrate-down:
	@if [ -z "$(n)" ]; then \
		migrate -path $(MIGRATIONS_FOLDER_PATH) -database "$(DATABASE_URL)" down 1; \
	else \
		migrate -path $(MIGRATIONS_FOLDER_PATH) -database "$(DATABASE_URL)" down $(n); \
	fi

current-migration:
	@migrate -path $(MIGRATIONS_FOLDER_PATH) -database "$(DATABASE_URL)" version

test:
	@if [ -z "$(feature)" ]; then \
		echo "Running all integration tests with coverage..."; \
		go test ./internal/tests -count=1 \
			-cover \
			-coverpkg=./internal/... \
			-coverprofile=coverage.out; \
	else \
		echo "Running $(feature) integration tests with coverage..."; \
		feature_pkg="./internal/features/$(shell echo $(feature) | tr A-Z a-z)/..."; \
		pattern="(?i)^Test$(feature)[A-Z]"; \
		go test ./internal/tests -count=1 \
			-run "$$pattern" \
			-cover \
			-coverpkg=$$feature_pkg \
			-coverprofile=coverage.out; \
	fi

investigate-coverage:
	@go tool cover -html=coverage.out
