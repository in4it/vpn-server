#!/usr/bin/env bash
#
# Publish a new VM image version to the Azure Marketplace offer.
#
# After "make install-azure" and "make install-azure-arm64" built the new image
# versions in the Azure Compute Gallery, this script automates the "add a VM
# image" step in Partner Center, using the Product Ingestion API:
#
#   1. Finds the VM offers of the account (or the offers in AZURE_OFFER_ID), and
#      reads the technical configuration of every plan (draft). Only plans whose
#      newest image comes from the vpn-server gallery images (GALLERY_IMAGE_MATCH)
#      are updated, other plans and offers are skipped.
#   2. For each plan, takes the gallery image of its newest image version and
#      adds the new version of that gallery image (e.g. .../versions/1.1.20 ->
#      .../versions/1.1.21), so amd64 and arm64 plans keep their own image.
#   3. Checks that the new gallery image versions exist (az sig image-version).
#   4. Prints the changes and asks for confirmation (once, for all offers).
#   5. Submits the changes, publishes every changed offer to preview, and waits
#      for the results.
#
# After checking the offers in preview, run it again with --live to publish the
# preview submissions live.
#
# Usage:
#   azure_marketplace_add_version.sh [version] [--live]
#
#   [version]   Optional. Defaults to the contents of ./latest, e.g. v1.1.21.
#   --live      Publish the current preview submission of the offers live (no
#               other changes).
#
# Configuration (environment, or provisioning/azure.env which is gitignored;
# see provisioning/azure.env.example):
#   AZURE_PARTNER_TENANT_ID      Tenant of the Microsoft Entra app that is added to
#   AZURE_PARTNER_CLIENT_ID      Partner Center (Account settings > Users) with the
#   AZURE_PARTNER_CLIENT_SECRET  Manager role, and its client id and key.
#   AZURE_OFFER_ID               Optional. Offer ID(s) to update, comma separated.
#                                Defaults to all VM offers of the account.
#   GALLERY_IMAGE_MATCH          Optional. Only plans with an image whose resource id
#                                contains this are updated. Defaults to
#                                /galleries/vpnserver/images/in4it-vpn-server
#   ASSUME_YES=1                 Skip the confirmation prompt.
#   DRY_RUN=1                    Show the change, but don't submit it.
#
# The gallery check uses the az cli (az login).
#
set -euo pipefail

API="https://graph.microsoft.com/rp/product-ingestion"
SCHEMA_VERSION="2022-03-01-preview3"
TECH_CONFIG_SCHEMA="https://schema.mp.microsoft.com/schema/virtual-machine-plan-technical-configuration"
CONFIGURE_SCHEMA="https://schema.mp.microsoft.com/schema/configure/2022-03-01-preview2"
SUBMISSION_SCHEMA="https://schema.mp.microsoft.com/schema/submission/2022-03-01-preview2"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVISIONING_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${PROVISIONING_DIR}/.." && pwd)"

die() { echo "error: $*" >&2; exit 1; }

usage() {
  sed -n '2,36p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  exit 1
}

# add_image_version <technical-configuration-json> <version-number>
# Prints the technical configuration with the new image version added (the
# gallery image of the newest existing version, with the version replaced), or
# nothing when the version is already there. Fails when there's no gallery image to base it on.
add_image_version() {
  jq -c --arg v "$2" '
    if any(.vmImageVersions[]?; .versionNumber == $v) then empty
    else
      (.vmImageVersions // [] | map(select(.vmImages[0].source.sourceType == "sharedImageGallery"))
        | sort_by(.versionNumber | split(".") | map(tonumber)) | last) as $prev
      | if $prev == null then error("no gallery image version to base the new version on") else . end
      | .vmImageVersions += [{
          versionNumber: $v,
          vmImages: [$prev.vmImages[] | .source.sharedImage.resourceId |= sub("/versions/[^/]+$"; "/versions/" + $v)]
        }]
    end' <<<"$1"
}

# shellcheck disable=SC2317
if [ "${1:-}" = "--self-test-add-image-version" ]; then
  add_image_version "$2" "$3"
  exit 0
fi

LIVE=0
VERSION=""
for arg in "$@"; do
  case "$arg" in
    --live) LIVE=1;;
    -h|--help) usage;;
    -*) die "unknown option $arg";;
    *) VERSION="$arg";;
  esac
done

if [ -f "${PROVISIONING_DIR}/azure.env" ]; then
  # shellcheck disable=SC1091
  set -a; . "${PROVISIONING_DIR}/azure.env"; set +a
