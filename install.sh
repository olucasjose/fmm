#!/bin/sh
set -e

VERSION=$(date +'%y.%m.%d.%H%M')
echo "Compilando fmm (versão $VERSION)..."
go build -ldflags "-X 'fmm/internal/cli.Version=$VERSION'" -o fmm cmd/fmm/main.go

DEST_BIN="/usr/local/bin/fmm"
DEST_COMPLETION="/usr/share/bash-completion/completions/fmm"

echo "A instalação requer privilégios de root para gravar em $DEST_BIN e configurar o autocompletar."

# Instala o binário
sudo mv fmm "$DEST_BIN"
sudo chmod +x "$DEST_BIN"

# Gera e instala o script de autocompletar dinamicamente
echo "Gerando script de autocompletar nativo..."
sudo mkdir -p /usr/share/bash-completion/completions
"$DEST_BIN" completion bash | sudo tee "$DEST_COMPLETION" > /dev/null
sudo chmod 644 "$DEST_COMPLETION"

echo ""
echo "Sucesso! 'fmm' instalado em $DEST_BIN."
echo "Para ativar o autocompletar na sessão atual, execute:"
echo "source $DEST_COMPLETION"
