# Highway

A fast, native API client for macOS and Linux. Organize, edit, and run your HTTP requests — locally, privately, no account required.

![Highway](docs/screenshot.png)

## Why Highway?

Highway keeps everything on your machine: your collections, requests, and variables live in plain JSON files in your own user directory — no cloud, no account, no telemetry. Paste a `curl` command and it becomes an editable request in one step, with familiar tabs and collections to keep your APIs organized.

## Installation

### macOS

Download the DMG for your architecture (Intel or Apple Silicon) from the [latest release](https://github.com/herrmann13/highway/releases), open it, and drag Highway into **Applications**. On first launch, right-click the app and choose **Open** to confirm you trust the unsigned build.

### Linux

Download the `.deb` package from the [latest release](https://github.com/herrmann13/highway/releases) and install it:

```bash
sudo apt install ./highway_0.1.0-1_amd64.deb
```

This installs the `highway` binary, an application-menu entry, and the app icon. Remove it at any time with `sudo apt remove highway`.

> Prefer to build it yourself? See [Building from source](#building-from-source).

## Features

### Organize your requests
- **Collections** group related requests and are saved locally as JSON.
- A **tree** on the left shows your collections and requests, with inline renaming (double-click) and quick deletion.
- Each request opens in its own **tab**; reopening one you already have open focuses its existing tab instead of duplicating it.
- Requests are **visually tagged by type** (HTTP, GraphQL, WebSocket, gRPC, SSE), each with its own icon.

### Build requests
- Separate tabs for **method & URL**, **Query Params**, **Headers**, **Body**, and **Authorization**.
- Body formats: **raw**, **`x-www-form-urlencoded`**, and **`multipart/form-data`**.

### Variables
- Reference `{{variable_name}}` in any field — URL, headers, body, or auth — and define the values per collection. Variables are expanded automatically before each request is sent.

### Authentication
- **No Auth**, **Basic Auth**, **Bearer Token**, **API Key** (header or query), **Digest Auth**, **OAuth 1.0**, and **OAuth 2.0** (client credentials, password, and more).
- Signing and token retrieval happen automatically, so you don't need to generate headers by hand.

### Send & inspect
- Send HTTP requests and read the result at a glance: **status codes are color-coded by range**, with **response time** and **response headers** shown.
- A **line-numbered response viewer** handles large bodies comfortably, up to **50 MB** per response.

### Import from cURL
- Paste a `curl` command and Highway converts the method, URL, query params, headers, body (raw, `--data-urlencode`, multipart `-F`), and auth (`-u`, Basic/Bearer via the `Authorization` header) into an editable request.
- **macOS Services integration**: select a `curl` command in any app, right-click → **Services → Open in Highway**, and the import dialog opens (or focuses your already-open window).
- Optional **clipboard auto-detection** captures `curl` commands as you copy them, keeping a handy history in the sidebar.

### Stay up to date
- **Check for updates** from the sidebar to compare your version against the latest GitHub release.
- Updates are **verified**: Highway picks the installer for your system, validates its published **SHA-256** checksum, and asks for confirmation before installing.

### Your data, your files
- Collections, requests, and variables are stored as JSON files in your user directory — no server, no account.
- Collections from previous project versions or names are **migrated automatically**.

## Quick start

1. Click **New Collection** to create a collection.
2. Add a request from the collection menu, or import one from `curl` (Options → Import → cURL).
3. Fill in the URL, hit **Send**, and inspect the response.

## Export & import

Use **Export** in the sidebar to save selected collections (with their requests and variables) to a JSON file. Restore them on another machine via Options → **Import → Collections**. Imports are additive and never delete existing data; duplicate names automatically get a suffix like `API (2)`.

## Building from source

Highway is built with Go and [Fyne](https://fyne.io).

### macOS

```bash
make macos
```

The app is created at `dist/Highway.app`. To build a distributable DMG:

```bash
make macos-dmg VERSION=0.1.0
```

### Linux (Ubuntu/Debian x86_64)

```bash
sudo apt install -y build-essential libgl1-mesa-dev xorg-dev libxkbcommon-dev libwayland-dev librsvg2-bin
make deb
```

## Releases

Pushing a `vMAJOR.MINOR.PATCH` tag triggers the GitHub release workflow, which produces macOS DMGs for Intel and Apple Silicon, the `.deb` for Linux, and the `SHA256SUMS` file used by the updater.

## License

Released under the [MIT License](LICENSE).
