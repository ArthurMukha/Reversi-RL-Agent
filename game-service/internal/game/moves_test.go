package game

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func board(rows [8]string) [8][8]Cell {

	let2cell := map[byte]Cell{
		'.': Empty,
		'W': White,
		'B': Black,
	}

	var b [8][8]Cell
	for idx := 0; idx < 64; idx++ {
		b[idx/8][idx%8] = let2cell[rows[idx/8][idx%8]]
	}

	return b
}

func render(b [8][8]Cell) string {
	cell2let := map[Cell]byte{
		Empty: '.',
		White: 'W',
		Black: 'B',
	}

	var sb strings.Builder
	sb.Grow(8 * 9) // 8 строк по 8 символов плюс перевод строки

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			sb.WriteByte(cell2let[b[row][col]])
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

func TestApplyMove(t *testing.T) {

	tests := []struct {
		name        string
		board       [8][8]Cell
		current     Cell
		move        Move
		wantBoard   [8][8]Cell
		wantCurrent Cell
		wantErr     bool
	}{
		{
			name:    "пустая клетка вдали от фишек",
			board:   New().Board,
			current: White,
			move: Move{
				Row: 0,
				Col: 0,
			},
			wantBoard:   New().Board,
			wantCurrent: White,
			wantErr:     true,
		},
		{
			name:    "клетка занята",
			board:   New().Board,
			current: White,
			move: Move{
				Row: 3,
				Col: 3,
			},
			wantBoard:   New().Board,
			wantCurrent: White,
			wantErr:     true,
		},
		{
			name:    "легален для Black, но не для White",
			board:   New().Board,
			current: White,
			move: Move{
				Row: 2,
				Col: 3,
			},
			wantBoard:   New().Board,
			wantCurrent: White,
			wantErr:     true,
		},
		{
			name: "поворот в трех направлениях",
			board: board([8]string{
				"........",
				"........",
				"........",
				"....BW..",
				"...BB...",
				"...W.W..",
				"........",
				"BWW.....",
			}),
			current: White,
			move: Move{
				Row: 3,
				Col: 3,
			},
			wantBoard: board([8]string{
				"........",
				"........",
				"........",
				"...WWW..",
				"...WW...",
				"...W.W..",
				"........",
				"BWW.....",
			}),
			wantCurrent: Black,
			wantErr:     false,
		},
		{
			name: "пропуск хода",
			board: board([8]string{
				"WBB.....",
				"........",
				"........",
				"........",
				"........",
				"........",
				"........",
				"WBB.....",
			}),
			current: White,
			move: Move{
				Row: 0,
				Col: 3,
			},
			wantBoard: board([8]string{
				"WWWW....",
				"........",
				"........",
				"........",
				"........",
				"........",
				"........",
				"WBB.....",
			}),
			wantCurrent: White,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				s := State{
					Board:   tt.board,
					Current: tt.current,
				}
				err := s.ApplyMove(tt.move)

				if (err != nil) != tt.wantErr {
					t.Errorf("ApplyMove(%v) = nil, want error", tt.move)
				}

				if tt.wantCurrent != s.Current {
					t.Errorf("ApplyMove(%v): s.Current = %s, want %s", tt.move, s.Current, tt.wantCurrent)
				}

				if tt.wantBoard != s.Board {
					t.Errorf("ApplyMove(%v): доска совпала:\nОжидали:\n%s\nПолучили:\n%s", tt.move, render(tt.wantBoard), render(s.Board))
				}
			})
	}
}

func assertMoves(t *testing.T, what string, got, want []Move) {
	t.Helper()

	got, want = slices.Clone(got), slices.Clone(want)

	sortF := func(a, b Move) int {
		if c := cmp.Compare(a.Row, b.Row); c != 0 {
			return c
		}
		return cmp.Compare(a.Col, b.Col)
	}

	slices.SortFunc(got, sortF)

	slices.SortFunc(want, sortF)

	if !slices.Equal(got, want) {
		t.Errorf("%s:\ngot = %v\nwant %v", what, got, want)
	}
}

func TestLegalMoves(t *testing.T) {

	tests := []struct {
		name     string
		state    State
		movesFor Cell
		want     []Move
	}{
		{
			name: "проверка легальных ходов для White",
			state: State{
				Board: board(
					[8]string{
						"........",
						"........",
						"....W...",
						"...WW...",
						"...BW...",
						"........",
						"........",
						"........",
					},
				),
				Current: Black,
			},
			movesFor: White,
			want:     []Move{{4, 2}, {5, 2}, {5, 3}},
		},
		{
			name: "проверка легальных ходов для Black",
			state: State{
				Board: board(
					[8]string{
						"........",
						"........",
						"....W...",
						"...WW...",
						"...BW...",
						"........",
						"........",
						"........",
					},
				),
				Current: Black,
			},
			movesFor: Black,
			want:     []Move{{2, 3}, {2, 5}, {4, 5}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			moves := tt.state.LegalMoves(tt.movesFor)

			assertMoves(t, fmt.Sprintf("LegalMoves(%s)", tt.movesFor), moves, tt.want)

		})
	}

}

func TestNextTurn(t *testing.T) {
	tests := []struct {
		name        string
		state       State
		wantCurrent Cell
	}{
		{
			name:        "у черных есть ходы",
			state:       *New(),
			wantCurrent: Black,
		},
		{
			name: "доска-якорь: у черных ходов нет",
			state: State{
				Board: board([8]string{
					"WBB.....",
					"........",
					"........",
					"........",
					"........",
					"........",
					"........",
					"WBB.....",
				}),
				Current: White,
			},
			wantCurrent: White,
		},
		{
			name: "только белые фишки: ни у кого нет ходов",
			state: State{
				Board: board([8]string{
					"WWWWWWWW",
					"WWWWWWWW",
					"WWWWWWWW",
					"WWWWWWWW",
					"WWWWWWWW",
					"WWWWWWWW",
					"WWWWWWWW",
					"WWWWWWWW",
				}),
				Current: White,
			},
			wantCurrent: White,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.NextTurn()

			if tt.state.Current != tt.wantCurrent {
				t.Errorf("NextTurn(): got current = %s, want %s", tt.state.Current, tt.wantCurrent)
			}
		})
	}
}

func TestIsGameOver(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{
			name:  "начало игры",
			state: *New(),
			want:  false,
		},
		{
			name: "доска в шахматную раскраску, все 64 фишки",
			state: State{
				Board: board([8]string{
					"WBWBWBWB",
					"BWBWBWBW",
					"WBWBWBWB",
					"BWBWBWBW",
					"WBWBWBWB",
					"BWBWBWBW",
					"WBWBWBWB",
					"BWBWBWBW",
				}),
				Current: White,
			},
			want: true,
		},
		{
			name: "только белые фишки",
			state: State{
				Board: board([8]string{
					"........",
					"........",
					"........",
					"...WW...",
					"...WW...",
					"........",
					"........",
					"........",
				}),
				Current: White,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.IsGameOver()
			if got != tt.want {
				t.Errorf("IsGameOver(): got=%v want %v", got, tt.want)
			}
		})
	}
}
