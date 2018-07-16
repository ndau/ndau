#!/bin/bash
# This helper script returns the values used for circle ci

echo gomu="$(escape_newlines "$(cat $DIR/../gomu)")"
echo helm_ca_pem="$(escape_newlines "$(cat ~/.helm/ca.pem)")"
echo helm_cert_pem="$(escape_newlines "$(cat ~/.helm/cert.pem)")"
echo helm_key_pem="$(escape_newlines "$(cat ~/.helm/key.pem)")"
echo kube_config="$(escape_newlines "$(cat ~/.kube/dev-chaos.yaml)")"
