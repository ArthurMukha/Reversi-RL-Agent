from model_service.v1 import model_service_pb2 as pb
from model_service.v1 import model_service_pb2_grpc as pb_grpc

from model_service.checkpoint import load_checkpoint, LoadedModel
from model_service.checkpoint import CHECKPOINTS_DIR
from model_service import inference

from concurrent import futures
from pathlib import Path
import logging
import grpc

log = logging.getLogger(__name__)

DEFAULT_TEMPERATURE = 0.2

class ModelServicer(pb_grpc.ModelServiceServicer):
    def __init__(self, server_id:str, checkpoints_path: Path):
        self.server_id = server_id
        self.checkpoints_path = checkpoints_path
        self.load_models()

    def load_models(self) -> None:
        self.models = {}
        for path in sorted(self.checkpoints_path.glob("*.pt")):
            loaded = load_checkpoint(path)
            if loaded.model_id in self.models:
                raise RuntimeError(f"models with same name in folder: {loaded.model_id}")
            self.models[loaded.model_id] = loaded

        if not self.models:
            raise RuntimeError(f"no one model in folder: {self.checkpoints_path}")
        

    def SelectMove(self, request, context):
        
        if len(request.state.board) != 64:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"invalid board len: {len(request.state.board)}")
        if len(request.state.legal_moves) == 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, "empty legal_moves")
        for lm in request.state.legal_moves:
            if not (0 <= lm.row < 8 and 0 <= lm.col < 8):
                context.abort(grpc.StatusCode.INVALID_ARGUMENT, "illegal move in legal_moves")
        if request.model_id not in self.models:
            context.abort(grpc.StatusCode.NOT_FOUND, "model didnt found")
        
        temperature = request.temperature if request.HasField("temperature") else DEFAULT_TEMPERATURE

        if temperature < 0:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"invalid temperature: {temperature}")
        
        response = pb.SelectMoveResponse()
        
        try:
            (row, col), value, policy = inference.select_move(
                self.models[request.model_id].model,
                request.state.board,
                request.state.current,
                request.state.legal_moves,
                temperature,
            )
        except ValueError as err:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(err))

        response.move.row = row
        response.move.col = col
        response.value = value

        response.policy.extend(policy)
        response.model_id = request.model_id

        return response 

    
def serve(address: str = "127.0.0.1:50051") -> None:

    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    pb_grpc.add_ModelServiceServicer_to_server(
        ModelServicer(
            "iter13-wr72",
            CHECKPOINTS_DIR
            ), 
            server,
        )
    port = server.add_insecure_port(address)
    if port == 0:
        raise RuntimeError(f"не удалось занять {address}")
    server.start()
    log.info("selectmove-server слушает %s", address)
    server.wait_for_termination()

if __name__ == "__main__":
    serve()