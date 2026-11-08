export type VoiceStatus = "idle" | "requesting" | "waiting" | "connecting" | "connected" | "error";
export type VoiceState = { status: VoiceStatus; muted: boolean; microphoneActive: boolean; error: string };
type VoiceSignal = {
  type: string;
  role?: "explorer" | "navigator";
  peerReady?: boolean;
  callId?: string;
  iceServers?: RTCIceServer[];
  description?: RTCSessionDescriptionInit;
  candidate?: RTCIceCandidateInit;
  error?: string;
};
type VoiceOptions = {
  endpoint: string;
  token: string;
  onState: (state: VoiceState) => void;
  onRemoteStream: (stream: MediaStream | null) => void;
  // Tests supply generated audio through the same media/connection lifecycle.
  getMedia?: () => Promise<MediaStream>;
};

export const initialVoiceState: VoiceState = { status: "idle", muted: false, microphoneActive: false, error: "" };

/** Owns one opt-in voice session, including all tracks, signaling, and cleanup. */
export class VoiceCall {
  private state: VoiceState = { ...initialVoiceState };
  private socket: WebSocket | null = null;
  private peer: RTCPeerConnection | null = null;
  private stream: MediaStream | null = null;
  private generation = 0;
  private callId = "";
  private pendingICE: RTCIceCandidateInit[] = [];
  private timeout: ReturnType<typeof setTimeout> | undefined;
  private signalChain: Promise<void> = Promise.resolve();

  constructor(private options: VoiceOptions) {}

