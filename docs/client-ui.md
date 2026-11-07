# Browser interface

The browser client is a single-player game surface. The maze is the primary
artifact; supporting instructions remain beside it on desktop and above it on
small screens. Controls support keyboard and touch without a separate mode.

The selected palette is **Electric Blue**: pale blue surfaces, navy walls, a
blue player, and a yellow goal.

## Visual tokens

| Token | Value | Purpose |
| --- | --- | --- |
| Paper | `#f2f6ff` | Page background |
| Ink | `#1c2d53` | Primary text |
| Muted | `#526181` | Supporting text |
| Accent | `#315ee8` | Player and primary actions |
| Goal | `#ffcc4d` | Goal |
| Path | `#ffffff` | Walkable cells |
| Wall | `#243762` | Blocked cells |
| Line | `#ccd5ec` | Dividers and control borders |
| Frame | `#dce5fa` | Maze frame |
| Accent hover | `#254ac4` | Primary button hover |
| Control hover | `#e7edff` | Direction button hover |
| Goal ink | `#382a10` | Goal label |
| Error | `#b13836` | Request failures |

P and G labels distinguish the player and goal independently of color. DM Sans
is the body face and Manrope is used for titles, with sans-serif fallbacks.

## States and controls

Start exploring creates a game. New maze creates another game. Movement is
serialized while requests are in progress. Blocks return feedback without
increasing the successful move count. Winning updates the heading, announces
success, and disables movement. Refresh game reads server state. Errors are
announced and a new game can be requested after an API restart.

Focus outlines support keyboard navigation. Feedback uses a polite live region;
errors use an alert. The maze exposes player/goal coordinates as an image label,
but does not provide a full nonvisual wall-navigation representation yet.

The full maze is visible. Fog of war and multiplayer UI are outside this version.
