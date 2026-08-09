import torch
import torch.nn.functional as F

from model_service.v1 import model_service_pb2 as pb
# from model_service.v1 import model_service_pb2_grpc as pb_grpc

def encode(board, current) -> torch.Tensor:
    if current != pb.CELL_WHITE and current != pb.CELL_BLACK:
        raise ValueError(f"invalid current: {current}")
    
    own = torch.zeros((8, 8), dtype=torch.float32)
    opp = torch.zeros((8, 8), dtype=torch.float32)

    for r in range(8):
        for c in range(8):
            if board[r * 8 + c] == current:
                own[r][c] = 1.0
            if board[r * 8 + c] == opponent(current):
                opp[r][c] = 1.0
    return torch.stack([own, opp]).unsqueeze(0)



def opponent(cell) -> pb.Cell:
    return pb.CELL_BLACK if cell == pb.CELL_WHITE else pb.CELL_WHITE

def action_mask(legal_moves) -> torch.Tensor:
    
    if len(legal_moves) == 0:
        raise ValueError("no one legal move")
    
    mask = torch.full((8 * 8,), float("-inf"), dtype=torch.float32)

    for m in legal_moves:
        mask[m.row * 8 + m.col] = 0.0
    
    return mask

def select_move(model, board, current, legal_moves, temperature) -> tuple[tuple[int, int], float, list[float]]:
    
    with torch.inference_mode():
        state = encode(board, current)
        mask = action_mask(legal_moves)

        
        policy, value = model(state)
        masked = policy + mask
        probs = torch.softmax(masked, dim=1)

        if temperature == 0.0:
            index = int(masked.argmax(dim=1)[0])
        else:
            sample = torch.softmax(masked / temperature, dim=1)   # для выбора
            index = int(torch.multinomial(sample, 1)[0])
        r, c = index // 8, index % 8
        return (r, c), value.item(), probs.squeeze(0).tolist()

    