  async join(): Promise<void> {
    if (!["idle", "error"].includes(this.state.status)) return;
    const generation = ++this.generation;
    this.setState({ ...initialVoiceState, status: "requesting" });
    let acquired: MediaStream | null = null;
    try {
      if (typeof RTCPeerConnection === "undefined") throw new Error("Voice is not supported by this browser.");
      const getMedia = this.options.getMedia ?? (() => {
        if (!navigator.mediaDevices?.getUserMedia) throw new Error("Voice needs HTTPS or localhost and a supported browser.");
        return navigator.mediaDevices.getUserMedia({ video: false, audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true } });
      });
      acquired = await getMedia();
      if (generation !== this.generation) { acquired.getTracks().forEach(track => track.stop()); return; }
      if (acquired.getAudioTracks().length === 0) throw new Error("No microphone audio is available.");
      this.stream = acquired;
      for (const track of acquired.getAudioTracks()) {
        track.onended = () => { if (generation === this.generation) this.fail("The microphone disconnected. Connect it and join again."); };
      }
      this.setState({ status: "connecting", microphoneActive: true });
      const socket = new WebSocket(this.options.endpoint);
      this.socket = socket;
      this.timeout = setTimeout(() => { if (generation === this.generation) this.fail("Could not reach voice chat. Join again to retry."); }, 10_000);
      socket.onopen = () => {
        if (generation === this.generation) socket.send(JSON.stringify({ type: "authenticate", token: this.options.token }));
      };
      socket.onmessage = event => {
        this.signalChain = this.signalChain.then(async () => {
          if (generation !== this.generation) return;
          await this.handle(JSON.parse(event.data) as VoiceSignal, generation);
        }).catch(() => { if (generation === this.generation) this.fail("Could not connect voice. Leave and join again."); });
      };
      socket.onclose = () => { if (generation === this.generation) this.fail("Voice disconnected. Join again to reconnect."); };
      socket.onerror = () => { if (generation === this.generation) this.fail("Could not reach voice chat. Join again to retry."); };
    } catch (error) {
      // Permission may resolve after leaving or unmounting. Never keep that track.
      if (acquired && acquired !== this.stream) acquired.getTracks().forEach(track => track.stop());
      if (generation !== this.generation) return;
      const name = error instanceof DOMException ? error.name : "";
      const message = name === "NotAllowedError" ? "Microphone permission denied. Allow it in browser settings and try again."
        : name === "NotFoundError" ? "No microphone found. Connect one and try again."
        : name === "NotReadableError" ? "The microphone is unavailable. Check your device and try again."
        : error instanceof Error ? error.message : "Could not start the microphone.";
      this.fail(message);
    }
  }

  toggleMute(): void {
    if (!this.stream) return;
    const muted = !this.state.muted;
    this.stream.getAudioTracks().forEach(track => { track.enabled = !muted; });
    this.setState({ muted });
  }

  leave(): void { ++this.generation; this.cleanup(); this.setState({ ...initialVoiceState }); }
  dispose(): void { ++this.generation; this.cleanup(); }

  private setState(update: Partial<VoiceState>): void {
    this.state = { ...this.state, ...update };
    this.options.onState({ ...this.state });
  }

  private fail(error: string): void {
    ++this.generation;
    this.cleanup();
    this.setState({ ...initialVoiceState, status: "error", error });
  }

  private closePeer(): void {
    clearTimeout(this.timeout);
    this.timeout = undefined;
    if (this.peer) {
      this.peer.onicecandidate = null;
      this.peer.ontrack = null;
      this.peer.onconnectionstatechange = null;
      this.peer.close();
      this.peer = null;
    }
    this.callId = "";
    this.pendingICE = [];
    this.options.onRemoteStream(null);
  }

  private cleanup(): void {
    this.closePeer();
    if (this.socket) {
      this.socket.onopen = this.socket.onmessage = this.socket.onclose = this.socket.onerror = null;
      if (this.socket.readyState < WebSocket.CLOSING) this.socket.close(1000, "left voice");
      this.socket = null;
    }
    const stream = this.stream;
    this.stream = null;
    stream?.getTracks().forEach(track => { track.onended = null; track.stop(); });
    this.signalChain = Promise.resolve();
  }

  private send(signal: VoiceSignal): void {
    if (this.socket?.readyState === WebSocket.OPEN) this.socket.send(JSON.stringify({ ...signal, callId: this.callId }));
  }

  private async handle(message: VoiceSignal, generation: number): Promise<void> {
    if (message.type === "error") { this.fail(message.error || "Could not join voice chat."); return; }
    if (message.type === "peer-left") {
      this.closePeer();
      this.setState({ status: "waiting" });
      return;
    }
    if (message.type === "ready") {
      if (!message.peerReady) {
        clearTimeout(this.timeout);
        this.setState({ status: "waiting" });
        return;
      }
      if (!message.callId || message.callId === this.callId) return;
      this.closePeer();
      this.callId = message.callId;
      this.setState({ status: "connecting" });
      const peer = new RTCPeerConnection({ iceServers: message.iceServers ?? [] });
      this.peer = peer;
      const current = () => generation === this.generation && peer === this.peer;
      peer.onicecandidate = event => { if (current() && event.candidate) this.send({ type: "candidate", candidate: event.candidate.toJSON() }); };
      peer.ontrack = event => {
        if (current() && event.track.kind === "audio") this.options.onRemoteStream(event.streams[0] ?? new MediaStream([event.track]));
      };
      peer.onconnectionstatechange = () => {
        if (!current()) return;
        if (peer.connectionState === "connected") {
          clearTimeout(this.timeout);
          this.setState({ status: "connected" });
        } else if (peer.connectionState === "failed") {
          this.fail("Voice could not connect across this network. Join again to retry.");
        } else if (peer.connectionState === "disconnected") {
          this.setState({ status: "connecting" });
          clearTimeout(this.timeout);
          this.timeout = setTimeout(() => { if (current() && peer.connectionState !== "connected") this.fail("Voice connection lost. Join again to reconnect."); }, 8_000);
        }
      };
      for (const track of this.stream?.getAudioTracks() ?? []) peer.addTrack(track, this.stream!);
      this.timeout = setTimeout(() => { if (current() && peer.connectionState !== "connected") this.fail("Voice could not connect. Check your network and join again."); }, 20_000);
      // Fixed two-person roles choose a single caller and avoid simultaneous offers.
      if (message.role === "explorer") {
        await peer.setLocalDescription(await peer.createOffer());
        if (current()) this.send({ type: "description", description: peer.localDescription!.toJSON() });
      }
      return;
    }
    if (message.callId !== this.callId || !this.peer) return;
    const peer = this.peer;
    const current = () => generation === this.generation && peer === this.peer;
    if (message.description) {
      await peer.setRemoteDescription(message.description);
      if (!current()) return;
      for (const candidate of this.pendingICE) await peer.addIceCandidate(candidate);
      this.pendingICE = [];
      if (message.description.type === "offer") {
        await peer.setLocalDescription(await peer.createAnswer());
        if (current()) this.send({ type: "description", description: peer.localDescription!.toJSON() });
      }
    } else if (message.candidate) {
      if (!peer.remoteDescription) {
        if (this.pendingICE.length >= 64) throw new Error("Too many pending candidates");
        this.pendingICE.push(message.candidate);
      } else { await peer.addIceCandidate(message.candidate); }
    }
  }
}
