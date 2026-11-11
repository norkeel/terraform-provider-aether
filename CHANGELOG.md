# Changelog

Release notes for the GitHub Release are written from this file; GoReleaser's
generated changelog is disabled (`.goreleaser.yml`).

## 0.2.0 (Unreleased)

Aether is now Norkeel, and the provider is renamed with it. This is the first
release under the new name; v0.1.0 remains available under the old address.

BREAKING CHANGES:

* provider: the address is now `norkeel/norkeel` and the provider block is
  `provider "norkeel"`. The old `plane-shift/aether` address gets no further
  releases.
* provider: every resource and data source is renamed from `aether_*` to
  `norkeel_*` (for example `norkeel_cluster` → `norkeel_cluster`).
* provider: the environment variables are now `NORKEEL_API_TOKEN`,
  `NORKEEL_ENDPOINT` and `NORKEEL_ORGANIZATION_ID`. The `AETHER_*` names are no
  longer read.
* Unchanged, because the platform still issues or serves them: the default
  endpoint `https://api.aetherplatform.cloud`, the `aether_` API token prefix,
  the `aether-<organization slug>` registry project, the `aether.cloud/` node
  label prefix and the `aether-block` / `aether-shared` storage classes.

UPGRADING FROM v0.1:

1. Change `source` to `norkeel/norkeel`, the provider block to
   `provider "norkeel"`, and every `aether_` type in your configuration to
   `norkeel_`.
2. Rename the environment variables.
3. Point existing state at the new provider and types: run
   `terraform state replace-provider plane-shift/aether norkeel/norkeel`, then
   for each resource `terraform state rm aether_<type>.<name>` and
   `terraform import norkeel_<type>.<name> <id>` (the import IDs are unchanged).
   `terraform plan` should then show no changes.


The Norkeel API now enforces organization roles on every request. No resource,
data source or attribute changes; the provider keeps calling the same routes.

ENHANCEMENTS:

* resource/norkeel_cluster: `capi_cluster_name` (also on `data.norkeel_cluster`) is described as the suffixed name the API gives clusters created after 2026-10, `<organization slug>-<name>-<6 random lowercase alphanumerics>`; earlier clusters keep `<organization slug>-<name>`. The API now refuses a new cluster whose organization slug plus name exceeds 22 characters with `422 validation_error` on `name`; existing clusters are unaffected and a rename is not re-checked. Descriptions only, no schema change.
* provider: an API refusal with `403 insufficient_role` is reported as
  `The API token's organization role does not allow this operation`, and
  `403 token_not_permitted` as `API tokens cannot perform this operation`. The
  detail carries the API message and the request id. Other `401`/`403` answers
  keep the existing `Norkeel API rejected the request` diagnostic.
* resource/norkeel_organization_member: the API's `409 last_owner` on a member
  removal is reported as `Cannot remove the organization's last owner`. The
  provider's own last-owner check still runs before any removal is sent; its
  diagnostic now uses the same summary.

NOTES:

* Only an organization owner can create API tokens (portal, Settings → API
  Tokens). An organization token acts with the `admin` role on everything the
  provider manages, may add members with any role (`owner` included) and may
  remove owners except the last one.
* resource/norkeel_organization_member: member emails are compared
  case-insensitively by the API; adding an address that differs from an
  existing member's only in letter case is refused with `409`. Emails must
  still be written in lower case in configuration.
* An API token is not tied to the member who created it. After removing an
  owner, revoke the API tokens they created.

## 0.1.0 (2026-10-12)

First public release on the Terraform Registry and the OpenTofu Registry as
`plane-shift/aether`.
