"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  createGame,
  getGame,
  movePlayer,
  type Direction,
  type Game,
} from "@/lib/game-api";

const keys: Record<string, Direction> = {
  ArrowUp: "up",
  w: "up",
  W: "up",
  ArrowDown: "down",
  s: "down",
  S: "down",
  ArrowLeft: "left",
  a: "left",
  A: "left",
  ArrowRight: "right",
  d: "right",
  D: "right",
};

export default function Home() {
  const [game, setGame] = useState<Game | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("Your next wrong turn starts here.");
  const [error, setError] = useState("");
  const locked = useRef(false);

  const start = async () => {
    if (locked.current) return;
    locked.current = true;
    setBusy(true);
    setError("");
    try {
      setGame(await createGame());
      setMessage("Find G. Your other left might help.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create a game.");
    } finally {
      locked.current = false;
      setBusy(false);
    }
  };

  const move = useCallback(
    async (direction: Direction) => {
      if (!game || game.status !== "playing" || locked.current) return;
      locked.current = true;
      setBusy(true);
      setError("");
      try {
        const updated = await movePlayer(game.id, direction);
        setGame(updated);
        setMessage(
          updated.status === "won"
            ? "Goal reached! Your sense of direction survived."
            : updated.moved
              ? "Keep going. You’ve got this."
              : "That’s a wall. Try your other left.",
        );
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Could not move the player.",
        );
      } finally {
        locked.current = false;
        setBusy(false);
      }
    },
    [game],
  );

  const refresh = async () => {
    if (!game || locked.current) return;
    locked.current = true;
    setBusy(true);
    setError("");
    try {
      setGame(await getGame(game.id));
      setMessage("Game state refreshed.");
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not refresh the game.",
      );
    } finally {
      locked.current = false;
      setBusy(false);
    }
  };

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.ctrlKey || event.metaKey || event.altKey) return;
      const direction = keys[event.key];
      if (!direction || !game || game.status !== "playing") return;
      if (
        event.target instanceof HTMLElement &&
        event.target.closest("input, textarea, select, [contenteditable=true]")
      )
        return;
      event.preventDefault();
      void move(direction);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [game, move]);

  const canMove = !!game && game.status === "playing" && !busy;

  return (
    <main>
      <header className="masthead">
        <a href="/" aria-label="LEFTN’T home">
          LEFTN’T
        </a>
        <nav className="mode-nav" aria-label="Game modes">
          <a href="/engine/index.html">Solo garden</a>
          <a href="/multiplayer/">Play together</a>
        </nav>
      </header>
      <section className="game-layout" aria-labelledby="game-title">
        <div className="intro">
          <h1 id="game-title">
            A few wrong turns.
            <br />
            One way to the goal.
          </h1>
          <p>
            Find your way from <strong>P</strong> to <strong>G</strong>. The
            whole map is yours to explore.
          </p>
          <button
            className="primary"
            onClick={() => void start()}
            disabled={busy}
          >
            {busy && !game
              ? "Making your maze…"
              : game
                ? "New maze"
                : "Start exploring"}
          </button>
          <div className="instructions">
            <h2>Your directions</h2>
            <p>
              Use the arrow keys or <kbd>W</kbd> <kbd>A</kbd> <kbd>S</kbd>{" "}
              <kbd>D</kbd>. On a touch screen, use the buttons below the map.
            </p>
          </div>
          <dl className="legend">
            <div>
              <dt className="legend-player">P</dt>
              <dd>You</dd>
            </div>
            <div>
              <dt className="legend-goal">G</dt>
              <dd>Your goal</dd>
            </div>
            <div>
              <dt className="legend-wall" />
              <dd>Wall</dd>
            </div>
          </dl>
        </div>
        <div className="play-area">
          <div className="board-heading">
            <h2>
              {game?.status === "won" ? "You found your way." : "Find your way"}
            </h2>
            <span>{game ? `${game.moves} moves` : "Ready when you are"}</span>
          </div>
          <div className="board-frame" aria-busy={busy}>
            {game ? (
              <div
                className="maze"
                style={{
                  gridTemplateColumns: `repeat(${game.maze.width}, 1fr)`,
                }}
                role="img"
                aria-label={`Maze. Player at column ${game.player.x}, row ${game.player.y}. Goal at column ${game.goal.x}, row ${game.goal.y}.`}
              >
                {game.maze.cells.flatMap((row, y) =>
                  row.map((walkable, x) => {
                    const player = game.player.x === x && game.player.y === y;
                    const goal = game.goal.x === x && game.goal.y === y;
                    return (
                      <span
                        key={`${x}-${y}`}
                        aria-hidden="true"
                        className={`cell ${walkable ? "path" : "wall"} ${player ? "player" : goal ? "goal" : ""}`}
                      >
                        {player ? "P" : goal ? "G" : ""}
                      </span>
                    );
                  }),
                )}
              </div>
            ) : (
              <div className="empty-board">
                <div className="empty-route" aria-hidden="true">
                  <span>P</span>
                  <i />
                  <span>G</span>
                </div>
                <h3>A fresh maze is waiting.</h3>
                <p>Start exploring to find your way.</p>
              </div>
            )}
          </div>
          <p
            className={`feedback ${game?.status === "won" ? "won" : ""}`}
            role="status"
            aria-live="polite"
          >
            {message}
          </p>
          {error && (
            <p className="error" role="alert">
              {error}
            </p>
          )}
          <div className="board-controls">
            <div className="direction-pad" aria-label="Movement controls">
              {(["up", "left", "down", "right"] as Direction[]).map(
                (direction) => (
                  <button
                    key={direction}
                    className={`direction ${direction}`}
                    disabled={!canMove}
                    onClick={() => void move(direction)}
                    aria-label={`Move ${direction}`}
                  >
                    {direction}
                  </button>
                ),
              )}
            </div>
            {game && (
              <div className="position">
                <span>
                  Position ({game.player.x}, {game.player.y})
                </span>
                <button
                  className="text-button"
                  disabled={busy}
                  onClick={() => void refresh()}
                >
                  Refresh game
                </button>
              </div>
            )}
          </div>
        </div>
      </section>
      <footer>Made for curiosity. Built for a few wrong turns.</footer>
    </main>
  );
}
