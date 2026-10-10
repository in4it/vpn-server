#!/usr/bin/env bash
#
# Publish a new AMI version to an AWS Marketplace product.
#
# After "make install-aws" builds and shares a new AMI, this script automates
# what is otherwise the manual "Add version" step in the Marketplace UI:
#
#   1. Reads the previous data for the product from the most recent successful
#      "AddDeliveryOptions" change set (this is exactly the payload that was
#      submitted last time, including the access role, usage instructions,
#      security groups, recommended instance type, etc.).
#   2. Reuses that payload, swapping in the new AMI id and a new version title.
#   3. Prints the change it is about to make and asks for confirmation.
#   4. Submits it with marketplace-catalog start-change-set.
#
# Usage:
#   aws_marketplace_add_version.sh <product> [ami-id|latest] [version-title] [release-notes]
#
#   <product>        One of:
#                      vpn-server-boyl        (prod-hfbswxenloyaa)
#                      vpn-server             (prod-3l43xqobg7hni)
#                      vpn-server-boyl-arm64  (prod-mk6qahxtiwblc)
#                      all                    all of the above, with the newest
#                                             AMI of each product. All products
#                                             are checked first and confirmed
#                                             once; nothing is submitted if one
#                                             of them fails a check.
#                    The "byol" spelling is accepted as an alias.
#                    "make release-aws-marketplace" runs it for all products.
#   [ami-id|latest]  Optional. The new AMI to publish, e.g. ami-0123456789abcdef0.
#                    Defaults to "latest": the newest AMI built by
#                    "make install-aws" for the product (in4it-vpn-server-byol-*,
#                    in4it-vpn-server-licensed-* or in4it-vpn-server-arm64-byol-*,
#                    owned by this account in AMI_REGION). Stops when that AMI
#                    isn't available yet, or is already the published version.
#   [version-title]  Optional. Defaults to the contents of ./latest. Must be
#                    unique within the product.
#   [release-notes]  Optional. Defaults to the release notes of the GitHub
#                    release <version-title> (in4it/vpn-server), without the
#                    "What's Changed" list of pull requests. Falls back to
#                    "Version <version-title>" if the release can't be found.
#
# Environment:
#   AWS_REGION       Defaults to us-east-1
#   AMI_REGION       Region of the AMIs for the "latest" lookup. Defaults to us-east-1
#   ASSUME_YES=1     Skip the confirmation prompt (for automation).
#   DRY_RUN=1        Show the change, but don't submit it.
#   GITHUB_TOKEN     Optional. Used for the GitHub API request (rate limits).
#
set -euo pipefail

CATALOG="AWSMarketplace"
: "${AWS_REGION:=us-east-1}"
export AWS_REGION

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

die() { echo "error: $*" >&2; exit 1; }

command -v aws >/dev/null 2>&1 || die "aws cli not found in PATH"
command -v jq  >/dev/null 2>&1 || die "jq not found in PATH"
command -v curl >/dev/null 2>&1 || die "curl not found in PATH"

GITHUB_REPO="in4it/vpn-server"
MAX_RELEASE_NOTES_LENGTH=30000

# github_release_notes <tag>: prints the notes of the GitHub release as plain
# text: the part before "What's Changed", markdown links and code formatting
# removed. Prints nothing when the release can't be retrieved.
github_release_notes() {
  local tag="$1" auth=()
  [ -n "${GITHUB_TOKEN:-}" ] && auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
  curl -fsSL --max-time 15 "${auth[@]}" \
    -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${tag}" 2>/dev/null \
  | jq -r '.body // empty' \
  | tr -d '\r' \
  | awk '/^#+ *What.s Changed/ {exit} {print}' \
  | sed -E 's/\[([^]]+)\]\(([^)]+)\)/\1 (\2)/g; s/`//g; s/^\* /- /' \
  | sed -e :a -e '/^[[:space:]]*$/{$d;N;ba' -e '}'
}

# --- arguments ---------------------------------------------------------------
PRODUCT="${1:-}"
NEW_AMI="${2:-}"
NEW_TITLE="${3:-}"
NEW_NOTES="${4:-}"

usage() {
  sed -n '2,40p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  exit 1
}

[ -n "$PRODUCT" ] || usage
[ -n "$NEW_AMI" ] || NEW_AMI="latest"

# all products: check every product (dry run), confirm once, then submit
if [ "$PRODUCT" = "all" ]; then
  [ "$NEW_AMI" = "latest" ] || die "product 'all' always uses the newest AMI of each product, don't pass an AMI id"
  PRODUCTS=(vpn-server-boyl vpn-server vpn-server-boyl-arm64)
  for p in "${PRODUCTS[@]}"; do
    echo "==================== ${p} ====================" >&2
    DRY_RUN=1 "${BASH_SOURCE[0]}" "$p" latest "$NEW_TITLE" "$NEW_NOTES" \
      || die "check for ${p} failed, nothing was submitted"
  done
  if [ "${ASSUME_YES:-}" != "1" ]; then
    printf '\nSubmit the new version for all products (%s)? [y/N] ' "${PRODUCTS[*]}" >&2
    read -r reply
    case "$reply" in
      y|Y|yes|YES) ;;
      *) echo "Aborted." >&2; exit 1;;
    esac
  fi
  for p in "${PRODUCTS[@]}"; do
    echo "==================== submitting ${p} ====================" >&2
    ASSUME_YES=1 "${BASH_SOURCE[0]}" "$p" latest "$NEW_TITLE" "$NEW_NOTES" \
      || die "submitting ${p} failed; products listed before it were submitted"
  done
  exit 0
