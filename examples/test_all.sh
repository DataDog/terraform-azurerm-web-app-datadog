#!/usr/bin/env bash

set -auo pipefail


for os in * ; do
    if [[ ! -d "$os" ]]; then
        continue
    fi
    cd "$os" || exit
    for runtime in * ; do
        if [[ ! -d "$runtime" ]]; then
            continue
        fi
        app_name=$(../name.sh)
        echo "========== Testing $app_name =========="
        curl "https://$app_name.azurewebsites.net"
        echo -e "\n\n"
    done
    cd ..
done

echo "✅ Done"