fi
: "${AZURE_PARTNER_TENANT_ID:?set AZURE_PARTNER_TENANT_ID (environment or provisioning/azure.env)}"
: "${AZURE_PARTNER_CLIENT_ID:?set AZURE_PARTNER_CLIENT_ID (environment or provisioning/azure.env)}"
: "${AZURE_PARTNER_CLIENT_SECRET:?set AZURE_PARTNER_CLIENT_SECRET (environment or provisioning/azure.env)}"
: "${AZURE_OFFER_ID:=}"
: "${GALLERY_IMAGE_MATCH:=/galleries/vpnserver/images/in4it-vpn-server}"

for cmd in curl jq az; do
  command -v "$cmd" >/dev/null 2>&1 || die "$cmd not found in PATH"
done

if [ -z "$VERSION" ]; then
  [ -f "${REPO_ROOT}/latest" ] || die "no version given and ${REPO_ROOT}/latest not found"
  VERSION="$(tr -d '[:space:]' < "${REPO_ROOT}/latest")"
fi
VERSION_NUMBER="${VERSION#v}"
[[ "$VERSION_NUMBER" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "invalid version '${VERSION}' (expected e.g. v1.1.21)"

confirm() {
  [ "${ASSUME_YES:-}" = "1" ] && return 0
  printf '\n%s [y/N] ' "$1" >&2
  read -r reply
  case "$reply" in
    y|Y|yes|YES) ;;
    *) echo "Aborted." >&2; exit 1;;
  esac
}

# --- Partner Center API -------------------------------------------------------
echo "==> Getting a Partner Center access token ..." >&2
TOKEN="$(
  curl -fsS --max-time 30 "https://login.microsoftonline.com/${AZURE_PARTNER_TENANT_ID}/oauth2/v2.0/token" \
    --data-urlencode "grant_type=client_credentials" \
    --data-urlencode "client_id=${AZURE_PARTNER_CLIENT_ID}" \
    --data-urlencode "client_secret=${AZURE_PARTNER_CLIENT_SECRET}" \
    --data-urlencode "scope=https://graph.microsoft.com/.default" \
  | jq -r '.access_token // empty'
)" || die "could not get an access token (check AZURE_PARTNER_* settings)"
[ -n "$TOKEN" ] || die "could not get an access token (check AZURE_PARTNER_* settings)"

api() { # api <method> <path> [body]
  local method="$1" path="$2" body="${3:-}"
  if [ -n "$body" ]; then
    curl -fsS --max-time 60 -X "$method" -H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json" "${API}/${path}" --data "$body"
  else
    curl -fsS --max-time 60 -X "$method" -H "Authorization: Bearer ${TOKEN}" "${API}/${path}"
  fi
}

# --- offers --------------------------------------------------------------------
# list_offers prints "<product id> <offer id>" for the VM offers to update
list_offers() {
  local page next="product?type=azureVirtualMachine&\$version=${SCHEMA_VERSION}"
  while [ -n "$next" ]; do
    page="$(api GET "$next")" || die "could not list the VM offers"
    jq -r '.value[] | "\(.id) \(.identity.externalId // .alias // .id)"' <<<"$page"
    next="$(jq -r '.continuationToken // empty' <<<"$page")"
    [ -n "$next" ] && next="product?type=azureVirtualMachine&continuationToken=${next}&\$version=${SCHEMA_VERSION}"
  done
}

OFFERS=()
while read -r product_id offer_id; do
  [ -n "$product_id" ] || continue
  if [ -n "$AZURE_OFFER_ID" ] && [[ ",${AZURE_OFFER_ID// /}," != *",${offer_id},"* ]]; then
    continue
  fi
  OFFERS+=("${product_id} ${offer_id}")
done < <(list_offers)
[ "${#OFFERS[@]}" -gt 0 ] || die "no VM offers found${AZURE_OFFER_ID:+ matching ${AZURE_OFFER_ID}}"

# wait_for_job <configure-status-json> <offer-id>
wait_for_job() {
  local job_id status
  job_id="$(jq -r '.jobId // empty' <<<"$1")"
  [ -n "$job_id" ] || die "$2: unexpected configure response: $1"
  echo "==> $2: job ${job_id} submitted, waiting for the result ..." >&2
  for _ in $(seq 1 120); do
    status="$(api GET "configure/${job_id}/status?\$version=${SCHEMA_VERSION}")"
    if [ "$(jq -r '.jobStatus' <<<"$status")" = "completed" ]; then
      if [ "$(jq -r '.jobResult' <<<"$status")" = "succeeded" ]; then
        echo "==> $2: succeeded." >&2
        return 0
      fi
      jq '.errors' <<<"$status" >&2
      die "$2: job ${job_id} failed"
    fi
    sleep 10
  done
  die "$2: job ${job_id} is still running, check Partner Center"
}

