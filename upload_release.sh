#!/usr/bin/env bash
# This script was adapted from: https://codeberg.org/eduVPN/development/src/branch/main/development-setup-v3.md#codeberg
# Thanks fkooman again!

# exit on error
set -e

API_KEY_FILE="${HOME}/.config/codeberg.org/api.key"

if [ ! -f "$API_KEY_FILE" ]; then
    echo "You have to create a Codeberg API key and put it in $API_KEY_FILE, see: https://codeberg.org/eduVPN/development/src/branch/main/development-setup-v3.md#codeberg"
	exit 1
fi
ORG=eduVPN
API_KEY=$(cat "$API_KEY_FILE")
PROJECT_NAME=$(basename "$(pwd)")
PROJECT_VERSION=$(grep -o 'const Version = "[^"]*' internal/version/version.go | cut -d '"' -f 2)
PRERELEASE=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -p|--prerelease) PRERELEASE=true ;;
		-v|--version) PROJECT_VERSION="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

CHANGES_TRIM=$(sed "/^# $PROJECT_VERSION/,/^#/!d;//d" "CHANGES.md")
if [ "$PRERELEASE" = true ]; then
    CHANGES=$(printf "These pre-releases are signed with \`keys/app+linux+dev@eduvpn.org.asc\` and \`keys/minisign-CA9409316AC93C07.pub\`\nChangelog:\n%s" "${CHANGES_TRIM}")
else
    CHANGES=$(printf "These releases are signed with \`keys/app+linux@eduvpn.org.asc\` and \`keys/minisign-CA9409316AC93C07.pub\`\nChangelog:\n%s" "${CHANGES_TRIM}")
fi

if ! command -v "curl" &>/dev/null; then
    echo "please install curl for contacting the Codeberg API"
    exit 1
fi

if ! command -v "jq" &>/dev/null; then
    echo "please install jq for parsing JSON"
    exit 1
fi

JSON_BODY="{\"body\": \"${CHANGES}\", \"tag_name\": \"${PROJECT_VERSION}\", \"prerelease\": ${PRERELEASE}}"

# create the release
RELEASE_ID=$(curl -s \
    -H "Authorization: token ${API_KEY}" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -d "${JSON_BODY}" \
    "https://codeberg.org/api/v1/repos/${ORG}/${PROJECT_NAME}/releases" | jq -r .id)

# upload the artifact(s)
for F in release/*"${PROJECT_VERSION}"*; do
    curl \
        -s \
        -X "POST" \
        -H "Authorization: token ${API_KEY}" \
        -H "Accept: application/json" \
        -H "Content-Type: multipart/form-data" \
        -F "attachment=@${F}" \
        "https://codeberg.org/api/v1/repos/${ORG}/${PROJECT_NAME}/releases/${RELEASE_ID}/assets"
done
