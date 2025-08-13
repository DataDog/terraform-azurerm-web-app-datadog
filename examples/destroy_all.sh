#!/usr/bin/env bash

set -auo pipefail

if ! command -v terraform &> /dev/null; then
    echo "Error: terraform command not found. Please install Terraform."
    exit 1
fi

export TF_IN_AUTOMATION=true
for os in * ; do
    if [[ ! -d "$os" ]]; then
        continue
    fi
    cd "$os" || exit
    for runtime in * ; do
        if [[ ! -d "$runtime" ]]; then
            continue
        fi
        echo "Destroying $runtime resources"
        cd "$runtime" || exit
        if [[ ! -f "test.tfvars" ]]; then
            echo "Error: test.tfvars file not found in $runtime"
            continue
        fi
        if [[ ${1:-} == "-f" || ${1:-} == "--force" ]]; then
            az group delete -n "$(./name.sh)-rg" --yes &
            cd .. && continue
        fi
        if [[ ! -f terraform.tfstate ]]; then
            echo "Error: terraform.tfstate file not found in $runtime. Please deploy first."
            continue
        fi
        terraform destroy -auto-approve -var-file=test.tfvars -compact-warnings &
        cd ..
    done
    cd ..
done
wait

echo "✅ All resources have been destroyed 💥"
