package game

import (
	"fmt"
)

var directions = [8][2]int{
	{-1, -1}, {-1, 0}, {-1, 1},
	{0, -1}, {0, 1},
	{1, -1}, {1, 0}, {1, 1},
}

type Move struct {
	Row, Col int
}

func (s *State) LegalMoves(player Cell) []Move {
	seen := make(map[Move]bool)

	opp := opponent(player)

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if s.Board[row][col] != player {
				continue
			}

			for _, d := range directions {
				dr, dc := d[0], d[1]
				r, c := row+dr, col+dc
				if !(0 <= r && r < 8 && 0 <= c && c < 8) {
					continue
				}
				if s.Board[r][c] != opp {
					continue
				}

				for {
					r += dr
					c += dc

					if !(0 <= r && r < 8 && 0 <= c && c < 8) {
						break
					}
					if s.Board[r][c] == opp {
						continue
					}

					if s.Board[r][c] == Empty {
						m := Move{Row: r, Col: c}
						seen[m] = true
						break
					}

					if s.Board[r][c] == player {
						break
					}
				}
			}
		}
	}

	moves := make([]Move, 0, len(seen))
	for m := range seen {
		moves = append(moves, m)
	}

	return moves
}

func opponent(currentPlayer Cell) Cell {
	if currentPlayer == Empty {
		return Empty
	} else if currentPlayer == White {
		return Black
	}
	return White
}

func (s *State) isLegal(player Cell, m Move) bool {
	for _, v := range s.LegalMoves(player) {
		if v == m {
			return true
		}
	}
	return false
}

func (s *State) ApplyMove(m Move) error {

	if !s.isLegal(s.Current, m) {
		return fmt.Errorf("move %v isn't legal", m)
	}

	enemy := opponent(s.Current)

	allReverses := make([]Move, 0)

	for _, d := range directions {
		dr, dc := d[0], d[1]
		r, c := m.Row, m.Col
		reverse := make([]Move, 0)

		for {
			r, c = r+dr, c+dc
			if !(0 <= r && r < 8 && 0 <= c && c < 8) {
				reverse = make([]Move, 0)
				break
			}

			if s.Board[r][c] == enemy {
				reverse = append(reverse, Move{r, c})
				continue
			}

			if s.Board[r][c] == s.Current {
				break
			}

			if s.Board[r][c] == Empty {
				reverse = make([]Move, 0)
				break
			}
		}

		if len(reverse) > 0 {
			allReverses = append(allReverses, reverse...)
		}
	}

	s.Board[m.Row][m.Col] = s.Current

	for _, rev := range allReverses {
		s.Board[rev.Row][rev.Col] = s.Current
	}

	s.NextTurn()

	return nil
}

func (s *State) NextTurn() {
	opp := opponent(s.Current)
	if len(s.LegalMoves(opp)) != 0 {
		s.Current = opp
	}
}

func (s *State) IsGameOver() bool {
	return len(s.LegalMoves(White)) == 0 && len(s.LegalMoves(Black)) == 0
}
