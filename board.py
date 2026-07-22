from config import *

board = [
    [empty] * cols for _ in range(rows)
]

def setup():
    
    for i in range(len(board)):
        for j in range(len(board[i])):
            board[i][j] = empty

    board[3][3] = white_chip
    board[4][4] = white_chip
    board[3][4] = black_chip
    board[4][3] = black_chip