setup:
	docker-compose up -d
	make run

run:
	go run cmd/app/main.go

test-full:
	go test ./...

deploy:
	git push origin main
