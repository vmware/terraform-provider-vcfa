# Configure Identity Providers

# Configures the LDAP for the Provider (System) org. Read more:
# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/provider_ldap
resource "vcfa_provider_ldap" "example" {
  count = var.ldap_host == "" ? 0 : 1

  auto_trust_certificate  = true
  server                  = var.ldap_host
  port                    = var.ldap_port
  is_ssl                  = var.ldap_ssl
  username                = var.ldap_username
  password                = var.ldap_password
  base_distinguished_name = var.ldap_searchbase
  connector_type          = "ACTIVE_DIRECTORY"
  custom_ui_button_label  = "Hello, System"

  user_attributes {
    object_class                = "user"
    unique_identifier           = "objectGuid"
    display_name                = "displayName"
    username                    = "sAMAccountName"
    given_name                  = "givenName"
    surname                     = "sn"
    telephone                   = "telephoneNumber"
    group_membership_identifier = "dn"
    email                       = "mail"
    group_back_link_identifier  = "tokenGroups"
  }

  group_attributes {
    name                        = "cn"
    object_class                = "group"
    membership                  = "member"
    unique_identifier           = "objectGuid"
    group_membership_identifier = "dn"
    group_back_link_identifier  = "objectSid"
  }
}
