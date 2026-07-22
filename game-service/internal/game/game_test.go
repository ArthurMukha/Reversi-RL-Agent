package game

import "testing"

func TestNew(t *testing.T) {
	g := New()

	cases := []struct {
		row, col int
		want     Cell
	}{
		{3, 3, White},
		{4, 4, White},
		{3, 4, Black},
		{4, 3, Black},
	}

	for _, c := range cases {
		got := g.Board[c.row][c.col]
		if got != c.want {
			t.Errorf("state[%d][%d] = %d. Ожидали %d", c.row, c.col, got, c.want)
		}
	}
}

func TestNewCurrentPlayer(t *testing.T) {
	g := New()

	if g.Current != White {
		t.Errorf("current player is %d. expected %d", g.Current, White)
	}
}

func TestScoreInitial(t *testing.T) {
	g := New()

	w, b := g.Score()

	if w != 2 || b != 2 {
		t.Errorf("current score: w=%d b=%d. expected w=%d b=%d", w, b, 2, 2)
	}
}

func TestValidMovesInitial(t *testing.T) {
	g := New()

	w_moves := g.ValidMoves(White)

	w_correct_moves := []Move{
		{3, 5}, {5, 3}, {2, 4}, {4, 2},
	}

	if len(w_moves) != len(w_correct_moves) {
		t.Errorf(
			"White got %d moves. expected %d moves",
			len(w_moves),
			len(w_correct_moves),
		)
	}

	w_set := map[Move]bool{}
	for _, m := range w_moves {
		w_set[m] = true
	}
	for _, w := range w_correct_moves {
		if !w_set[w] {
			t.Errorf("Whate: expected %v move, but it isn`t in %v", w, w_moves)
		}
	}

	b_moves := g.ValidMoves(Black)

	b_correct_moves := []Move{
		{3, 2}, {5, 4}, {4, 5}, {2, 3},
	}

	if len(b_moves) != len(b_correct_moves) {
		t.Errorf(
			"Black got %d moves. expected %d moves",
			len(b_moves),
			len(b_correct_moves),
		)
	}

	b_set := map[Move]bool{}
	for _, m := range b_moves {
		b_set[m] = true
	}
	for _, w := range b_correct_moves {
		if !b_set[w] {
			t.Errorf("Black: expected %v move, but it isn`t in %v", w, b_moves)
		}
	}
}

func TestApplyMoveFlips(t *testing.T) {
	g := New()

	g.ApplyMove(Move{2, 4})

	if g.Board[3][4] != White {
		t.Errorf("s[3][4] = %d. must be %d", g.Board[3][4], White)
	}
	if g.Board[2][4] != White {
		t.Errorf("s[2][4] = %d. must be %d", g.Board[2][4], White)
	}

	w, b := g.Score()
	if w != 4 || b != 1 {
		t.Errorf("Score: w=%d b=%d. Must be: w=4 b=1", w, b)
	}

	if g.Current != Black {
		t.Errorf("Current player: %d. Must be %d", g.Current, Black)
	}
}
