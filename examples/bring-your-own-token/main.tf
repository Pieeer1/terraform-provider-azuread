terraform {
    required_providers {
        azuread = {
        source  = "terraform.test.com/alpha/azuread"
        version = "0.0.1"
        }
    }
}

provider "azuread" {
    use_pre_existing_token = true
    pre_existing_token = ""
}


resource "azuread_user" "example" {
  user_principal_name = "jdoe@hashicorp.com"
  display_name        = "J. Doe"
  mail_nickname       = "jdoe"
  password            = "SecretP@sswd99!"
}