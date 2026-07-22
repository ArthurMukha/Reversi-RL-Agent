from board import board
from config import *

def check_chip(player, chip_coords):

    row, col = chip_coords
    # print(row, col, player)

    if player == WHITE:
        my = white_chip
        enemy = black_chip
    else:
        my = black_chip
        enemy = white_chip

    steps = []

    for dr, dc in DIRECTIONS:

        r = row + dr
        c = col + dc

        # сосед должен быть противником
        if not (0 <= r < rows and 0 <= c < cols):
            continue
        
        # print(f"check: {r, c, board[r][c]}")

        if board[r][c] != enemy:
            continue

        # идем дальше
        while True:

            r += dr
            c += dc

            if not (0 <= r < rows and 0 <= c < cols):
                break
            
            # print(f"check next: {r, c, board[r][c]}")

            if board[r][c] == enemy:
                continue

            if board[r][c] == empty:
                steps.append((r, c))
                # print("found step place")
                break

            if board[r][c] == my:
                break

    return steps

def get_steps(player):
    if player == WHITE:
        p_chip = white_chip
        o_chip = black_chip
    else:
        p_chip = black_chip
        o_chip = white_chip

    steps = set()
    for row in range(rows):
        for col in range(cols):
            if board[row][col] == p_chip:
                for step in check_chip(player, [row, col]):
                    steps.add(step)
    return steps


def count_score():
    wc, bc = 0, 0
    for line in board:
        for chip in line:
            if chip == empty:
                continue
            wc, bc = (wc+1, bc) if chip == white_chip else (wc, bc+1)
    
    return wc, bc