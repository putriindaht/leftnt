# Browser interface

The browser client includes the original single-player surface and the
Explorer/Navigator lobby with an embedded Ebitengine game. The maze is the primary
artifact; supporting instructions remain beside it on desktop and above it on
small screens. Controls support keyboard and touch without a separate mode.

The selected palette is **Garden**: warm parchment, dark brown text, leaf-green
actions, wood-colored frames, sandy paths, hedge-green walls, and lantern gold.
It follows the cozy garden colors in the Ebitengine renderer. The same tokens
style the home page, multiplayer lobby, room information, forms, and voice controls.

## Visual tokens

| Token | Value | Purpose |
| --- | --- | --- |
| Paper | `#f4eedf` | Warm parchment background |
| Ink | `#3e3025` | Dark brown primary text |
| Muted | `#72604c` | Earth-brown supporting text |
| Accent | `#466335` | Leaf-green actions, links, and focus |
| On accent | `#fff8e7` | Light labels on solid actions |
| Surface | `#fffaf0` | Inputs and secondary controls |
| Panel | `#e6dfc9` | Quiet multiplayer explanation surface |
| Player | `#815839` | Brown P marker in the original browser maze |
| Goal | `#efc16d` | Lantern-gold goal |
| Goal ink | `#4b351b` | Goal label |
| Path | `#eddbb9` | Sandy walkable cells |
| Wall | `#3c5431` | Hedge-green blocked cells |
| Line | `#c5b596` | Soft dividers |
| Control border | `#958365` | Distinct control boundaries |
| Frame | `#cdb58e` | Wood-colored maze frame |
| Accent hover | `#334c27` | Deep-green primary button hover |
| Control hover | `#e8eddb` | Pale leaf-green secondary hover |
| Error | `#a13d2c` | Brick-red failures |

P and G labels distinguish the player and goal independently of color. DM Sans
is the body face and Manrope is used for titles, with sans-serif fallbacks.
The Ebitengine character sprites keep their own palettes. The browser theme
also colors text selection, input carets, native accent controls, and scrollbars.

## States and controls

Start exploring creates a game. New maze creates another game. Movement is
serialized while requests are in progress. Blocks return feedback without
increasing the successful move count. Winning updates the heading, announces
success, and disables movement. Refresh game reads server state. Errors are
announced and a new game can be requested after an API restart.

Focus outlines support keyboard navigation. Feedback uses a polite live region;
errors use an alert. The maze exposes player/goal coordinates as an image label,
but does not provide a full nonvisual wall-navigation representation yet.

The original browser prototype shows its full maze. Multiplayer uses the
role-filtered Ebitengine views described in [multiplayer](multiplayer.md).
