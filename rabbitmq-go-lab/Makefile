include .env
export

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down -v

setup-topology:
	go run ./cmd/admin --setup

run-consumer:
	go run ./cmd/consumer

publish:
	go run ./cmd/publisher --count 10 --key demo.info

republish-dlq:
	go run ./cmd/admin --republish-dlq --limit 100
