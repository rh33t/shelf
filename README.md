# shelf ツ

Terminal UI for managing CTF and box workspaces. Organizes a lab directory structure and spawns a tmux session for the selected target.

<p align="center"><img src="shelf.gif" alt="shelf demo" /></p>

## Modes

| Mode | Structure |
|------|-----------|
| `ctf` | `$SHELF_BASE_DIR/training/challenges/<platform>/<category>/<challenge>` |
| `box` | `$SHELF_BASE_DIR/training/boxes/<platform>/<box>` |

`$SHELF_BASE_DIR` defaults to `~/work`. Directories are created automatically on selection — nothing needs to exist upfront. In CTF mode, selecting an empty platform prompts to generate default categories.

## Requirements

- Unix-like OS (Linux, macOS, WSL)
- `tmux`

## Build

Produces a static `linux/amd64` binary, compressed with `upx` if available.

```bash
./build.sh
```

## Usage

```bash
shelf        # select mode interactively
shelf ctf    # launch in CTF mode
shelf box    # launch in box mode
```

Set `$SHELF_BASE_DIR` to override the default workspace root (`~/work`). On selection, a tmux session is created or reused for the target directory.

## Keybindings

| Key | Action |
|-----|--------|
| `↑` `↓` `j` `k` | Navigate |
| `enter` | Select |
| `esc` | Go back |
| `/` | Filter current list |
| `ctrl+f` | Search all directories from current level |
| `n` | Create new entry |
| `r` | Rename entry |
| `d` | Delete entry |
| `q` | Quit |

## Todo

- [ ] Support custom commands on selection instead of spawning a tmux session