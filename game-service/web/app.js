// ---------------------------------------------------------------------------
// Реверси — клиент. Общается с Go-сервером по /api/*.
//
// Ключевая идея рендера: 64 клетки создаются ОДИН раз, дальше DOM не
// перестраивается, а обновляется точечно. Только так фишка может физически
// перевернуться (CSS-transition на transform), а не «моргнуть» новым цветом.
// ---------------------------------------------------------------------------

const WHITE = 1;
const BLACK = 2;

const boardEl = document.getElementById("board");
const statusEl = document.getElementById("status");
const toastEl = document.getElementById("toast");
const overlayEl = document.getElementById("overlay");
const overlayResultEl = document.getElementById("overlay-result");
const overlayScoreEl = document.getElementById("overlay-score");

const players = {
  [WHITE]: {
    card: document.getElementById("player-white"),
    score: document.getElementById("score-white"),
    name: "White",
  },
  [BLACK]: {
    card: document.getElementById("player-black"),
    score: document.getElementById("score-black"),
    name: "Black",
  },
};

const cells = [];
let lastMove = null; // индекс клетки последнего хода
let busy = false; // запрос в полёте — клики игнорируем
let gameOver = false;
let toastTimer = null;

// --------------------------------------------------------------- построение

function buildBoard() {
  // точки-ориентиры на пересечениях 2/3 и 6/7 линий, как на настоящей доске
  const dots = new Set(["1,1", "1,5", "5,1", "5,5"]);

  for (let row = 0; row < 8; row++) {
    for (let col = 0; col < 8; col++) {
      const cell = document.createElement("div");
      cell.className = "cell";
      cell.dataset.row = row;
      cell.dataset.col = col;
      cell.setAttribute("role", "gridcell");

      if (dots.has(`${row},${col}`)) {
        cell.classList.add("has-dot");
      }

      const marker = document.createElement("span");
      marker.className = "marker";
      cell.append(marker);

      boardEl.append(cell);
      cells.push(cell);
    }
  }
}

function buildLabels() {
  const ranks = document.getElementById("ranks");
  const files = document.getElementById("files");

  for (let i = 0; i < 8; i++) {
    const rank = document.createElement("span");
    rank.textContent = i + 1;
    ranks.append(rank);

    const file = document.createElement("span");
    file.textContent = "abcdefgh"[i];
    files.append(file);
  }
}

function createDisc(isBlack) {
  const disc = document.createElement("div");
  disc.className = "disc";

  // цвет ставим ДО вставки в DOM: иначе браузер проиграет переворот
  // как переход из белого состояния в чёрное, хотя фишку только положили
  if (isBlack) {
    disc.classList.add("is-black");
  }

  const white = document.createElement("div");
  white.className = "disc__face disc__face--white";

  const black = document.createElement("div");
  black.className = "disc__face disc__face--black";

  disc.append(white, black);

  disc.classList.add("is-new");
  disc.addEventListener(
    "animationend",
    () => disc.classList.remove("is-new"),
    { once: true },
  );

  return disc;
}

// ------------------------------------------------------------------- рендер

function render(state) {
  gameOver = state.gameOver;

  const valid = new Set();
  for (const m of state.validMoves) {
    valid.add(m.row * 8 + m.col);
  }

  for (let i = 0; i < 64; i++) {
    const cell = cells[i];
    const value = state.board[(i / 8) | 0][i % 8];
    let disc = cell.querySelector(".disc");

    if (value === 0) {
      if (disc) disc.remove();
    } else if (!disc) {
      cell.append(createDisc(value === BLACK));
    } else {
      disc.classList.toggle("is-black", value === BLACK);
    }

    cell.classList.toggle("is-valid", !gameOver && valid.has(i));
    cell.classList.toggle("is-last", lastMove === i);
  }

  boardEl.dataset.turn = state.current === BLACK ? "black" : "white";

  renderScores(state);
  renderStatus(state);
}

function renderScores(state) {
  countTo(players[WHITE].score, state.whiteScore);
  countTo(players[BLACK].score, state.blackScore);

  for (const side of [WHITE, BLACK]) {
    const active = !state.gameOver && state.current === side;
    players[side].card.classList.toggle("is-active", active);
  }
}

