"use client";

import { useEffect, useState } from "react";
import { VoiceChat } from "@/components/voice-chat";

type Session = { code: string; token: string; role: "explorer" | "navigator" };
const sessionKey = "leftnt.multiplayer.session";

export default function Multiplayer() {
  const [session, setSession] = useState<Session | null>(null);
  const [code, setCode] = useState("");
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  useEffect(() => {
    const invite = new URL(window.location.href).searchParams.get("room");
    if (invite) setCode(invite.toUpperCase());
    try {
      const stored = JSON.parse(sessionStorage.getItem(sessionKey) ?? "null");
      if (stored && /^[A-Z2-7]{8}$/.test(stored.code) && typeof stored.token === "string" && ["explorer", "navigator"].includes(stored.role)) setSession(stored);
    } catch { /* An unavailable browser store does not prevent joining a room. */ }
    setReady(true);
  }, []);

  async function enter(join: boolean) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const response = await fetch(join ? `/api/rooms/${encodeURIComponent(code)}/join` : "/api/rooms", { method: "POST", cache: "no-store" });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error ?? "Could not enter the room.");
      const next = data as Session;
      try { sessionStorage.setItem(sessionKey, JSON.stringify(next)); } catch { /* The session remains usable in this tab. */ }
      setSession(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not reach the game server.");
    } finally { setBusy(false); }
  }

  function leave() {
    try { sessionStorage.removeItem(sessionKey); } catch { /* Nothing to clear. */ }
    setSession(null);
    setError("");
    setNotice("");
  }

  async function copyInvite() {
    if (!session) return;
    const invite = new URL("/multiplayer/", window.location.origin);
    invite.searchParams.set("room", session.code);
    try {
      await navigator.clipboard.writeText(invite.href);
      setNotice("Room link copied.");
    } catch { setNotice(`Share this room code: ${session.code}`); }
  }

  const parameters = session ? new URLSearchParams({ code: session.code, token: session.token }) : new URLSearchParams();
  const socketServer = process.env.NEXT_PUBLIC_GAME_WS_URL ?? (process.env.NODE_ENV === "development" ? "ws://127.0.0.1:8080" : "");
  if (socketServer) parameters.set("server", socketServer);
  const gameParams = parameters.toString();

  return (
    <main>
      <header className="masthead">
        <a href="/" aria-label="LEFTN’T home">LEFTN’T</a>
        <nav className="mode-nav" aria-label="Game modes">
          <a href="/engine/index.html">Solo garden</a>
          <a href="/multiplayer/" aria-current="page">Play together</a>
        </nav>
      </header>
      <section className={`coop-layout ${session ? "in-room" : ""}`} aria-labelledby="coop-title">
        <div className="coop-intro">
          <h1 id="coop-title">One maze.{" "}<br />Two perspectives.</h1>
          {!ready ? <p>Restoring your room…</p> : session ? (
            <>
              <p>You are the <strong>{session.role === "explorer" ? "Explorer" : "Navigator"}</strong>.</p>
              <div className="room-code"><span>Room code</span><strong>{session.code}</strong></div>
              <button className="primary" onClick={() => void copyInvite()}>Copy invite link</button>
              <p className="feedback" role="status">{notice || "Your partner joins with the room code."}</p>
              <VoiceChat session={session} socketServer={socketServer} />
              <div className="instructions">
                <h2>{session.role === "explorer" ? "Follow your navigator" : "Be their sense of direction"}</h2>
                <p>{session.role === "explorer" ? "Find the friend in red. Your navigator sees the whole garden. Lanterns brighten your path for 20 seconds." : "You see the whole map. Use the arrow keys, WASD, or direction buttons to guide the explorer toward the friend in red."}</p>
                <p>Click the game before using your keyboard. The game pauses when either player disconnects.</p>
              </div>
              <button className="text-button" onClick={leave}>Leave room</button>
            </>
          ) : (
            <>
              <p>One friend explores in the dark. The other has the map. Find the lost friend in red, together.</p>
              <button className="primary" disabled={busy} onClick={() => void enter(false)}>{busy ? "Opening room…" : "Create room as Explorer"}</button>
              <form className="join-room" onSubmit={(event) => { event.preventDefault(); void enter(true); }}>
                <label htmlFor="room-code">Have an invite? Join as Navigator</label>
                <div>
                  <input id="room-code" value={code} onChange={(event) => setCode(event.target.value.toUpperCase().replace(/[^A-Z2-7]/g, "").slice(0, 8))} placeholder="Room code" autoComplete="off" autoCapitalize="characters" spellCheck={false} minLength={8} maxLength={8} required />
                  <button className="primary" disabled={busy || code.length !== 8} type="submit">Join room</button>
                </div>
              </form>
              {error && <p className="error" role="alert">{error}</p>}
            </>
          )}
        </div>
        {session ? (
          <div className="coop-game">
            <iframe key={session.token} title={`LEFTN’T ${session.role} game`} src={`/engine/index.html#${gameParams}`} allow="autoplay" />
            <p className="feedback">{session.role === "explorer" ? "Arrows / WASD to move · T for walls · R for a new maze" : "Arrows / WASD send directions · Explorer starts the next maze"}</p>
          </div>
        ) : (
          <div className="coop-explanation">
            <h2>A little teamwork.<br />A lot of wrong turns.</h2>
            <ol>
              <li><strong>Explorer creates a room.</strong><span>Share the eight-character room code with a friend.</span></li>
              <li><strong>Navigator joins.</strong><span>Their map reveals the paths, lanterns, and lost friend.</span></li>
              <li><strong>Find your way together.</strong><span>The navigator sends directions. The explorer makes the moves.</span></li>
            </ol>
          </div>
        )}
      </section>
      <footer>Two players. One maze. Questionable directions.</footer>
    </main>
  );
}
