# Kion Terraform Provider Migration Script
#
# This script migrates your Kion resources that are managed by the legacy cloudtamer-io/cloudtamerio
# Terraform provider to the kionsoftware/kion provider.
#
# Usage:
#
# bash kion-provider-migration.sh <terraform root dir>
#
# The provided <terraform root dir> directory is expected to contain the folders for all managed resource modules.
# For example:
#
# ├── <terraform root dir>
# │ ├── aws-cloudformation-template
# │ ├── aws-iam-policy
# │ ├── azure-policy
# │ ├── azure-role
# │ ├── cloud-rule
# │ ├── compliance-check
# │ ├── compliance-standard
# │ ├── import_resource_state.sh
# │ ├── main.tf
# │ ├── ou-cloud-access-role
# │ ├── project-cloud-access-role
# │ ├── terraform.tfstate
# │ └── terraform.tfstate.backup
#
# Example:
#
# bash kion-provider-migration.sh /home/me/repos/gitlab.mycompany.com/infrastructure/kion/

# validate arguments
if [ "$#" -ne 1 ];
  then
  echo -e "\nUsage: $0 <terraform root dir> \n"
  exit 1
fi

# cd into provided root dir
cd "$1" || exit 1

# validate that the required ENV vars are present, exit if not
if ! env | grep CLOUDTAMERIO_API_KEY >> /dev/null; then
  echo "Environment variable not found: CLOUDTAMERIO_API_KEY. You must set this and then re-run the script"
  QUIT=1
fi
if ! env | grep CLOUDTAMERIO_URL >> /dev/null; then
  echo "Environment variable not found: CLOUDTAMERIO_URL. You must set this and then re-run the script"
  QUIT=1
fi
if [[ "$QUIT" -eq 1 ]]; then
  exit 1
fi

# export the env vars needed by the kion provider
export KION_APIKEY=$CLOUDTAMERIO_API_KEY
export KION_URL=$CLOUDTAMERIO_URL

# replace the provider. This updates the terraform.tfstate file
terraform state replace-provider cloudtamer-io/cloudtamerio kionsoftware/kion

# run an init to install the new kion provider
terraform init

# make replacements in the terraform.tfstate file
sed -i '' -e 's|cloudtamer-io/cloudtamerio|kionsoftware/kion|' terraform.tfstate
sed -i '' -e 's/cloudtamerio_/kion_/' terraform.tfstate

# make replacements in the resource files
for file in $(find . -type f -name "*.tf" | grep -v -E "(main|provider).tf"); do sed -i '' -e 's|cloudtamerio|kion|' "$file"; done

# save the kion provider's version to a variable, we'll need it next
KION_PROV_VER=$(grep -A 5 "kion" .terraform.lock.hcl  | grep version | awk '{print $3}' | tr -d '"')

# make final replacements to update the required_providers blocks
# in the main.tf and all provider.tf files
find . -name '*.tf' -exec sed -i '' -e 's|cloudtamer-io/cloudtamerio|kionsoftware/kion|' {} \;
find . -name '*.tf' -exec sed -i '' -e 's|cloudtamerio|kion|' {} \;
find . -name '*.tf' -exec sed -i '' -e "s|version = \".*\"|version = \"$KION_PROV_VER\"|" {} \;

echo
echo "Migration finished. Run a 'terraform plan' to ensure that no changes will be made."