fi
: "${AMI_REGION:=us-east-1}"

# Resolve product name -> entity id and the name prefix of its AMIs (see provisioning/packer-aws.pkr.hcl).
case "$PRODUCT" in
  vpn-server-boyl|vpn-server-byol)             PRODUCT_ID="prod-hfbswxenloyaa"; AMI_NAME_PREFIX="in4it-vpn-server-byol-";;
  vpn-server)                                  PRODUCT_ID="prod-3l43xqobg7hni"; AMI_NAME_PREFIX="in4it-vpn-server-licensed-";;
  vpn-server-boyl-arm64|vpn-server-byol-arm64) PRODUCT_ID="prod-mk6qahxtiwblc"; AMI_NAME_PREFIX="in4it-vpn-server-arm64-byol-";;
  *) die "unknown product '$PRODUCT' (expected: vpn-server-boyl, vpn-server, vpn-server-boyl-arm64)";;
esac

# Find the newest AMI for the product. The newest one (by creation date) must be
# available: if it's still being created, wait instead of picking an older AMI.
NEW_AMI_NAME=""
if [ "$NEW_AMI" = "latest" ]; then
  echo "==> Looking up the newest ${AMI_NAME_PREFIX}* AMI in ${AMI_REGION} ..." >&2
  LATEST_IMAGE="$(
    aws ec2 describe-images \
      --region "$AMI_REGION" \
      --owners self \
      --filters "Name=name,Values=${AMI_NAME_PREFIX}*" \
      --output json \
    | jq -c '.Images | sort_by(.CreationDate) | last // empty | {id: .ImageId, name: .Name, state: .State, created: .CreationDate}'
  )"
  [ -n "$LATEST_IMAGE" ] || die "no ${AMI_NAME_PREFIX}* AMI found in ${AMI_REGION}; build it with make install-aws or pass an AMI id"
  NEW_AMI="$(jq -r '.id' <<<"$LATEST_IMAGE")"
  NEW_AMI_NAME="$(jq -r '"\(.name), created \(.created)"' <<<"$LATEST_IMAGE")"
  LATEST_STATE="$(jq -r '.state' <<<"$LATEST_IMAGE")"
  [ "$LATEST_STATE" = "available" ] || die "newest AMI ${NEW_AMI} (${NEW_AMI_NAME}) is ${LATEST_STATE}; wait until it is available"
fi

case "$NEW_AMI" in
  ami-*) ;;
  *) die "invalid AMI id '$NEW_AMI' (expected ami-... or latest)";;
esac

# Default version title from the repo's ./latest marker.
if [ -z "$NEW_TITLE" ]; then
  [ -f "${REPO_ROOT}/latest" ] || die "no version-title given and ${REPO_ROOT}/latest not found"
  NEW_TITLE="$(tr -d '[:space:]' < "${REPO_ROOT}/latest")"
fi
[ -n "$NEW_TITLE" ] || die "empty version title"
if [ -z "$NEW_NOTES" ]; then
  echo "==> Getting release notes from the GitHub release ${NEW_TITLE} ..." >&2
  NEW_NOTES="$(github_release_notes "$NEW_TITLE")"
  if [ -z "$NEW_NOTES" ]; then
    echo "warning: no GitHub release notes found for ${NEW_TITLE}, using \"Version ${NEW_TITLE}\"" >&2
    NEW_NOTES="Version ${NEW_TITLE}"
  fi
fi
if [ "${#NEW_NOTES}" -gt "$MAX_RELEASE_NOTES_LENGTH" ]; then
  NEW_NOTES="${NEW_NOTES:0:$((MAX_RELEASE_NOTES_LENGTH - 3))}..."
fi

echo "==> Looking up previous version data for ${PRODUCT} (${PRODUCT_ID}) ..." >&2

