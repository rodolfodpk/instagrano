.PHONY: run test docker-up docker-down migrate clean

run:
	go run cmd/api/main.go

test:
	go test -v -cover ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/001_create_users.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/002_create_posts.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/003_create_likes.up.sql
	docker exec -i instagrano-postgres-1 psql -U postgres -d instagrano < migrations/004_create_comments.up.sql

clean:
	docker-compose down --volumes
	rm -f instagrano
	go clean
