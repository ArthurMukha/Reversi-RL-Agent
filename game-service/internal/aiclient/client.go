package aiclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/game"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	ErrUnavailable = errors.New("aiclient: model service unavailable")
	ErrBadResponse = errors.New("aiclient: bad response from model")
)

type Client struct {
	conn    *grpc.ClientConn
	stub    pb.ModelServiceClient
	timeout time.Duration
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, fmt.Errorf("aiclient: %w", err)
	}

	return &Client{
		conn:    conn,
		stub:    pb.NewModelServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func toPBState(st *game.State, legalMoves []game.Move) *pb.State {
	board := st.Board
	current := st.Current

	flattenBoard := make([]pb.Cell, 64)
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			flattenBoard[row*8+col] = toPBCell(board[row][col])
		}
	}

	pbLegalMoves := make([]*pb.Move, 0, len(legalMoves))

	for _, m := range legalMoves {
		pbLegalMoves = append(pbLegalMoves, toPBMove(m))
	}

	return &pb.State{
		Board:      flattenBoard,
		Current:    toPBCell(current),
		LegalMoves: pbLegalMoves,
	}
}

func toPBCell(cell game.Cell) pb.Cell {
	if cell == game.White {
		return pb.Cell_CELL_WHITE
	} else if cell == game.Black {
		return pb.Cell_CELL_BLACK
	}
	return pb.Cell_CELL_EMPTY
}

func toPBMove(move game.Move) *pb.Move {
	return &pb.Move{
		Row: int32(move.Row),
		Col: int32(move.Col),
	}
}

func (c *Client) SelectMove(ctx context.Context, st *game.State, modelID string) (game.Move, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	legalMoves := st.LegalMoves(st.Current)

	resp, err := c.stub.SelectMove(
		ctx,
		&pb.SelectMoveRequest{
			State:   toPBState(st, legalMoves),
			ModelId: modelID,
		},
	)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable, codes.DeadlineExceeded:
			return game.Move{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
		default:
			return game.Move{}, fmt.Errorf("aiclient: select move: %w", err)
		}
	}

	move := game.Move{
		Row: int(resp.GetMove().GetRow()),
		Col: int(resp.GetMove().GetCol()),
	}

	isLegal := false
	for _, lm := range legalMoves {
		if lm.Row == move.Row && lm.Col == move.Col {
			isLegal = true
			break
		}
	}

	if !isLegal {
		return game.Move{}, fmt.Errorf("%w: illegal move (%d,%d)", ErrBadResponse, move.Row, move.Col)
	}

	return move, nil
}
