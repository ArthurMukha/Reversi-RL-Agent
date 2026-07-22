
import curses
from config import WHITE, BLACK, white_chip, black_chip
from input import make_move
from rules import get_steps
from renderer import draw_board, draw_score
from board import setup

def reverse_player(current_player):
    if current_player == WHITE:
        return BLACK
    return WHITE
    

def main(stdscr):
    stdscr.keypad(True)

    current_player = WHITE
    cursor_row = 3
    cursor_col = 3

    while True:

        if not get_steps(WHITE) and not get_steps(BLACK):
            setup()
            current_player = WHITE
            cursor_row = 3
            cursor_col = 3

        stdscr.clear()

        steps = get_steps(current_player)

        draw_board(stdscr, steps, cursor_row, cursor_col)
        draw_score(stdscr, current_player)

        stdscr.refresh()

        key = stdscr.getch()

        if key == ord("q"):
            break
        if key == ord("r"):
            setup()

        if key == curses.KEY_UP:
            cursor_row = max(0, cursor_row - 1)

        elif key == curses.KEY_DOWN:
            cursor_row = min(7, cursor_row + 1)

        elif key == curses.KEY_LEFT:
            cursor_col = max(0, cursor_col - 1)

        elif key == curses.KEY_RIGHT:
            cursor_col = min(7, cursor_col + 1)
        
        if key in (10, 13, curses.KEY_ENTER):
            if (cursor_row, cursor_col) in steps:
                
                make_move(current_player, cursor_row, cursor_col)


                current_player = reverse_player(current_player)
        

        if not get_steps(current_player):
            current_player = reverse_player(current_player)


curses.wrapper(main)

"""
включает режим curses
выключает эхо клавиатуры
включает специальный режим терминала
при ошибке восстанавливает терминал

Без wrapper после падения программы терминал может стать "сломанным".
"""