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

func TestLegalMovesInitial(t *testing.T) {
	g := New()

	whiteMoves := g.LegalMoves(White)

	whiteCorrectMoves := []Move{
		{3, 5}, {5, 3}, {2, 4}, {4, 2},
	}

	if len(whiteMoves) != len(whiteCorrectMoves) {
		t.Errorf(
			"White got %d moves. expected %d moves",
			len(whiteMoves),
			len(whiteCorrectMoves),
		)
	}

	whiteSet := map[Move]bool{}
	for _, m := range whiteMoves {
		whiteSet[m] = true
	}
	for _, w := range whiteCorrectMoves {
		if !whiteSet[w] {
			t.Errorf("Whate: expected %v move, but it isn`t in %v", w, whiteMoves)
		}
	}

	blackMoves := g.LegalMoves(Black)

	blackCorrectMoves := []Move{
		{3, 2}, {5, 4}, {4, 5}, {2, 3},
	}

	if len(blackMoves) != len(blackCorrectMoves) {
		t.Errorf(
			"Black got %d moves. expected %d moves",
			len(blackMoves),
			len(blackCorrectMoves),
		)
	}

	blackSet := map[Move]bool{}
	for _, m := range blackMoves {
		blackSet[m] = true
	}
	for _, w := range blackCorrectMoves {
		if !blackSet[w] {
			t.Errorf("Black: expected %v move, but it isn`t in %v", w, blackMoves)
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
