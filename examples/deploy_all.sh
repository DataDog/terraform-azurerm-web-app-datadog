#!/usr/bin/env bash

set -auo pipefail

if [[ -z "$DD_API_KEY" ]]; then
    echo "Error: DD_API_KEY environment variable is not set."
    exit 1
fi

if ! command -v terraform &> /dev/null; then
    echo "Error: terraform command not found. Please install Terraform."
    exit 1
fi

sub_id=$(az account show --query id -o tsv)
export TF_IN_AUTOMATION=true
for example in * ; do
    if [[ ! -d "$example" ]]; then
        continue
    fi
    echo "Deploying $example"
    app_name=$(./name.sh)
    cd "$example" || exit
    echo "datadog_api_key = \"$DD_API_KEY\"
location = \"eastus2\"
name = \"$app_name\"
resource_group_name = \"$app_name-rg\"
subscription_id = \"$sub_id\"" > test.tfvars
    terraform init -upgrade || { echo "failed to init $example" && continue; }
    terraform apply -auto-approve -var-file=test.tfvars -compact-warnings &
    cd ..
done
wait

echo "✅ All resources have been deployed successfully 🚀"
