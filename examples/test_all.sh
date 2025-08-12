#!/usr/bin/env bash

set -uo pipefail


name=$(tr -cd '[:alnum:]' <<< "$USER")

for os in * ; do
    if [[ ! -d "$os" ]]; then
        continue
    fi
    cd "$os" || exit
    for runtime in * ; do
        if [[ ! -d "$runtime" ]]; then
            continue
        fi
        app_name="$name-$runtime-$os-webapp"
        echo "Testing $app_name"
        curl "https://$app_name.azurewebsites.net"
    done
    cd ..
done

echo "✅ Done"
