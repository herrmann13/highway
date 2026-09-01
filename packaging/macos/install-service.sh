#!/bin/zsh
set -eu

app_path="/Applications/Highway.app"
workflow_name="Abrir no Highway.workflow"
source_dir="${0:A:h}/${workflow_name}"
destination_dir="${HOME}/Library/Services/${workflow_name}"

if [[ ! -d "$app_path" ]]; then
  print -u2 "Instale Highway.app em /Applications antes de instalar o serviço."
  exit 1
fi

mkdir -p "${HOME}/Library/Services"
rm -rf "$destination_dir"
cp -R "$source_dir" "$destination_dir"
killall pbs 2>/dev/null || true
print "Serviço instalado em: $destination_dir"
print "O registro de Services do macOS foi atualizado."
print "Selecione um cURL, clique com o botão direito e procure 'Abrir no Highway' em Serviços."
