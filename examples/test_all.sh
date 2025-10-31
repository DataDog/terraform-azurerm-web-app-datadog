#!/usr/bin/env bash

set -auo pipefail

for example in * ; do
    if [[ ! -d "$example" ]]; then
        continue
    fi
    app_name=$(./name.sh)
    echo "========== Testing $app_name =========="
    curl "https://$app_name.azurewebsites.net"
    echo -e "\n\n"
done

echo "✅ Done"
