from config import *
from board import board

def reverse_chips(chips, curr_chip):
    for x, y in chips:
        board[x][y] = curr_chip
    

def make_move(current_player, cursor_row, cursor_col):
    if current_player == WHITE:
        curr_chip = white_chip
        enemy = black_chip
    else:
        curr_chip = black_chip
        enemy = white_chip
    
    all_reverses = []
    available = False

    for dr, dc in DIRECTIONS:

        r = cursor_row
        c = cursor_col

        reverse = []
        
        while True:

            r += dr
            c += dc

            if not (0 <= r < rows and 0 <= c < cols):
                reverse = []
                break

            if board[r][c] == enemy:
                reverse.append((r, c))
                continue

            if board[r][c] == curr_chip:
                break

            if board[r][c] == empty:
                reverse = []
                break

        if reverse:
            available = True
            all_reverses.extend(reverse)
    

    if available:
        board[cursor_row][cursor_col] = curr_chip
        reverse_chips(all_reverses, curr_chip)
        

    return available