submission_payload() { # submission_payload <product-id> <targetType> [resources-json-array]
  jq -n --arg schema "$CONFIGURE_SCHEMA" --arg subschema "$SUBMISSION_SCHEMA" --arg product "$1" \
    --arg target "$2" --argjson resources "${3:-[]}" '
    {"$schema": $schema, resources: ($resources + [{"$schema": $subschema, product: $product, target: {targetType: $target}}])}'
}

if [ "$LIVE" = "1" ]; then
  echo >&2
  echo "About to publish the current preview submission of these offers live:" >&2
  for offer in "${OFFERS[@]}"; do echo "  ${offer#* } (${offer%% *})" >&2; done
  echo "Make sure you checked the offers in preview first." >&2
  [ "${DRY_RUN:-}" = "1" ] && { echo "==> Dry run: not submitted." >&2; exit 0; }
  confirm "Publish live?"
  for offer in "${OFFERS[@]}"; do
    wait_for_job "$(api POST "configure?\$version=2022-03-01-preview2" "$(submission_payload "${offer%% *}" live)")" "${offer#* }"
  done
  exit 0
fi

# --- add the new image version to the plans -----------------------------------
PAYLOADS=()     # "<offer id>\t<payload>" per offer with changes
SUMMARY=""
for offer in "${OFFERS[@]}"; do
  product_id="${offer%% *}"; offer_id="${offer#* }"
  echo "==> Reading the technical configuration of offer ${offer_id} ..." >&2
  tech_configs="$(api GET "resource-tree/${product_id}?\$version=${SCHEMA_VERSION}" \
    | jq -c --arg s "$TECH_CONFIG_SCHEMA" '[.resources[] | select(."$schema" | startswith($s))]')" \
    || die "could not read offer ${offer_id}"
  SUMMARY+="  offer ${offer_id}:"$'\n'
  updated_configs="[]"
  while read -r tech_config; do
    [ -n "$tech_config" ] || continue
    plan="$(jq -r '.plan' <<<"$tech_config")"
    if ! jq -e --arg m "$GALLERY_IMAGE_MATCH" 'any(.vmImageVersions[]?.vmImages[]?; (.source.sharedImage.resourceId // "") | contains($m))' <<<"$tech_config" >/dev/null; then
      SUMMARY+="    ${plan}: no ${GALLERY_IMAGE_MATCH} image, skipped"$'\n'
      continue
    fi
    updated="$(add_image_version "$tech_config" "$VERSION_NUMBER")" || die "${offer_id} ${plan}: could not add version ${VERSION_NUMBER}"
    if [ -z "$updated" ]; then
      SUMMARY+="    ${plan}: already has ${VERSION_NUMBER}, skipped"$'\n'
      continue
    fi
    while read -r resource_id; do
      state="$(az sig image-version show --ids "$resource_id" --query provisioningState -o tsv 2>/dev/null)" \
        || die "${offer_id} ${plan}: gallery image version ${resource_id} not found (build it with make install-azure / install-azure-arm64, and run az login)"
      [ "$state" = "Succeeded" ] || die "${offer_id} ${plan}: gallery image version ${resource_id} is ${state}"
      SUMMARY+="    ${plan}: add ${VERSION_NUMBER} (${resource_id##*/galleries/})"$'\n'
    done < <(jq -r --arg v "$VERSION_NUMBER" '.vmImageVersions[] | select(.versionNumber == $v) | .vmImages[].source.sharedImage.resourceId' <<<"$updated")
    updated_configs="$(jq -c --argjson r "$updated" '. + [$r]' <<<"$updated_configs")"
  done < <(jq -c '.[]' <<<"$tech_configs")
  if [ "$(jq length <<<"$updated_configs")" -gt 0 ]; then
    PAYLOADS+=("${offer_id}"$'\t'"$(submission_payload "$product_id" preview "$updated_configs")")
  fi
done

cat >&2 <<EOF

About to add version ${VERSION_NUMBER} to Azure Marketplace and publish the changed offers to preview:
${SUMMARY}
EOF
[ "${#PAYLOADS[@]}" -gt 0 ] || { echo "Nothing to do." >&2; exit 0; }

if [ "${DRY_RUN:-}" = "1" ]; then
  for p in "${PAYLOADS[@]}"; do echo "--- offer ${p%%$'\t'*}" >&2; jq . <<<"${p#*$'\t'}" >&2; done
  echo "==> Dry run: not submitted." >&2
  exit 0
fi
confirm "Submit and publish to preview?"
for p in "${PAYLOADS[@]}"; do
  wait_for_job "$(api POST "configure?\$version=2022-03-01-preview2" "${p#*$'\t'}")" "${p%%$'\t'*}"
done
cat >&2 <<EOF

Version ${VERSION_NUMBER} is being published to preview. Once the previews are
ready in Partner Center and checked, publish them live with:
  make release-azure-marketplace-live
EOF
