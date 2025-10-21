.PHONY: run test docker-up docker-down migrate clean stop start restart itest health swagger swagger-ui start-all k6-install k6-auth k6-cache k6-posts k6-posts-url k6-post-retrieval k6-journey k6-all

# Defaults (can be overridden)
PORT ?= 8080
SWAGGER_PORT ?= 8081
JWT_SECRET ?= super-secret-key-for-testing

run:
	go run cmd/api/main.go

test:
	go test -v -cover ./...

test-full:
	go test -race -v ./tests/...

swagger:
	$(shell go env GOPATH)/bin/swag init -g cmd/api/main.go -o docs

swagger-ui:
	@echo "Starting Swagger UI on :$(SWAGGER_PORT)"
	@SWAGGER_PORT=$(SWAGGER_PORT) go run cmd/swagger/main.go

start-all: stop docker-up
	@echo "Starting API on :$(PORT) and Swagger UI on :$(SWAGGER_PORT)"
	@JWT_SECRET='$(JWT_SECRET)' PORT=$(PORT) go run cmd/api/main.go & \
	SWAGGER_PORT=$(SWAGGER_PORT) go run cmd/swagger/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/001_create_users.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/002_create_posts.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/003_create_likes.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/004_create_comments.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/005_create_post_views.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/006_optimize_indexes.up.sql

clean:
	docker-compose down --volumes
	rm -f instagrano
	go clean

# Kill any server bound to $(PORT)
stop:
	@if lsof -ti:$(PORT) >/dev/null; then \
		echo "Stopping server on :$(PORT)..."; \
		lsof -ti:$(PORT) | xargs kill -9; \
	else \
		echo "No server listening on :$(PORT)"; \
	fi

# Start server with env vars, after ensuring port is free and services are running
start: stop docker-up
	@echo "Starting API server on :$(PORT) with JWT_SECRET set"
	@JWT_SECRET='$(JWT_SECRET)' PORT=$(PORT) go run cmd/api/main.go

# Convenience restart target
restart: stop start

# Quick health check
health:
	@curl -s http://localhost:$(PORT)/health || true

# Run integration tests (expects server already running)
itest:
	@./run_integration_tests.sh

# Redis utilities
.PHONY: redis-cli redis-flush

redis-cli:
	docker-compose exec redis redis-cli

redis-flush:
	docker-compose exec redis redis-cli FLUSHDB

# K6 Performance Testing
.PHONY: k6-install k6-auth k6-cache k6-posts k6-journey k6-all

k6-install:
	@echo "Installing K6..."
	@which k6 || (echo "K6 not found. Please install K6:" && echo "  macOS: brew install k6" && echo "  Linux: https://k6.io/docs/getting-started/installation/" && exit 1)

k6-auth:
	@echo "Running K6 authentication load test..."
	@k6 run --log-output=none tests/k6/scenarios/auth.js

k6-cache:
	@echo "Running K6 cache performance test..."
	@k6 run --log-output=none tests/k6/scenarios/feed-cache.js

k6-posts:
	@echo "Running K6 post creation load test..."
	@k6 run --log-output=none tests/k6/scenarios/posts.js

k6-posts-url:
	@echo "Running K6 post creation from URL test..."
	@k6 run --log-output=none tests/k6/scenarios/posts-url.js

k6-journey:
	@echo "Running K6 full user journey test..."
	@k6 run --log-output=none tests/k6/scenarios/user-journey.js

k6-post-retrieval:
	@echo "Running K6 post retrieval cache test..."
	@k6 run --log-output=none tests/k6/scenarios/post-retrieval.js

k6-all: k6-install
	@echo "Running all K6 performance tests..."
	@k6 run --log-output=none tests/k6/scenarios/auth.js
	@k6 run --log-output=none tests/k6/scenarios/feed-cache.js
	@k6 run --log-output=none tests/k6/scenarios/posts.js
	@k6 run --log-output=none tests/k6/scenarios/posts-url.js
	@k6 run --log-output=none tests/k6/scenarios/post-retrieval.js
	@k6 run --log-output=none tests/k6/scenarios/user-journey.js
