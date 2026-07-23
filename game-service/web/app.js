async function loadState() {
  const res = await fetch("/api/state");
  const state = await res.json();
  renderBoard(state);
}

loadState();


function renderBoard(state) {
    const board = document.querySelector(".board");
    board.innerHTML = "";                        // очистить перед перерисовкой

    for (let r = 0; r < 8; r++) {
        for (let c = 0; c < 8; c++) {
            const cell = document.createElement("div");
            cell.classList.add("cell");
            
            cell.addEventListener("click", () => makeMove(r, c));

            const isValid = state.validMoves.some(m => m.row === r && m.col === c);
            if (isValid) {
                cell.classList.add("valid");
            }

            const value = state.board[r][c];
            if (value === 1 || value === 2) {
                const piece = document.createElement("div");
                piece.classList.add("piece", value === 1 ? "white" : "black");
                cell.append(piece);
            }

            board.append(cell);
        }
    }
    renderStatus(state);
}


async function makeMove(r, c) {
  const res = await fetch("/api/move", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ row: r, col: c }),
  });

  if (!res.ok) {
    console.log("нельзя так ходить");   // сервер вернул 400
    return;
  }

  const state = await res.json();
  renderBoard(state);
}


function renderStatus(state) {
  const status = document.querySelector("#status");

  if (state.gameOver) {
    let result;
    if (state.whiteScore > state.blackScore) result = "победили White";
    else if (state.blackScore > state.whiteScore) result = "победили Black";
    else result = "ничья";
    status.textContent = `Игра окончена — ${result} (${state.whiteScore} : ${state.blackScore})`;
    return;
  }

  const turn = state.current === 1 ? "White" : "Black";
  status.textContent = `Ход: ${turn} — White ${state.whiteScore} : ${state.blackScore} Black`;
}


async function newGame() {
  const res = await fetch("/api/new", { method: "POST" });
  const state = await res.json();
  renderBoard(state);
}
document.querySelector("#new-game").addEventListener("click", newGame);