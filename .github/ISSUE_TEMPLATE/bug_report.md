---
name: Bug report
about: Create a report for an error, failure, or unexpected behavior
title: ''
labels: ''
assignees: ''

---

Please answer these questions when submitting your issue. Thanks!

1. What is your Terraform version? Run `terraform -v` to show the version. If you are not running the latest version of Terraform, please upgrade because your issue may have already been fixed.


2. Which operating system, processor architecture, and Go version are you using (`go env`)?


3. What are the affected resources? For example, kion_aws_iam_policy, kion_compliance_check, etc.


4. What does your Terraform configuration file look like?
```hcl
# Copy-paste your Terraform configurations here - for large Terraform configs,
# please use a service like Dropbox and share a link to the ZIP file. For
# security, you can also encrypt the files using our GPG public key.
```


5. Please provide a link to a GitHub Gist containing the complete debug output: https://www.terraform.io/docs/internals/debugging.html. Please do NOT paste the debug output in the issue; just paste a link to the Gist.


6. If Terraform produced a panic, please provide a link to a GitHub Gist containing the output of the `crash.log`.


7. What did you expect to see?


8. What did you actually see?


9. What steps can we run to reproduce the issue?
```bash
# Apply
terraform apply


```

10. Is there anything atypical about your accounts that we should know? For example: Running in EC2 Classic? Custom version of OpenStack? Tight ACLs?


11. Are there any other GitHub issues (open or closed) or Pull Requests that should be linked here?
