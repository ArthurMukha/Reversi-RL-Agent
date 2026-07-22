package main

import (
	"fmt"
	// "log"
	// "net/http"
	"github.com/ArthurMukha/reversi-rl-agent/game-service/internal/game"
)

func main() {
	// mux := http.NewServeMux()
	// mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintln(w, "Hello, world!")
	// })

	// log.Println("слушаю на http://localhost:8080")
	// log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
	g := game.New()
	fmt.Println(g.ValidMoves(1))
	fmt.Println(g.ValidMoves(2))
}
