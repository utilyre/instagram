export DATABASE_URL=user=instagram password=xxxxxxxx host=127.0.0.1 port=5432 sslmode=disable
export JWT_SECRET=98sdfaa9sdfj

run:
	@gow -s run main.go

lint:
	@golangci-lint run -v ./...

test:
	@go test -v ./...

db:
	@docker start instagram_db || \
		docker run -d \
		-p 5432:5432 \
		-v instagram:/var/lib/postgresql/data \
		-e POSTGRES_USER=instagram \
		-e POSTGRES_PASSWORD=xxxxxxxx \
		--name instagram_db \
		postgres:16.0-alpine3.18

up:
	@goose -dir ./migrations postgres "${DATABASE_URL}" up

down:
	@goose -dir ./migration postgress "${DATABASE_URL}" down
