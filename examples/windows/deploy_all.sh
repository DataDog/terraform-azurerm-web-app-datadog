#!/usr/bin/env bash

set -uo pipefail

if [[ -z "$DD_API_KEY" ]]; then
    echo "Error: DD_API_KEY environment variable is not set."
    exit 1
fi

if ! command -v terraform &> /dev/null; then
    echo "Error: terraform command not found. Please install Terraform."
    exit 1
fi

name=$(tr -cd '[:alnum:]' <<< "$USER")
sub_id=$(az account show --query id -o tsv)
export TF_IN_AUTOMATION=true

for dir in * ; do
    if [[ ! -d "$dir" ]]; then
        continue
    fi
    echo "Deploying $dir"
    cd "$dir" || exit
    echo "datadog_api_key = \"$DD_API_KEY\"
location = \"eastus2\"
name = \"$name-$dir-windows-webapp\"
resource_group_name = \"$name-$dir-windows-webapp-rg\"
subscription_id = \"$sub_id\"" > test.tfvars
    terraform init -upgrade || { echo "failed to init $dir" && continue; }
    terraform apply -auto-approve -var-file=test.tfvars -compact-warnings &
    cd ..
done
wait

echo "✅ All resources have been deployed successfully 🚀"
