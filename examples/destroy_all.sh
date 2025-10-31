#!/usr/bin/env bash

set -auo pipefail

if ! command -v terraform &> /dev/null; then
    echo "Error: terraform command not found. Please install Terraform."
    exit 1
fi

export TF_IN_AUTOMATION=true

for example in *; do
    if [[ ! -d "$example" ]]; then
        continue
    fi
    echo "Destroying $example"
    if [[ ${1:-} == "-f" || ${1:-} == "--force" ]]; then
        az group delete -n "$(./name.sh)-rg" --yes &
        continue
    fi
    cd "$example" || exit
    if [[ ! -f "test.tfvars" ]]; then
        echo "Error: test.tfvars file not found in $example"
        continue
    fi
    if [[ ! -f terraform.tfstate ]]; then
        echo "Error: terraform.tfstate file not found in $example. Please deploy first."
        continue
    fi
    terraform destroy -auto-approve -var-file=test.tfvars -compact-warnings &
    cd ..
done
wait

echo "✅ All resources have been destroyed 💥"
