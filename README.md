# Highway

Cliente HTTP desktop para organizar e executar requisicoes de API.

## Build local no macOS

Requisitos: Go, Xcode Command Line Tools e conexao com a internet na primeira execucao para baixar a ferramenta Fyne.

```bash
make macos
```

O app sera criado em `dist/Highway.app`. Arraste-o para `Aplicativos` ou abra-o diretamente. Como o build nao e assinado, o macOS pode exigir clicar com o botao direito e escolher `Abrir` na primeira execucao.

## Build local no Ubuntu/Debian x86_64

Instale as dependencias graficas e de compilacao:

```bash
sudo apt update
sudo apt install -y build-essential libgl1-mesa-dev xorg-dev libxkbcommon-dev librsvg2-bin
```

Na maquina Linux, execute:

```bash
make linux
```

O arquivo `dist/highway-linux-amd64.tar.gz` contem o binario, o lancador `highway.desktop` e o icone.

Para instalar apenas para o usuario atual:

```bash
mkdir -p ~/.local/bin ~/.local/share/applications ~/.local/share/icons/hicolor/512x512/apps
cp dist/highway ~/.local/bin/highway
cp dist/highway.desktop ~/.local/share/applications/highway.desktop
cp dist/highway.png ~/.local/share/icons/hicolor/512x512/apps/highway.png
chmod +x ~/.local/bin/highway
```

Garanta que `~/.local/bin` esteja no `PATH` e abra Highway pelo menu de aplicativos ou com `highway` no terminal.

## Testes

```bash
make test
```