# --- find the most recent successful AddDeliveryOptions change set -----------
# Each change-set summary lists the entities it touched, so we can filter
# client-side without relying on server-side entity filters.
mapfile -t CHANGE_SET_IDS < <(
  aws marketplace-catalog list-change-sets \
    --catalog "$CATALOG" \
    --sort 'SortBy=StartTime,SortOrder=DESCENDING' \
    --output json \
  | jq -r --arg eid "$PRODUCT_ID" '
      .ChangeSetSummaryList[]
      | select(.Status == "SUCCEEDED")
      | select((.EntityIdList // []) | index($eid))
      | .ChangeSetId'
)

[ "${#CHANGE_SET_IDS[@]}" -gt 0 ] || die "no successful change sets found for ${PRODUCT_ID} (nothing to copy previous data from)"

TEMPLATE=""        # previous AddDeliveryOptions Details (write-view JSON)
ENTITY_TYPE=""     # e.g. AmiProduct@1.0, taken from the previous change
for csid in "${CHANGE_SET_IDS[@]}"; do
  match="$(
    aws marketplace-catalog describe-change-set \
      --catalog "$CATALOG" \
      --change-set-id "$csid" \
      --output json \
    | jq -c '
        [ .ChangeSet[]
          | select(.ChangeType == "AddDeliveryOptions")
          | { entityType: .Entity.Type,
              details: (.DetailsDocument // (.Details | fromjson)) }
          # only keep ones that actually carry an AMI delivery option
          | select(.details.DeliveryOptions[0].Details.AmiDeliveryOptionDetails != null)
        ] | .[0] // empty'
  )"
  if [ -n "$match" ]; then
    ENTITY_TYPE="$(jq -r '.entityType' <<<"$match")"
    TEMPLATE="$(jq -c '.details' <<<"$match")"
    break
  fi
done

[ -n "$TEMPLATE" ] || die "could not find a previous AddDeliveryOptions payload for ${PRODUCT_ID}"
[ -n "$ENTITY_TYPE" ] && [ "$ENTITY_TYPE" != "null" ] || ENTITY_TYPE="AmiProduct@1.0"

PREV_TITLE="$(jq -r '.Version.VersionTitle // "unknown"' <<<"$TEMPLATE")"
PREV_AMI="$(jq -r '.DeliveryOptions[0].Details.AmiDeliveryOptionDetails.AmiSource.AmiId // "unknown"' <<<"$TEMPLATE")"

if [ "$NEW_AMI" = "$PREV_AMI" ]; then
  die "AMI ${NEW_AMI} is already published as version ${PREV_TITLE}; build a new AMI with make install-aws first"
fi
if [ "$NEW_TITLE" = "$PREV_TITLE" ]; then
  die "version title '${NEW_TITLE}' is already the latest published version for this product; pass a distinct <version-title>"
fi

# --- build the new payload: reuse previous data, swap AMI + version ----------
NEW_DETAILS="$(
  jq -c \
    --arg ami "$NEW_AMI" \
    --arg title "$NEW_TITLE" \
    --arg notes "$NEW_NOTES" '
      del(.Version.Id)
      | .Version.VersionTitle = $title
      | .Version.ReleaseNotes = $notes
      | .DeliveryOptions = (
          .DeliveryOptions
          | map( del(.Id)
                 | .Details.AmiDeliveryOptionDetails.AmiSource.AmiId = $ami ) )
    ' <<<"$TEMPLATE"
)"

CHANGE_SET="$(
  jq -n \
    --arg type "$ENTITY_TYPE" \
    --arg id "$PRODUCT_ID" \
    --argjson details "$NEW_DETAILS" '
      [ { ChangeType: "AddDeliveryOptions",
          Entity: { Type: $type, Identifier: $id },
          DetailsDocument: $details } ]'
)"

CHANGE_SET_NAME="add-version-${PRODUCT}-${NEW_TITLE}"
CHANGE_SET_FILE="$(mktemp -t marketplace-changeset.XXXXXX.json)"
trap 'rm -f "$CHANGE_SET_FILE"' EXIT
printf '%s' "$CHANGE_SET" > "$CHANGE_SET_FILE"

# --- show the change and confirm ---------------------------------------------
cat >&2 <<EOF

About to add a new version to AWS Marketplace:

  Product:           ${PRODUCT} (${PRODUCT_ID})
  Entity type:       ${ENTITY_TYPE}
  Catalog:           ${CATALOG}  (region=${AWS_REGION})

  Previous version:  ${PREV_TITLE}   (AMI ${PREV_AMI})
  New version:       ${NEW_TITLE}   (AMI ${NEW_AMI}${NEW_AMI_NAME:+, ${NEW_AMI_NAME}})
  Release notes:
${NEW_NOTES}


Full start-change-set payload:
EOF
jq . "$CHANGE_SET_FILE" >&2

if [ "${DRY_RUN:-}" = "1" ]; then
  echo "==> Dry run: not submitted." >&2
  exit 0
fi

if [ "${ASSUME_YES:-}" != "1" ]; then
  printf '\nProceed with start-change-set? [y/N] ' >&2
  read -r reply
  case "$reply" in
    y|Y|yes|YES) ;;
    *) echo "Aborted." >&2; exit 1;;
  esac
fi

# --- submit ------------------------------------------------------------------
echo "==> Submitting change set ..." >&2
RESULT="$(
  aws marketplace-catalog start-change-set \
    --catalog "$CATALOG" \
    --change-set-name "$CHANGE_SET_NAME" \
    --change-set "file://${CHANGE_SET_FILE}" \
    --output json
)"

CS_ID="$(jq -r '.ChangeSetId // empty' <<<"$RESULT")"
echo "$RESULT"
echo "==> Submitted. ChangeSetId=${CS_ID}" >&2
echo "==> Track status with:" >&2
echo "      AWS_REGION=${AWS_REGION} aws marketplace-catalog describe-change-set --catalog ${CATALOG} --change-set-id ${CS_ID}" >&2
