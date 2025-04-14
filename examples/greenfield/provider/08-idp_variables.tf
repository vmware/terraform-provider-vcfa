variable "ldap_host" {
  type        = string
  description = "LDAP server"
}

variable "ldap_port" {
  type        = number
  description = "LDAP port"
}

variable "ldap_ssl" {
  type        = bool
  description = "LDAPS or LDAP"
}

variable "ldap_username" {
  type        = string
  description = "Username of LDAP"
}

variable "ldap_password" {
  type        = string
  description = "Password of LDAP"
  sensitive   = true
}

variable "ldap_searchbase" {
  type        = string
  description = "Base distinguished name for the LDAP"
}
