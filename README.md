# shelf ツ

Terminal UI for managing CTF and box workspaces. Organizes your lab directory structure and spawns a tmux session for the selected target.

## Modes

| Mode | Structure |
|------|-----------|
| `ctf` | `$SHELF_BASE_DIR/training/challenges/<platform>/<category>/<challenge>` |
| `box` | `$SHELF_BASE_DIR/training/boxes/<platform>/<box>` |

`$SHELF_BASE_DIR` defaults to `~/work`. All directories are created automatically on selection — nothing needs to exist upfront. In CTF mode, picking an empty platform offers to generate default categories.


## Build

```bash
./build.sh
```

## Usage

All created directories and notes files will be under `$SHELF_BASE_DIR` backed up by ~/work. You can set this environment variable to change the base directory.
```bash
shelf        # pick mode interactively
shelf ctf    # start in CTF mode
shelf box    # start in box mode
```

On selection, a tmux session is created (or reused) for the target directory.

## Keybindings

| Key | Action |
|-----|--------|
| `↑↓` `j` `k` | Navigate |
| `enter` | Select |
| `esc` | Back |
| `/` | Filter list |
| `ctrl+f` | Search all dirs from current level |
| `n` | New |
| `r` | Rename |
| `d` | Delete |
| `q` | Quit |

Static `linux/amd64` binary, compressed with `upx` if available.
