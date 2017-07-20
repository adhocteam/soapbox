terraform {
  backend "s3" {
    # Copy backend.tfvars.sample to backend.tfvars, populate
    # with the appropriate values and initialize the backend with:
    # terraform init -backend-config=backend.tfvars
  }
}
