from board import board
from config import *
from rules import count_score

def screen_x(col):
    return BOARD_X + col * 4

def screen_y(row):
    return BOARD_Y + 2 + row

def draw_score(stdscr, current_player):
    white_count, black_count = count_score()

    score_y = screen_y(rows) + 1
    score_x = screen_x(0)

    stdscr.addstr(score_y, score_x, f"Белых фишек: {white_count}")
    stdscr.addstr(score_y+1, score_x, f"Белых фишек: {black_count}")
    stdscr.addstr(score_y+2, score_x, f"Ход {white_chip if current_player == WHITE else black_chip}")



def draw_board(stdscr, steps, cursor_row, cursor_col):
    for row in range(8):
        for col in range(8):

            x = screen_x(col)
            y = screen_y(row)

            if row == cursor_row and col == cursor_col:
                br_op = '▶'
                br_cl = '◀ '
            else:
                br_op = '['
                br_cl = '] '

            if (row, col) in steps:
                stdscr.addstr(y, x, br_op + '+' + br_cl)
            else:
                stdscr.addstr(y, x, br_op + board[row][col] + br_cl)