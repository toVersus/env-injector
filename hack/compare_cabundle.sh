#!/bin/bash

set -e

CA_BUNDLE=$(kubectl get mutatingwebhookconfiguration -o jsonpath='{.items[0].webhooks[0].clientConfig.caBundle}')
CA_CERTS=$(kubectl get secret -n injector env-injector-certs -o jsonpath='{.data.ca-cert\.pem}')

if [ "$CA_BUNDLE" = "$CA_CERTS" ]; then
  echo "CA certificate found in caBundle field of webhook configuration!"
else
  echo "Hmm, something is wrong with the webhook reconciliation :(
    got: ${CA_BUNDLE}
    want: ${CA_CERTS}"
fi
