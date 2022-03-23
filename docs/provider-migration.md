# Terraform Provider Migration Guide

This guide contains the steps required to migrate your managed resources from the original cloudtamer-io/cloudtamerio Terraform provider
to the kionsoftware/kion provider.

There is a script to automate this process, but the steps are also detailed in the `Manual Migration` section.

## Automated Migration

This repository contains a script located at `provider-migration-script/kion-provider-migration.sh` that can automate
the migration process for you.

The script takes one argument, which is the path to the root directory of your Terraform managed Kion resources.

Here is an example:

```bash
<path to this repo>/provider-migration-script/kion-provider-migration.sh <path to terraform root dir>
```

## Manual Migration

1. Ensure the terraform state is in sync with your installation of Kion.

    The output of `terraform plan` should say:
      ```
      No changes. Your infrastructure matches the configuration.
      ```

2. Run the following command to replace the provider for your managed resources:

    `terraform state replace-provider cloudtamer-io/cloudtamerio kionsoftware/kion`

3. Review the output of the command above, and enter `yes` at the prompt.

4. Run `terraform init` to install the new provider. You should see output similar to below:

    ```
      Initializing modules...

      Initializing the backend...

      Initializing provider plugins...
      - Reusing previous version of cloudtamer-io/cloudtamerio from the dependency lock file
      - Finding latest version of kionsoftware/kion...
      - Using previously-installed cloudtamer-io/cloudtamerio v0.2.0
      - Installing kionsoftware/kion v0.3.0...
      - Installed kionsoftware/kion v0.3.0 (signed by a HashiCorp partner, key ID B72EFDFAF07C8126)

      .
      .
      .

      Terraform has been successfully initialized!
    ```

5. Export the new environment variables required by the new provider.

    ```bash
    export KION_URL=$CLOUDTAMERIO_URL
    export KION_API_KEY=$CLOUDTAMERIO_API_KEY
    ```

    **You'll need to ensure these variables are present in any automation that uses the Terraform provider.**

6. Run these commands to make the required replacments in the `terraform.tfstate` file:

    ```bash
    sed -i '' -e 's|cloudtamer-io/cloudtamerio|kionsoftware/kion|' terraform.tfstate
    sed -i '' -e 's/cloudtamerio_/kion_/' terraform.tfstate
    ```

7. Run this command to update all of the managed resource files:

    ```bash
    for file in $(find . -type f -name "*.tf" | grep -v -E "(main|provider).tf"); do sed -i '' -e 's|cloudtamerio|kion|' "$file"; done
   ```

8. Finally, run these commands to update the `main.tf` and all `provider.tf` files:

    ```bash
    KION_PROV_VER=$(grep -A 5 "kion" .terraform.lock.hcl  | grep version | awk '{print $3}' | tr -d '"')
    find . -name '*.tf' -exec sed -i '' -e 's|cloudtamer-io/cloudtamerio|kionsoftware/kion|' {} \;
    find . -name '*.tf' -exec sed -i '' -e 's|cloudtamerio|kion|' {} \;
    find . -name '*.tf' -exec sed -i '' -e "s|version = \".*\"|version = \"$KION_PROV_VER\"|" {} \;
    ```