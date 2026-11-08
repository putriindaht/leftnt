"use client";

import { useEffect, useRef, useState } from "react";
import { initialVoiceState, VoiceCall, type VoiceState } from "@/lib/voice-call";

type VoiceSession = { code: string; token: string };

export function VoiceChat({ session, socketServer }: { session: VoiceSession; socketServer: string }) {
  const [state, setState] = useState<VoiceState>(initialVoiceState);
  const [soundBlocked, setSoundBlocked] = useState(false);
  const call = useRef<VoiceCall | null>(null);
  const audio = useRef<HTMLAudioElement | null>(null);

  useEffect(() => {
    let active = true;
    const endpoint = new URL(`/api/rooms/${session.code}/voice`, socketServer || window.location.origin);
    endpoint.protocol = ["https:", "wss:"].includes(endpoint.protocol) ? "wss:" : "ws:";
    const controller = new VoiceCall({
      endpoint: endpoint.href,
      token: session.token,
      onState: next => { if (active) setState(next); },
      onRemoteStream: stream => {
        const element = audio.current;
        if (!element) return;
        element.srcObject = stream;
        if (!stream) { element.pause(); if (active) setSoundBlocked(false); return; }
        void element.play().then(() => { if (active && element.srcObject === stream) setSoundBlocked(false); }).catch(() => { if (active && element.srcObject === stream) setSoundBlocked(true); });
      },
    });
    call.current = controller;
    setState(initialVoiceState);
    setSoundBlocked(false);
    return () => { active = false; controller.dispose(); if (call.current === controller) call.current = null; };
  }, [session.code, session.token, socketServer]);

  const joined = !["idle", "error"].includes(state.status);
  const labels: Record<VoiceState["status"], string> = {
    idle: "Join to talk with your partner.",
    requesting: "Waiting for microphone permission…",
    waiting: "Waiting for your partner to join voice…",
    connecting: "Connecting voice…",
    connected: "Voice connected. Talk to your partner.",
    error: "Voice is disconnected.",
  };

  return (
    <section className="voice-chat" aria-labelledby="voice-title">
      <h2 id="voice-title">Voice chat</h2>
      <p className="voice-status" role="status">{labels[state.status]}</p>
      <p className="microphone-state">{state.microphoneActive ? state.muted ? "Microphone muted" : "Microphone on" : "Microphone off"}</p>
      <div className="voice-controls">
        {joined ? (
          <>
            <button type="button" className="voice-button" disabled={!state.microphoneActive} aria-pressed={state.muted} onClick={() => call.current?.toggleMute()}>{state.muted ? "Unmute" : "Mute"}</button>
            <button type="button" className="voice-button" onClick={() => call.current?.leave()}>{state.status === "requesting" ? "Cancel" : "Leave call"}</button>
          </>
        ) : <button type="button" className="primary" onClick={() => void call.current?.join()}>Join voice</button>}
        {soundBlocked && <button type="button" className="voice-button" onClick={() => { void audio.current?.play().then(() => setSoundBlocked(false)).catch(() => setSoundBlocked(true)); }}>Enable sound</button>}
      </div>
      {state.error && <p className="error" role="alert">{state.error}</p>}
      <p className="voice-note">Both players choose Join voice. Leaving the room also turns off your microphone.</p>
      <audio ref={audio} autoPlay aria-label="Your partner’s voice" />
    </section>
  );
}
