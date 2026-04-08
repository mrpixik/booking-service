up_local: # запуск бд и миграций через докер. Само приложение запустится через go run
	docker-compose -f docker-compose.local.yaml up -d
	go run cmd/main.go

down_local: # остановка контейнеров. чтобы завершить работу сервиса необходимо еще отправить ctr+c в консоль
	docker compose -f docker-compose.local.yaml stop

up:
	docker-compose up --build -d

seed:
	docker exec -i booking-service-db psql -U pixik -d booking-db < seeds/seed.sql

unit-test:
	go test ./... -v

integration-test:
	docker-compose -f docker-compose.test.yaml up --build -d
	go test ./internal/integration_test/ -v

swagger:
	swag init -g cmd/main.go -o docs

lint:
	golangci-lint run ./...

# Деплой (для себя)
deploy_local: #для личного удобства
	docker build -t booking-service:latest -f ./deploy/docker/Dockerfile .

deploy_push_remote: #для личного удобства
	docker login
	docker tag booking-service:latest pixik/booking-service:latest
	docker push pixik/booking-service:latest
