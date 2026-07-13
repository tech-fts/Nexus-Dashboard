.PHONY: infra-up infra-down api ingestor build-api build-ingestor frontend-dev

# Infrastructure
infra-up:
	docker compose up -d

infra-down:
	docker compose down

# Backend
build-api:
	cd backend && go build -o bin/api ./cmd/api

build-ingestor:
	cd backend && go build -o bin/ingestor ./cmd/ingestor

api: build-api
	./backend/bin/api

ingestor: build-ingestor
	./backend/bin/ingestor

# Frontend
frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# All-in-one
run: infra-up build-api build-ingestor
	@echo "Start API and ingestor in separate terminals or use docker-compose profiles"
