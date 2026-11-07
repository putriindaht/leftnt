export type Position = { x: number; y: number };
export type Direction = "up" | "down" | "left" | "right";
export type Game = {
  id: string;
  maze: { width: number; height: number; cells: boolean[][] };
  player: Position;
  goal: Position;
  status: "playing" | "won";
  moves: number;
  moved?: boolean;
};

async function request(path: string, init?: RequestInit): Promise<Game> {
  let response: Response;
  try {
    response = await fetch(`/api${path}`, { ...init, cache: "no-store" });
  } catch {
    throw new Error(
      "Cannot reach the game server. Check both servers and try again.",
    );
  }
  let data;
  try {
    data = await response.json();
  } catch {
    throw new Error(
      "The game server is unavailable. Start the Go API and try again.",
    );
  }
  if (!response.ok) {
    throw new Error(data.error ?? "The request failed. Please try again.");
  }
  return data as Game;
}

export function createGame(): Promise<Game> {
  return request("/games", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ width: 15, height: 11 }),
  });
}

export function getGame(id: string): Promise<Game> {
  return request(`/games/${encodeURIComponent(id)}`);
}

export function movePlayer(id: string, direction: Direction): Promise<Game> {
  return request(`/games/${encodeURIComponent(id)}/moves`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ direction }),
  });
}
