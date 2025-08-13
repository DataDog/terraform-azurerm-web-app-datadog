set -u
echo "$(tr -cd '[:alnum:]' <<< "$USER")-$runtime-$os-webapp"
