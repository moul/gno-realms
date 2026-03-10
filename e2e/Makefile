export COMPOSE_PROJECT_NAME=e2e
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

.PHONY: up down test test-only clean logs build

# Start all services (gno, atomone, tx-indexer, relayer)
up:
	docker compose up -d --build --force-recreate --progress plain

# Stop everything and remove volumes
down:
	docker compose down -v --progress plain

# Full e2e: build images, start services, run Go tests, teardown
test:
	docker compose up -d --build --force-recreate --progress plain
	go test -v -timeout 10m -count=1 ./...; ret=$$?; docker compose down -v --progress plain; exit $$ret

# Run Go tests against already-running infrastructure
test-only:
	go test -v -timeout 10m -count=1 ./...

# Stop everything, remove volumes and locally-built images
clean:
	docker compose down -v --rmi local --progress plain

# Follow logs for all services
logs:
	docker compose logs -f

# Build all images without running
build:
	docker compose build --progress plain
