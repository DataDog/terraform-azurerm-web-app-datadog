set -u
echo "$(tr -cd '[:alnum:]' <<< "$USER")-$example-webapp"
