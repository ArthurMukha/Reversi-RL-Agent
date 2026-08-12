package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/aiclient"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/game"
)

type fakeMover struct {
	move  game.Move
	err   error
	calls int // сколько раз хендлер сходил к модели
}

// Ресивер указательный: со значением инкремент ушёл бы в копию,
// и calls всегда оставался бы нулём.
func (f *fakeMover) SelectMove(ctx context.Context, st *game.State, modelID string) (game.Move, error) {
	f.calls++
	return f.move, f.err
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "405 - у state нет метода POST",
			method:     "POST",
			path:       "/api/state",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "405 - у move нет метода GET",
			method:     "GET",
			path:       "/api/move",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "405 - у new нет метода GET",
			method:     "GET",
			path:       "/api/new",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "405 - у ai-move нет метода GET",
			method:     "GET",
			path:       "/api/ai-move",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "404 - GET /api/nope",
			method:     "GET",
			path:       "/api/nope",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "200 - GET /",
			method:     "GET",
			path:       "/",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &server{}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			srv.routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got Status = %d, want %d", rec.Code, tt.wantStatus)
			}

		})
	}
}

func render(b [8][8]game.Cell) string {
	cell2let := map[game.Cell]byte{
		game.Empty: '.',
		game.White: 'W',
		game.Black: 'B',
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

func assertMoves(t *testing.T, what string, got, want []moveDTO) {
	t.Helper()

	got, want = slices.Clone(got), slices.Clone(want)

	sortF := func(a, b moveDTO) int {
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

func decodeOK(t *testing.T, rec *httptest.ResponseRecorder) stateResponse {
	t.Helper()

	wantStatus := http.StatusOK
	wantCT := "application/json"

	// rec.Code
	if rec.Code != wantStatus {
		t.Errorf("response code: %d want %d", rec.Code, wantStatus)
	}

	// rec.Header().Get("Content-Type")
	gotCT := rec.Header().Get("Content-Type")
	if gotCT != wantCT {
		t.Errorf("response header content-type: %s, want %s", gotCT, wantCT)
	}

	// rec.Body.String()
	var got stateResponse
	err := json.Unmarshal(rec.Body.Bytes(), &got)
	if err != nil {
		t.Fatalf("json body не разбирается: %v", err)
	}
	return got
}

func TestHandleState(t *testing.T) {

	srv := &server{
		game: game.New(),
	}

	req := httptest.NewRequest("GET", "/api/state", nil)
	rec := httptest.NewRecorder()
	srv.handleState(rec, req)

	got := decodeOK(t, rec)

	assertState(t, got, game.New())

}

func TestHandleNew(t *testing.T) {

	srv := &server{game: game.New()}

	if err := srv.game.ApplyMove(game.Move{Row: 2, Col: 4}); err != nil {
		t.Fatalf("preparing: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/new", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)

	got := decodeOK(t, rec)

	assertState(t, got, game.New())

}

func board(rows [8]string) [8][8]game.Cell {

	let2cell := map[byte]game.Cell{
		'.': game.Empty,
		'W': game.White,
		'B': game.Black,
	}

	var b [8][8]game.Cell
	for idx := 0; idx < 64; idx++ {
		b[idx/8][idx%8] = let2cell[rows[idx/8][idx%8]]
	}

	return b
}

// assertState сверяет ответ хендлера с ожидаемым состоянием партии:
// доска, очередь хода, легальные ходы текущего игрока, счёт, конец партии.
func assertState(t *testing.T, got stateResponse, want *game.State) {
	t.Helper()

	wantLM := want.LegalMoves(want.Current)
	wantLMDTO := make([]moveDTO, len(wantLM))
	for i, move := range wantLM {
		wantLMDTO[i] = moveDTO{Row: move.Row, Col: move.Col}
	}
	wantW, wantB := want.Score()

	if got.Board != want.Board {
		t.Errorf("got.Board = \n%s\n, want \n%s\n", render(got.Board), render(want.Board))
	}
	if got.Current != want.Current {
		t.Errorf("got.Current = %v, want %v", got.Current, want.Current)
	}

	assertMoves(t, "got.LegalMoves", got.LegalMoves, wantLMDTO)

	if got.GameOver != want.IsGameOver() {
		t.Errorf("got.GameOver = %t, want %t", got.GameOver, want.IsGameOver())
	}
	if got.White != wantW {
		t.Errorf("got.White = %d, want %d", got.White, wantW)
	}
	if got.Black != wantB {
		t.Errorf("got.Black = %d, want %d", got.Black, wantB)
	}
}

// assertGame сверяет партию на сервере — то, что осталось после вызова
// хендлера. Нужен отдельно от assertState: в ветках с ошибкой ответа-JSON
// нет, а проверить, что доска не поехала, всё равно надо.
func assertGame(t *testing.T, got, want *game.State) {
	t.Helper()

	if got.Board != want.Board {
		t.Errorf("srv.game.Board = \n%s\n, want \n%s\n", render(got.Board), render(want.Board))
	}
	if got.Current != want.Current {
		t.Errorf("srv.game.Current = %v, want %v", got.Current, want.Current)
	}
}

// assertErr — пара к decodeOK для веток, где сработал http.Error:
// там тело text/plain, а не JSON.
func assertErr(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()

	wantCT := "text/plain; charset=utf-8"

	if rec.Code != wantStatus {
		t.Errorf("response code: %d want %d", rec.Code, wantStatus)
	}

	gotCT := rec.Header().Get("Content-Type")
	if gotCT != wantCT {
		t.Errorf("response header content-type: %s, want %s", gotCT, wantCT)
	}

	if rec.Body.Len() == 0 {
		t.Error("тело ответа пустое, ожидалось сообщение об ошибке")
	}
}

func TestHandleMove(t *testing.T) {

	tests := []struct {
		name      string
		body      string
		wantCode  int
		wantState *game.State
	}{
		{
			name:     "легальный ход для Белых",
			body:     `{"row":2,"col":4}`,
			wantCode: http.StatusOK,
			wantState: &game.State{
				Board: board([8]string{
					"........",
					"........",
					"....W...",
					"...WW...",
					"...BW...",
					"........",
					"........",
					"........",
				}),
				Current: game.Black,
			},
		},
		{
			name:      "ход легален для Чёрных, но не для Белых",
			body:      `{"row":2,"col":3}`,
			wantCode:  http.StatusBadRequest,
			wantState: game.New(),
		},
		{
			name:      "некорректный json",
			body:      "какой-то бред",
			wantCode:  http.StatusBadRequest,
			wantState: game.New(),
		},
		{
			name:      "координаты вне доски",
			body:      `{"row":99,"col":-1}`,
			wantCode:  http.StatusBadRequest,
			wantState: game.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &server{game: game.New()}

			req := httptest.NewRequest("POST", "/api/move", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			srv.routes().ServeHTTP(rec, req)

			if tt.wantCode == http.StatusOK {
				assertState(t, decodeOK(t, rec), tt.wantState)
			} else {
				assertErr(t, rec, tt.wantCode)
			}

			assertGame(t, srv.game, tt.wantState)
		})
	}
}

func TestHandleAIMove(t *testing.T) {

	// Только белые фишки: ходов нет ни у кого, партия окончена.
	gameOver := [8]string{
		"WW......",
		"WW......",
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
	}

	// Белым ходить нечем (за Чёрной фишкой край доски),
	// у Чёрных есть ровно один ход — (0,2).
	pass := [8]string{
		"BW......",
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
	}

	tests := []struct {
		name      string
		start     *game.State
		aiMove    game.Move
		aiErr     error
		wantCode  int
		wantCalls int
		wantState *game.State
	}{
		{
			name:      "модель сходила",
			start:     game.New(),
			aiMove:    game.Move{Row: 2, Col: 4},
			wantCode:  http.StatusOK,
			wantCalls: 1,
			wantState: &game.State{
				Board: board([8]string{
					"........",
					"........",
					"....W...",
					"...WW...",
					"...BW...",
					"........",
					"........",
					"........",
				}),
				Current: game.Black,
			},
		},
		{
			name:      "сервис модели недоступен",
			start:     game.New(),
			aiErr:     aiclient.ErrUnavailable,
			wantCode:  http.StatusServiceUnavailable,
			wantCalls: 1,
			wantState: game.New(),
		},
		{
			name:      "модель вернула некорректный ответ",
			start:     game.New(),
			aiErr:     aiclient.ErrBadResponse,
			wantCode:  http.StatusBadGateway,
			wantCalls: 1,
			wantState: game.New(),
		},
		{
			name:      "неизвестная ошибка клиента",
			start:     game.New(),
			aiErr:     errors.New("что-то сломалось"),
			wantCode:  http.StatusInternalServerError,
			wantCalls: 1,
			wantState: game.New(),
		},
		{
			name:      "модель вернула нелегальный ход",
			start:     game.New(),
			aiMove:    game.Move{Row: 0, Col: 0},
			wantCode:  http.StatusInternalServerError,
			wantCalls: 1,
			wantState: game.New(),
		},
		{
			name:      "партия окончена",
			start:     &game.State{Board: board(gameOver), Current: game.White},
			wantCode:  http.StatusBadRequest,
			wantCalls: 0,
			wantState: &game.State{Board: board(gameOver), Current: game.White},
		},
		{
			name:      "ходов нет, ход передаётся сопернику",
			start:     &game.State{Board: board(pass), Current: game.White},
			wantCode:  http.StatusOK,
			wantCalls: 0,
			wantState: &game.State{Board: board(pass), Current: game.Black},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := &fakeMover{move: tt.aiMove, err: tt.aiErr}
			srv := &server{game: tt.start, ai: ai, modelId: "test-model"}

			req := httptest.NewRequest("POST", "/api/ai-move", nil)
			rec := httptest.NewRecorder()
			srv.routes().ServeHTTP(rec, req)

			if tt.wantCode == http.StatusOK {
				assertState(t, decodeOK(t, rec), tt.wantState)
			} else {
				assertErr(t, rec, tt.wantCode)
			}

			assertGame(t, srv.game, tt.wantState)

			if ai.calls != tt.wantCalls {
				t.Errorf("обращений к модели: %d, want %d", ai.calls, tt.wantCalls)
			}
		})
	}
}

// func TestMain(m *testing.M) {

// 	// modelServiceAddr := envOr("MODEL_SERVICE_ADDR", "127.0.0.1:50051")
// 	// modelId := envOr("MODEL_ID", "iter13-wr72")

// 	// client, err := aiclient.New(modelServiceAddr, 2*time.Second)
// 	// if err != nil {
// 	// 	// error
// 	// }
// 	// defer client.Close()

// }
