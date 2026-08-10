# shelf

Terminal picker for creating and opening CTF challenge and box workspaces.

Shelf runs on Linux. The release binary is built for x86-64 and uses `tmux` by default.

## Install

```bash
mkdir -p "$HOME/.local/bin"
curl -fsSL https://github.com/rh33t/shelf/releases/latest/download/shelf -o "$HOME/.local/bin/shelf"
chmod +x "$HOME/.local/bin/shelf"
shelf --version
```

Building from source requires Go 1.22 or newer:

```bash
git clone https://github.com/rh33t/shelf.git
cd shelf
go build -trimpath -o shelf .
install -Dm755 shelf "$HOME/.local/bin/shelf"
```

## Usage

Run `shelf`, choose `ctf` or `box`, then select or create a target. `shelf ctf` and `shelf box` skip the first picker.

The default workspace paths are:

```text
~/work/challenges/<source>/<category>/<challenge>
~/work/boxes/<platform>/<box>
```

New targets receive a `notes.md` file and open in a tmux session. The main keys are arrow keys or `hjkl` to navigate, `/` to filter, `ctrl+f` to search, `n` to create, `r` to rename, `d` to delete, and `q` to quit.

Shelf creates `~/.config/shelf/config.yaml` on first use. Run `shelf --config` to print the active path. The file controls the base directory, open command, CTF sources and categories, and box platforms. `SHELF_BASE_DIR` overrides `base_dir`.

For example, open targets in an editor instead of tmux:

```yaml
cmd: "$EDITOR $path"
```

The command can use `$session` for the target directory name and `$path` for its full path.

## Development

```bash
go test ./...
go build -o shelf .
```
