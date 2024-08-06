#!/bin/bash
if [ -e .env ] ; then
  set -a; source .env; set +a
fi
mkdocs build -d docs-build
AWS_PROFILE=in4it-vpn-server aws s3 sync --delete docs-build/ s3://in4it-vpn-documentation
AWS_PROFILE=in4it-vpn-server aws cloudfront create-invalidation --distribution-id ${CLOUDFRONT_DISTRIBUTION_ID} --paths "/*"
