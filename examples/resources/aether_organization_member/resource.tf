# The address must already have a Norkeel login (Bridge invitation flow).
# Changing role REPLACES the membership; the last owner cannot be removed.
# Write the email in lower case; the API compares member emails case-insensitively.
# After removing an owner, revoke the API tokens they created (the Bridge, Settings → API Tokens).
resource "aether_organization_member" "ops" {
  email = "ops@example.com"
  role  = "admin"
}
