#!/usr/bin/env bash
#
# Build and publish a new VM image for the Google Cloud Marketplace listing.
#
# Runs locally and does everything up to the Producer Portal step:
#
#   1. Builds the webapp and the linux/amd64 binaries (make build-webapp build-amd64).
#   2. Builds the image with packer (provisioning/packer-gcp-amd64.pkr.hcl) in the
#      public project, with the Marketplace license attached.
#   3. Verifies the license is attached to the new image.
#   4. Makes the image public (roles/compute.imageUser for allAuthenticatedUsers),
#      which Marketplace requires.
#   5. Prints the image to select in the Producer Portal.
#
# Adding the image as a new version in the Producer Portal (Deployment package)
# and submitting it for review is still a manual step: the Producer Portal has
# no API for VM image versions.
#
# Usage:
#   gcp_marketplace_release.sh [version]
#
#   [version]  Optional. Defaults to the contents of ./latest, e.g. v1.1.20.
#
# Configuration (environment, or provisioning/gcp.env which is gitignored;
# see provisioning/gcp.env.example):
#   GCP_PROJECT_ID           Public project the image is created in.
#   GCP_MARKETPLACE_LICENSE  License from the Producer Portal deployment package,
#                            projects/<project>/global/licenses/cloud-marketplace-<id>
#   GCP_ZONE                 Optional. Zone for the packer build VM (default us-east4-c).
#   SKIP_BUILD=1             Skip building the webapp and binaries.
#   ASSUME_YES=1             Skip the confirmation prompts.
#   DRY_RUN=1                Don't build anything: check the configuration, gcloud
#                            access, project, zone, license and packer template,
#                            and show the image that would be built.
#
# Authentication: packer uses Application Default Credentials, so run
#   gcloud auth login && gcloud auth application-default login
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVISIONING_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${PROVISIONING_DIR}/.." && pwd)"
MANIFEST="${PROVISIONING_DIR}/packer-gcp-manifest.json"

die() { echo "error: $*" >&2; exit 1; }

usage() {
  sed -n '2,36p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  exit 1
}

confirm() {
  [ "${ASSUME_YES:-}" = "1" ] && return 0
  printf '\n%s [y/N] ' "$1" >&2
  read -r reply
  case "$reply" in
    y|Y|yes|YES) ;;
    *) echo "Aborted." >&2; exit 1;;
  esac
}

case "${1:-}" in
  -h|--help) usage;;
esac

# --- configuration -----------------------------------------------------------
if [ -f "${PROVISIONING_DIR}/gcp.env" ]; then
  # shellcheck disable=SC1091
  set -a; . "${PROVISIONING_DIR}/gcp.env"; set +a
fi

: "${GCP_PROJECT_ID:?set GCP_PROJECT_ID (environment or provisioning/gcp.env)}"
: "${GCP_MARKETPLACE_LICENSE:?set GCP_MARKETPLACE_LICENSE (environment or provisioning/gcp.env)}"
: "${GCP_ZONE:=us-east4-c}"
export GCP_PROJECT_ID GCP_MARKETPLACE_LICENSE

case "$GCP_MARKETPLACE_LICENSE" in
  projects/*/global/licenses/*) ;;
  *) die "GCP_MARKETPLACE_LICENSE must look like projects/<project>/global/licenses/<name>";;
esac

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  [ -f "${REPO_ROOT}/latest" ] || die "no version given and ${REPO_ROOT}/latest not found"
  VERSION="$(tr -d '[:space:]' < "${REPO_ROOT}/latest")"
fi
[ -n "$VERSION" ] || die "empty version"

for cmd in gcloud packer jq; do
  command -v "$cmd" >/dev/null 2>&1 || die "$cmd not found in PATH"
done
if [ "${SKIP_BUILD:-}" != "1" ]; then
  for cmd in make go npm; do
    command -v "$cmd" >/dev/null 2>&1 || die "$cmd not found in PATH"
  done
fi

if [ "${DRY_RUN:-}" = "1" ]; then
  FAILED=0
  check() { # check <description> <command...>
    local desc="$1"; shift
    local out
    if out="$("$@" 2>&1)"; then
      echo "  ok    ${desc}" >&2
    else
      echo "  FAIL  ${desc}" >&2
      [ -n "$out" ] && sed 's/^/          /' <<<"$out" | tail -5 >&2
      FAILED=1
    fi
  }
  LICENSE_PROJECT="$(cut -d/ -f2 <<<"$GCP_MARKETPLACE_LICENSE")"
  LICENSE_NAME="${GCP_MARKETPLACE_LICENSE##*/}"
  IMAGE_VERSION="$(tr '[:upper:].' '[:lower:]-' <<<"$VERSION")"
  echo "==> Dry run: checking the configuration (nothing is built) ..." >&2
  check "gcloud is logged in ($(gcloud auth list --filter=status:ACTIVE --format='value(account)' 2>/dev/null | head -1))" \
    sh -c 'gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .'
  check "application default credentials (used by packer)" gcloud auth application-default print-access-token
  check "project ${GCP_PROJECT_ID} exists" gcloud projects describe "$GCP_PROJECT_ID" --format="value(projectId)"
  check "Compute Engine API is enabled in ${GCP_PROJECT_ID}" \
    sh -c "gcloud services list --enabled --project '$GCP_PROJECT_ID' --filter='config.name:compute.googleapis.com' --format='value(config.name)' | grep -q compute"
  check "zone ${GCP_ZONE} exists" gcloud compute zones describe "$GCP_ZONE" --project "$GCP_PROJECT_ID" --format="value(name)"
  # gcloud has no licenses command, use the Compute API (licenses.get)
  check "license ${LICENSE_NAME} exists in ${LICENSE_PROJECT}" \
    sh -c "curl -fsS --max-time 30 -H \"Authorization: Bearer \$(gcloud auth print-access-token)\" \
      'https://compute.googleapis.com/compute/v1/projects/${LICENSE_PROJECT}/global/licenses/${LICENSE_NAME}' >/dev/null"
  check "packer plugins (packer init)" sh -c "cd '$PROVISIONING_DIR' && packer init packer-gcp-amd64.pkr.hcl"
  if [ -f "${REPO_ROOT}/restserver-linux-amd64" ] && [ -f "${REPO_ROOT}/configmanager-linux-amd64" ] && [ -f "${REPO_ROOT}/reset-admin-password-linux-amd64" ]; then
    check "packer template is valid" sh -c "cd '$PROVISIONING_DIR' && packer validate -var 'image_version=${VERSION}' -var 'zone=${GCP_ZONE}' packer-gcp-amd64.pkr.hcl"
  else
    check "packer template syntax (binaries not built yet, full check runs during the release)" \
      sh -c "cd '$PROVISIONING_DIR' && packer validate -syntax-only packer-gcp-amd64.pkr.hcl"
  fi
  cat >&2 <<EOF

