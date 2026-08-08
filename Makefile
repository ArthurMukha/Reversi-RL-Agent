.PHONY: proto proto-go proto-py proto-clear

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