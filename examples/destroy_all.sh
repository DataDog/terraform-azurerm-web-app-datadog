#!/usr/bin/env bash

set -uo pipefail

if ! command -v terraform &> /dev/null; then
    echo "Error: terraform command not found. Please install Terraform."
    exit 1
fi

export TF_IN_AUTOMATION=true

for dir in * ; do
    if [[ ! -d "$dir" ]]; then
        continue
    fi
    echo "Destroying $dir resources"
    cd "$dir" || exit
    if [[ ! -f "test.tfvars" ]]; then
        echo "Error: test.tfvars file not found in $dir"
        continue
    fi
    if [[ $1 == "-f" || $1 == "--force" ]]; then
        az group delete -n "avasilver-$dir-linux-webapp-rg" --yes &
        cd .. && continue
    fi
    if [[ ! -f terraform.tfstate ]]; then
        echo "Error: terraform.tfstate file not found in $dir. Please deploy first."
        continue
    fi
    terraform destroy -auto-approve -var-file=test.tfvars -compact-warnings &
    cd ..
done
wait

echo "✅ All resources have been destroyed 💥"
