# Highway

Highway é um cliente HTTP desktop, leve e nativo (construído em Go com [Fyne](https://fyne.io)), para organizar, editar e executar requisições de API — uma alternativa simples a ferramentas como Postman ou Insomnia, com integração direta ao macOS e Linux.

## Funcionalidades

### Organização de requisições
- **Coleções**: agrupe requisições relacionadas em coleções nomeadas, persistidas localmente em JSON.
- **Árvore de navegação**: visualize coleções e requisições em uma árvore lateral, com renomeação inline (duplo clique) e exclusão.
- **Abas**: cada requisição aberta vira uma aba de edição; reabrir uma requisição já aberta apenas foca a aba existente em vez de duplicá-la.
- **Tipos de requisição**: identificação visual por tipo (HTTP, GraphQL, WebSocket, gRPC, SSE), cada um com ícone próprio na árvore.

### Editor de requisições
- Método, URL, **Query Params**, **Headers**, **Body** (raw, `x-www-form-urlencoded` ou `multipart/form-data`) e **Authorization**, cada um em sua própria aba.
- **Variáveis**: use `{{nome_da_variavel}}` em qualquer campo (URL, headers, body, autenticação) e defina os valores por coleção; a expansão acontece automaticamente antes do envio.
- **Autenticação integrada**: No Auth, Basic Auth, Bearer Token, API Key (header ou query), Digest Auth, OAuth 1.0 e OAuth 2.0 (client credentials, entre outros grant types), com assinatura e obtenção de token feitas internamente.

### Execução e resposta
- Envio de requisições HTTP com suporte a status code colorido por faixa, tempo de resposta e headers de retorno.
- Visualizador de resposta com numeração de linhas, otimizado para corpos grandes (limite de 50 MB por resposta).

### Importação via cURL
- **Importe comandos `curl` diretamente**, convertendo automaticamente método, URL, query params, headers, body (raw, `--data-urlencode`, multipart `-F`) e autenticação (`-u`, Basic/Bearer via header `Authorization`) em uma requisição editável.
- **Integração com o menu de contexto do macOS**: selecione um comando `curl` em qualquer app, clique com o botão direito → `Services > Abrir no Highway`, e o Highway abre (ou foca a janela já aberta) com o diálogo de importação preenchido.
- **Detecção automática via área de transferência**: ao copiar um comando `curl`, o Highway pode identificar o conteúdo automaticamente.
- Comunicação entre instâncias via socket Unix local, garantindo que apenas uma instância do app rode por vez e que novas importações sejam roteadas para a janela já aberta.

### Armazenamento
- Coleções, requisições e variáveis são salvas como arquivos JSON no diretório de configuração do usuário, sem dependência de servidor ou conta.
- Migração automática de coleções de versões/nomes anteriores do projeto.

## Instalação e build

### macOS

Requisitos: Go, Xcode Command Line Tools e conexão com a internet na primeira execução (para baixar a ferramenta Fyne).

```bash
make macos
```

O app será criado em `dist/Highway.app`. Arraste-o para `Aplicativos` ou abra-o diretamente. Como o build não é assinado, o macOS pode exigir clicar com o botão direito e escolher `Abrir` na primeira execução.

Para habilitar "Abrir no Highway" no menu de contexto do Finder/apps:

```bash
make macos-service
```

### Linux (Ubuntu/Debian x86_64)

```bash
sudo apt update
sudo apt install -y build-essential libgl1-mesa-dev xorg-dev libxkbcommon-dev librsvg2-bin
make linux
```

Gera `dist/highway-linux-amd64.tar.gz`, contendo o binário, o lançador `highway.desktop` e o ícone. Para instalar apenas para o usuário atual:

```bash
mkdir -p ~/.local/bin ~/.local/share/applications ~/.local/share/icons/hicolor/512x512/apps
cp dist/highway ~/.local/bin/highway
cp dist/highway.desktop ~/.local/share/applications/highway.desktop
cp dist/highway.png ~/.local/share/icons/hicolor/512x512/apps/highway.png
chmod +x ~/.local/bin/highway
```

Garanta que `~/.local/bin` esteja no `PATH` e abra o Highway pelo menu de aplicativos ou com `highway` no terminal.

## Testes

```bash
make test
```

## Stack técnica

- **Go** + **[Fyne](https://fyne.io)** para a interface gráfica nativa multiplataforma.
- Empacotamento nativo para macOS (`.app`) e Linux (binário + `.desktop`), via `fyne package` e Makefile.
