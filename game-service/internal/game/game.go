package game

// import (

// )

type Cell int8

const (
	Empty Cell = iota
	White
	Black
)

func (c Cell) String() string {
	switch c {
	case Empty:
		return "Empty"
	case White:
		return "White"
	case Black:
		return "Black"
	default:
		return "Unknown"
	}
}

type State struct {
	Board   [8][8]Cell
	Current Cell
}

func New() *State {
	s := &State{Current: White}

	s.Board[3][3] = White
	s.Board[4][4] = White
	s.Board[3][4] = Black
	s.Board[4][3] = Black

	return s
}

func (s *State) Score() (white, black int) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if s.Board[i][j] == White {
				white++
			} else if s.Board[i][j] == Black {
				black++
			}
		}
	}
	return white, black
}
