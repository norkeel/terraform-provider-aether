# Changelog

Release notes for the GitHub Release are written from this file; GoReleaser's
generated changelog is disabled (`.goreleaser.yml`).

## 0.2.0 (Unreleased)

Aether is now Norkeel, and the provider is renamed with it. This is the first
release under the new name; v0.1.0 stays available under the old address.

BREAKING CHANGES:

* provider: the address is now `norkeel/norkeel` and the provider block is
  `provider "norkeel"`. The old `plane-shift/aether` address gets no further
  releases.
* provider: every resource and data source is renamed from `aether_*` to
  `norkeel_*` (for example `aether_cluster` → `norkeel_cluster`).
* provider: the environment variables are now `NORKEEL_API_TOKEN`,
  `NORKEEL_ENDPOINT` and `NORKEEL_ORGANIZATION_ID`. The `AETHER_*` names are no
  longer read.
* provider: diagnostic summaries say `Norkeel API …` where they said
  `Aether API …`.
* Unchanged, because the platform still issues or serves them: the default
  endpoint `https://api.aetherplatform.cloud`, the `aether_` API token prefix,
  the `aether-<organization slug>` registry project, the `aether.cloud/` node
  label prefix and the `aether-block` / `aether-shared` storage classes.

UPGRADING FROM v0.1:

1. In your configuration, change `source` to `norkeel/norkeel`, the provider
   block to `provider "norkeel"`, every `aether_` type to `norkeel_`, and rename
   the environment variables. Run `terraform init`.
2. Add one `moved` block per resource (Terraform 1.8+), for example:

   ```hcl
   moved {
     from = aether_cluster.prod
     to   = norkeel_cluster.prod
   }
   ```

   Every resource accepts a move from its `aether_` type, keeping its state as
   is: robot account and webhook secrets are preserved and nothing is replaced.
   `terraform plan` should show only the moves.
3. On an engine without cross-type moves, remove the old addresses
   (`terraform state list | grep -E '(^|\.)aether_[a-z0-9_]+\.[A-Za-z0-9_-]+(\[[^]]*\])?$' | while IFS= read -r a; do terraform state rm "$a"; done`)
   and import each resource under its new type with an `import` block (same
   IDs). Imports do not restore `norkeel_registry_robot_account.secret` or
   `norkeel_webhook.secret`, so anything that reads them must be updated, and a
   `rotate_secret` already in the configuration will rotate the webhook's
   signing secret on the next apply.

Also in this release, carried over from the unreleased 0.1.1: the Norkeel API
now enforces organization roles on every request. No resource, data source or
attribute changes for that; the provider keeps calling the same routes.

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
