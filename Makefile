# Build all services
build:
	docker-compose build

# Run all services
run:
	docker-compose up -d

# Stop all services
stop:
	docker-compose down

# Rebuild and restart all services
restart: stop build run

# Clean up
clean: stop
	docker-compose rm -f
	docker rmi teleport_server || true

.PHONY: build run stop restart clean