Would build:
  Image:     in4it-vpn-server-ubuntu2404-x86-64-${IMAGE_VERSION}-<timestamp> (family in4it-vpn-server)
  Project:   ${GCP_PROJECT_ID}, build VM in ${GCP_ZONE}
  License:   ${GCP_MARKETPLACE_LICENSE}
  Binaries:  $([ "${SKIP_BUILD:-}" = "1" ] && echo "existing (SKIP_BUILD=1)" || echo "built with make build-webapp build-amd64")
and then make the image public (roles/compute.imageUser for allAuthenticatedUsers).
EOF
  [ "$FAILED" = "0" ] || die "dry run: some checks failed"
  echo "==> Dry run: all checks passed, nothing was built." >&2
  exit 0
fi

gcloud auth application-default print-access-token >/dev/null 2>&1 \
  || die "no application default credentials, run: gcloud auth application-default login"

cat >&2 <<EOF

About to build a Google Cloud Marketplace image:

  Version:   ${VERSION}
  Project:   ${GCP_PROJECT_ID}
  Zone:      ${GCP_ZONE}
  License:   ${GCP_MARKETPLACE_LICENSE}
  Build:     $([ "${SKIP_BUILD:-}" = "1" ] && echo "skipped (SKIP_BUILD=1)" || echo "make build-webapp build-amd64")
EOF
confirm "Proceed?"

# --- build binaries ----------------------------------------------------------
if [ "${SKIP_BUILD:-}" != "1" ]; then
  echo "==> Building webapp and linux/amd64 binaries ..." >&2
  make -C "$REPO_ROOT" build-webapp build-amd64
fi
for bin in configmanager-linux-amd64 restserver-linux-amd64 reset-admin-password-linux-amd64; do
  [ -f "${REPO_ROOT}/${bin}" ] || die "${bin} not found, build it first (or unset SKIP_BUILD)"
done

# --- build image -------------------------------------------------------------
echo "==> Building image with packer ..." >&2
rm -f "$MANIFEST"
(
  cd "$PROVISIONING_DIR"
  packer init packer-gcp-amd64.pkr.hcl
  packer build \
    -var "image_version=${VERSION}" \
    -var "zone=${GCP_ZONE}" \
    packer-gcp-amd64.pkr.hcl
)

[ -f "$MANIFEST" ] || die "packer manifest ${MANIFEST} not found"
IMAGE_NAME="$(jq -r '.builds[-1].artifact_id // empty' "$MANIFEST")"
[ -n "$IMAGE_NAME" ] || die "could not read image name from ${MANIFEST}"
echo "==> Built image ${IMAGE_NAME}" >&2

# --- verify license ----------------------------------------------------------
LICENSES="$(gcloud compute images describe "$IMAGE_NAME" \
  --project "$GCP_PROJECT_ID" --format=json | jq -r '.licenses[]? // empty')"
LICENSE_NAME="${GCP_MARKETPLACE_LICENSE##*/}"
grep -q "/licenses/${LICENSE_NAME}\$" <<<"$LICENSES" \
  || die "license ${GCP_MARKETPLACE_LICENSE} is not attached to ${IMAGE_NAME} (found: ${LICENSES:-none})"
echo "==> License attached" >&2

# --- make image public -------------------------------------------------------
confirm "Make ${IMAGE_NAME} public (roles/compute.imageUser for allAuthenticatedUsers)?"
gcloud compute images add-iam-policy-binding "$IMAGE_NAME" \
  --project "$GCP_PROJECT_ID" \
  --member=allAuthenticatedUsers \
  --role=roles/compute.imageUser >/dev/null
echo "==> Image is public" >&2

cat >&2 <<EOF

Image ready: projects/${GCP_PROJECT_ID}/global/images/${IMAGE_NAME}

Next steps in the Producer Portal (https://console.cloud.google.com/producer-portal):
  1. Open the vpn-server product > Deployment package.
  2. Add ${IMAGE_NAME} as a new image version (release notes: ${VERSION}) and
     set it as the default image.
  3. Save and submit the product for review.
EOF
