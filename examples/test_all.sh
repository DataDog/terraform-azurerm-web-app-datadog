#!/usr/bin/env bash

set -uo pipefail


name=$(tr -cd '[:alnum:]' <<< "$USER")

for dir in * ; do
    if [[ ! -d "$dir" ]]; then
        continue
    fi
    app_name="$name-$dir-linux-webapp"
    echo "Testing $app_name"
    open "https://$app_name.azurewebsites.net"
done

echo "✅ Done"
