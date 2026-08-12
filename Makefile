.PHONY: proto proto-go proto-py proto-clear check up dev down
.DEFAULT_GOAL := check
RUFF := model_service/.venv/bin/ruff

proto: proto-clear proto-go proto-py

proto-go:
	mkdir -p game-service/internal/pb
	protoc -I proto \
	--go_out=game-service --go_opt=module=github.com/ArthurMukha/reversi-rl-agent/game-service \
	--go-grpc_out=game-service --go-grpc_opt=module=github.com/ArthurMukha/reversi-rl-agent/game-service \
	proto/model_service/v1/model_service.proto

proto-py:
	model_service/.venv/bin/python -m grpc_tools.protoc -I proto \
	--python_out=. --pyi_out=. --grpc_python_out=. \
	proto/model_service/v1/model_service.proto


proto-clear:
	rm -rf game-service/internal/pb model_service/v1


check:
	@command -v $(RUFF) >/dev/null || { echo "нет $(RUFF) — поставь: model_service/.venv/bin/pip install -r model_service/requirements-dev.txt"; exit 1; }
	$(RUFF) check .
	$(RUFF) format --check .
	go -C game-service build ./...
	go -C game-service vet ./...
	test -z "$$(gofmt -l game-service/)"
	go -C game-service test -race ./...

up:
	docker compose up --build -d
	@echo "открой http://$$(docker compose port game-service 8080)"

dev: check up

down:
	docker compose down
