from model_service.v1 import model_service_pb2 as pb
from model_service.v1 import model_service_pb2_grpc as pb_grpc

from random import choice
from concurrent import futures
import grpc


MODELS = {
    "random"
}

DEFAULT_TEMPERATURE = 0.2

class ModelServicer(pb_grpc.ModelServiceServicer):
    def __init__(self, server_id:str):
        self.server_id = server_id
    
    def SelectMove(self, request, context):
        
        if len(request.state.board) != 64:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"invalid board len: {len(request.state.board)}")
        if len(request.state.legal_moves) == 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "empty legal_moves")
        for lm in request.state.legal_moves:
            if not (0 <= lm.row < 8 and 0 <= lm.col < 8):
                context.abort(grpc.StatusCode.INVALID_ARGUMENT, "illegal move in legal_moves")
        if request.model_id not in MODELS:
            context.abort(grpc.StatusCode.NOT_FOUND, "model didnt found")
        
        temperature = request.temperature if request.HasField("temperature") else DEFAULT_TEMPERATURE

        if temperature < 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"invalid temperature: {temperature}")
        
        response = pb.SelectMoveResponse()
        
        move = choice(request.state.legal_moves)
        response.move.row = move.row
        response.move.col = move.col
        response.value = 0.0
        policy = [0.0] * 64

        for m in request.state.legal_moves:
            policy[m.row * 8 + m.col] = 1.0 / len(request.state.legal_moves)
        response.policy.extend(policy)
        response.model_id = request.model_id

        return response 





    
def serve(address: str = "127.0.0.1:50051") -> None:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    pb_grpc.add_ModelServiceServicer_to_server(ModelServicer("random_selection"), server)
    port = server.add_insecure_port(address)
    if port == 0:
        raise RuntimeError(f"не удалось занять {address}")
    server.start()
    print(f"selectmove-server слушает {address}")
    server.wait_for_termination()


if __name__ == "__main__":
    serve()