function renderStatus(state) {
  if (state.gameOver) {
    statusEl.textContent = "Партия окончена";
    showOverlay(state);
    return;
  }

  overlayEl.hidden = true;

  const turn = players[state.current] ? players[state.current].name : "?";
  const left = 64 - state.whiteScore - state.blackScore;
  statusEl.textContent = `Ход ${turn} · свободно клеток: ${left}`;
}

function showOverlay(state) {
  let result;
  if (state.whiteScore > state.blackScore) result = "Победили White";
  else if (state.blackScore > state.whiteScore) result = "Победили Black";
  else result = "Ничья";

  overlayResultEl.textContent = result;
  overlayScoreEl.textContent = `${state.whiteScore} : ${state.blackScore}`;
  overlayEl.hidden = false;
}

// плавный «докрут» счёта — резкий скачок цифр выглядит дёшево
function countTo(el, value) {
  const from = Number(el.textContent) || 0;
  if (from === value) return;

  const started = performance.now();
  const duration = 380;

  function step(now) {
    const p = Math.min(1, (now - started) / duration);
    el.textContent = Math.round(from + (value - from) * p);
    if (p < 1) requestAnimationFrame(step);
  }

  requestAnimationFrame(step);
}

// -------------------------------------------------------------------- сеть

async function api(path, options) {
  const res = await fetch(path, options);

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    const err = new Error(text.trim() || `HTTP ${res.status}`);
    err.status = res.status;
    throw err;
  }

  return res.json();
}

async function loadState() {
  try {
    render(await api("/api/state"));
  } catch (err) {
    fail(err);
  }
}

async function newGame() {
  if (busy) return;

  busy = true;
  try {
    lastMove = null;
    overlayEl.hidden = true;
    render(await api("/api/new", { method: "POST" }));
    toast("Новая партия");
  } catch (err) {
    fail(err);
  } finally {
    busy = false;
  }
}

async function makeMove(row, col) {
  if (busy || gameOver) return;

  const cell = cells[row * 8 + col];

  // сервер тоже проверит легальность, но незачем гонять заведомо мимо
  if (!cell.classList.contains("is-valid")) {
    if (!cell.querySelector(".disc")) {
      deny(cell, "Сюда ходить нельзя");
    }
    return;
  }

  busy = true;
  boardEl.classList.add("is-busy");

  const mover = Number(boardEl.dataset.turn === "black" ? BLACK : WHITE);

  try {
    const state = await api("/api/move", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ row, col }),
    });

    lastMove = row * 8 + col;
    render(state);

    // ход не перешёл к сопернику — значит, у него не нашлось ходов
    if (!state.gameOver && state.current === mover) {
      const skipped = mover === WHITE ? BLACK : WHITE;
      toast(`У ${players[skipped].name} нет ходов — снова ходит ${players[mover].name}`);
    }
  } catch (err) {
    if (err.status === 400) {
      deny(cell, "Так ходить нельзя");
    } else {
      fail(err);
    }
  } finally {
    busy = false;
    boardEl.classList.remove("is-busy");
  }
}

// ------------------------------------------------------------ обратная связь

function toast(text) {
  toastEl.textContent = text;
  toastEl.classList.add("is-shown");

  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toastEl.classList.remove("is-shown"), 2400);
}

function deny(cell, text) {
  toast(text);
  cell.classList.add("is-denied");
  cell.addEventListener(
    "animationend",
    () => cell.classList.remove("is-denied"),
    { once: true },
  );
}

function fail(err) {
  console.error(err);
  statusEl.textContent = "Нет связи с сервером";
  toast("Сервер недоступен");
}

// ------------------------------------------------------------------- запуск

buildBoard();
buildLabels();

// один слушатель на всю доску вместо 64 отдельных
boardEl.addEventListener("click", (event) => {
  const cell = event.target.closest(".cell");
  if (!cell) return;
  makeMove(Number(cell.dataset.row), Number(cell.dataset.col));
});

document.getElementById("new-game").addEventListener("click", newGame);
document.getElementById("overlay-again").addEventListener("click", newGame);

loadState();
