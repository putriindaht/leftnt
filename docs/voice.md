# Room voice chat

## Use voice

Open `/multiplayer/`, create/join a room, then each player chooses **Join voice**.
Allow the browser's microphone request. Voice works alongside the Ebitengine
browser game; the direction buttons remain available.

- **Mute / Unmute** disables or enables outgoing microphone audio.
- **Leave call** closes voice and stops local microphone tracks.
- **Leave room** also closes voice and stops the microphone.
- When a partner leaves voice, the remaining participant waits for them to join
  again. Voice rejoining creates a new call without resetting the game.
- If voice signaling or its connection fails, microphone capture stops. Choose
  Join voice explicitly to try again.
- If browser autoplay blocks the incoming stream, choose **Enable sound**.

Microphone access starts only after Join voice. Cancel also handles a microphone
permission prompt that resolves late: any returned track is stopped immediately.
Voice is audio-only. The application does not record or persist the audio.

This feature is in the browser lobby. Standalone desktop Ebitengine still uses
the game's socket controls; browser voice is managed by the surrounding page.

## How it connects

A separate authenticated WebSocket at `/api/rooms/{code}/voice` carries WebRTC
setup. The same seat token authorizes it, independently of the game socket.
Only the other seat in that room receives a signal. Explorer creates the offer;
Navigator answers. Both exchange ICE candidates. Candidates arriving before a
remote description are queued until that description is installed.

The server creates a fresh call ID whenever two voice participants join. Old
candidates are ignored. One voice socket is allowed per seat; signaling is
bounded and checked for allowed message types and sizes. Setup never includes
the game map. Joining/leaving voice does not alter game connection flags.

The browser's `RTCPeerConnection` carries audio, either directly between players
or through an operator-configured TURN relay. Go relays setup messages only.

## Local tests and Internet hosting

Microphone access requires HTTPS or a secure local origin such as localhost.
Same-machine tests can use the Docker app at `http://localhost:8080/multiplayer/`
with two browser tabs. Use headphones when testing live audio near another mic.

The default ICE list is empty, suitable for local testing. Cross-network Internet
calls need configured STUN/TURN infrastructure; no hosted relay is provisioned.
For a public deployment, use HTTPS/WSS and provide working TURN settings for
networks that cannot connect directly.

### Private HTTPS through Tailscale

When the Mac and phone are connected to the same Tailscale network, Tailscale
Serve can provide trusted private HTTPS for testing phone microphone access.
On a Mac with the Tailscale app already installed and signed in:

```sh
/Applications/Tailscale.app/Contents/MacOS/Tailscale serve --bg --https=8443 http://127.0.0.1:8080
```

Open the HTTPS URL printed by the command, followed by `/multiplayer/`, on both
devices. Connect the phone's Tailscale app using the same account/network first.
Create a new room on this origin, then both players select Join voice and allow
the microphone. This Serve endpoint is private to the Tailscale network.

Stop this test proxy with:

```sh
/Applications/Tailscale.app/Contents/MacOS/Tailscale serve --https=8443 off
```

Port 8443 keeps this test separate from an existing Tailscale HTTPS service on
443. Docker must remain running for the proxied game to respond.

Set the server's `WEBRTC_ICE_SERVERS` variable to a JSON array. Docker Compose
reads it from the environment or a root `.env` file, which is excluded from Git:

```dotenv
WEBRTC_ICE_SERVERS=[{"urls":["stun:YOUR_HOST:3478"]},{"urls":["turn:YOUR_HOST:3478","turns:YOUR_HOST:5349"],"username":"YOUR_USERNAME","credential":"YOUR_CREDENTIAL"}]
```

Use your own relay and suitably short-lived credentials for public hosting.
Settings are sent only after room authentication. Browsers need the TURN
credentials to use the relay, so avoid privileged long-lived secrets. Do not
put relay secrets in `NEXT_PUBLIC_*` variables. Configure the relay credentials
and network/firewall ports on the chosen hosting platform before testing across
networks. Changing server settings requires restarting/recreating the container.

## Source map

- `client/src/lib/voice-call.ts`: opt-in capture, signaling, peer negotiation,
  ICE buffering, mute, and track cleanup.
- `client/src/components/voice-chat.tsx`: controls, state, and remote audio playback.
- `server/internal/usecase/voice.go`: room authentication, pairing, and relay.
- `server/internal/delivery/http/voice_handler.go`: voice socket transport.
- `server/internal/protocol/voice.go`: setup message types and ICE settings.

Backend tests cover authenticated relay, room isolation, roles, fresh call IDs,
duplicate subscriptions, bounded queues, and cleanup without affecting game
seats. Browser media testing uses generated audio through the production voice
controller, without using a physical microphone.

To reproduce the browser media test, install the client dependencies, then from
the repository root run:

```sh
./scripts/test-voice.sh
```

Open `http://localhost:8789` and choose **Run synthetic call test**. The fixture
uses the production controller, generated audio tracks, real peer connections,
and actual WebSocket signaling. It checks received audio, mute/unmute, rejoining,
pending-permission cancellation, permission denial, and capture cleanup after an
authentication error. Tones play at very low volume. The generated files stay
under `.local`; the test source is `client/tests/voice-call.html`. Stop the test
server with Ctrl+C. Use `VOICE_TEST_PORT` to choose another port.
