# Create GCS environment

```
# Set project
$ export CLOUDSDK_CORE_PROJECT=my-gcp-project

# Run terraform
$ terraform init
$ terraform apply

# Launch phlare with the following arguments:
$ terraform output -json | jq  -r "\"'\"+(.extra_flags.value | join(\"' '\"))+\"'\"" | xargs go run ../../../cmd/phlare/ --config.file ../../../cmd/phlare/phlare.yaml --phlaredb.max-block-duration 15s
```
