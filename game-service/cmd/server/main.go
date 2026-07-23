package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/game"
)

type server struct {
	game *game.State
}

type stateResponse struct {
	Board 		[8][8]game.Cell 	`json:"board"`
	Current		game.Cell			`json:"current"`
	ValidMoves 	[]moveDTO 			`json:"validMoves"`
	White 		int					`json:"whiteScore"`
	Black		int					`json:"blackScore"`
	GameOver 	bool				`json:"gameOver"`
}

type moveDTO struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type moveRequest struct {
	Row int		`json:"row"`
	Col int		`json:"col"`
}

func (s *server) WriteResponse(w http.ResponseWriter) {

	gm := s.game.ValidMoves(s.game.Current)

	dtos := make([]moveDTO, 0, len(gm))
	for _, m := range gm {
		dtos = append(dtos, moveDTO{
			Row: m.Row,
			Col: m.Col,
		})
	}

	ws, bs := s.game.Score()

	resp := stateResponse{
		Board: 		s.game.Board,
		Current: 	s.game.Current,
		ValidMoves: dtos,
		White: 		ws,
		Black: 		bs,
		GameOver:	s.game.IsGameOver(),
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Println(fmt.Sprintf("Couldn't convert board into json: %v", err))
	}
}

func (s *server) handleState(w http.ResponseWriter, r *http.Request) {

	if s.game == nil {
		http.Error(w, "нет активной игры", http.StatusInternalServerError)
		return
	}

	s.WriteResponse(w)
}

func (s *server) handleNew(w http.ResponseWriter, r *http.Request) {
	s.game = game.New()
	
	s.WriteResponse(w)
}

func (s *server) handleMove(w http.ResponseWriter, r *http.Request) {

	if s.game == nil {
		http.Error(w, "нет активной игры",  http.StatusBadRequest)
		return
	}

	var req moveRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "couldn't decode user request",  http.StatusBadRequest)
		return
	}

	m := game.Move{
		Row: req.Row,
		Col: req.Col,
	}

	if err := s.game.ApplyMove(m); err != nil {
		log.Println(fmt.Errorf("handleMove: %v", err))
		http.Error(w, "недопустимый ход",  http.StatusBadRequest)
		return
	}

	s.WriteResponse(w)
}

func main() {

	srv := &server{game.New()}

	// fmt.Println(svr.game.Current)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state",  srv.handleState)
	mux.HandleFunc("POST /api/new",  srv.handleNew)
	mux.HandleFunc("POST /api/move",  srv.handleMove)

	log.Println("слушаю на http://localhost:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}
