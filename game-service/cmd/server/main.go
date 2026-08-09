package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/aiclient"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/game"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/web"
	"log"
	"net/http"
	"sync"
	"time"
)

type mover interface {
	SelectMove(ctx context.Context, st *game.State, modelID string) (game.Move, error)
}

type server struct {
	game 		*game.State
	ai   		mover
	modelId 	string
	mu   		sync.Mutex
}

type stateResponse struct {
	Board      [8][8]game.Cell `json:"board"`
	Current    game.Cell       `json:"current"`
	LegalMoves []moveDTO       `json:"legalMoves"`
	White      int             `json:"whiteScore"`
	Black      int             `json:"blackScore"`
	GameOver   bool            `json:"gameOver"`
}

type moveDTO struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type moveRequest struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

func (s *server) writeState(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	gm := s.game.LegalMoves(s.game.Current)

	dtos := make([]moveDTO, 0, len(gm))
	for _, m := range gm {
		dtos = append(dtos, moveDTO{
			Row: m.Row,
			Col: m.Col,
		})
	}

	ws, bs := s.game.Score()

	resp := stateResponse{
		Board:      s.game.Board,
		Current:    s.game.Current,
		LegalMoves: dtos,
		White:      ws,
		Black:      bs,
		GameOver:   s.game.IsGameOver(),
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Couldn't convert board into json: %v", err)
	}
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.writeState(w)
}

func (s *server) handleNew(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.game = game.New()

	s.writeState(w)
}

func (s *server) handleMove(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var req moveRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "couldn't decode user request", http.StatusBadRequest)
		return
	}

	m := game.Move{
		Row: req.Row,
		Col: req.Col,
	}

	if err := s.game.ApplyMove(m); err != nil {
		log.Printf("handleMove: %v", err)
		http.Error(w, "недопустимый ход", http.StatusBadRequest)
		return
	}

	s.writeState(w)
}

func (s *server) handleAIMove(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.game.IsGameOver() {
		http.Error(w, "game over", http.StatusBadRequest)
		return
	}

	legalMoves := s.game.LegalMoves(s.game.Current)
	if len(legalMoves) == 0 {
		s.game.NextTurn()
		s.writeState(w)
		return
	}

	move, err := s.ai.SelectMove(r.Context(), s.game, s.modelId)

	if err != nil {
		log.Printf("handleAIMove: %v", err)
		switch {
		case errors.Is(err, aiclient.ErrUnavailable):
			http.Error(w, "сервис модели недоступен", http.StatusServiceUnavailable)
		case errors.Is(err, aiclient.ErrBadResponse):
			http.Error(w, "модель вернула некорректный ход", http.StatusBadGateway)
		default:
			http.Error(w, "не удалось получить ход модели", http.StatusInternalServerError)
		}
		return
	}

	if err := s.game.ApplyMove(move); err != nil {
		log.Printf("handleAIMove: applying %v: %v", move, err)
		http.Error(w, "не удалось применить ход модели", http.StatusInternalServerError)
		return
	}
	s.writeState(w)
}

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := aiclient.New("127.0.0.1:50051", 2*time.Second)
	if err != nil {
		log.Fatal(fmt.Errorf("main: %w", err))
	}
	defer client.Close()

	srv := &server{
		game: 		game.New(),
		ai:   		client,
		modelId: 	"iter13-wr72",
	}

	// fmt.Println(svr.game.Current)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(web.Files)))

	mux.HandleFunc("GET /api/state", srv.handleState)
	mux.HandleFunc("POST /api/new", srv.handleNew)
	mux.HandleFunc("POST /api/move", srv.handleMove)
	mux.HandleFunc("POST /api/ai-move", srv.handleAIMove)

	log.Println("слушаю на http://localhost:8080")
	return http.ListenAndServe("127.0.0.1:8080", mux)
}
