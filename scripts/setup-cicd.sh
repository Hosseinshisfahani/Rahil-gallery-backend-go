#!/usr/bin/env bash
# One-time CI/CD bootstrap: deploy SSH key + GitHub Actions secrets for both repos.
#
# Usage:
#   export GITHUB_TOKEN="$(github_personal_token)"   # or a PAT with repo scope
#   bash scripts/setup-cicd.sh \
#     --host 46.249.101.208 \
#     --user root \
#     --server-path /root/source/Rahil-Gallery-Server \
#     --client-path /root/source/Rahil-Gallery-Client
#
# Optional: install the public key on the VPS automatically:
#   bash scripts/setup-cicd.sh ... --install-key-on-server
set -euo pipefail

HOST=""
USER="root"
SERVER_PATH="/root/source/Rahil-Gallery-Server"
CLIENT_PATH="/root/source/Rahil-Gallery-Client"
INSTALL_KEY=false
GITHUB_OWNER="${GITHUB_OWNER:-aliakbarebrahimy}"
KEY_PATH="${DEPLOY_KEY_PATH:-$HOME/.ssh/rahil_github_deploy}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) HOST="$2"; shift 2 ;;
    --user) USER="$2"; shift 2 ;;
    --server-path) SERVER_PATH="$2"; shift 2 ;;
    --client-path) CLIENT_PATH="$2"; shift 2 ;;
    --install-key-on-server) INSTALL_KEY=true; shift ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "${HOST}" ]]; then
  echo "ERROR: --host is required" >&2
  exit 1
fi

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  if command -v github_personal_token >/dev/null 2>&1; then
    GITHUB_TOKEN="$(github_personal_token)"
  else
    echo "ERROR: set GITHUB_TOKEN or install github_personal_token helper" >&2
    exit 1
  fi
fi

if [[ ! -f "${KEY_PATH}" ]]; then
  echo "==> Generating deploy key at ${KEY_PATH}"
  ssh-keygen -t ed25519 -f "${KEY_PATH}" -N "" -C "github-actions-rahil-gallery"
fi

if [[ "${INSTALL_KEY}" == true ]]; then
  echo "==> Installing public key on ${USER}@${HOST}"
  ssh-copy-id -i "${KEY_PATH}.pub" "${USER}@${HOST}"
else
  echo "==> Add this public key to ${USER}@${HOST} (~/.ssh/authorized_keys):"
  echo
  cat "${KEY_PATH}.pub"
  echo
fi

set_github_secret() {
  local repo="$1"
  local name="$2"
  local value="$3"

  python3 - "${GITHUB_OWNER}" "${repo}" "${name}" "${value}" "${GITHUB_TOKEN}" <<'PY'
import base64, json, sys, urllib.request
owner, repo, name, value, token = sys.argv[1:6]
headers = {
    "Authorization": f"Bearer {token}",
    "Accept": "application/vnd.github+json",
    "X-GitHub-Api-Version": "2022-11-28",
}
pk_req = urllib.request.Request(
    f"https://api.github.com/repos/{owner}/{repo}/actions/secrets/public-key",
    headers=headers,
)
with urllib.request.urlopen(pk_req) as resp:
    pk = json.load(resp)

try:
    from nacl import encoding, public
except ImportError:
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "-q", "pynacl"])
    from nacl import encoding, public

public_key = public.PublicKey(pk["key"].encode("utf-8"), encoding.Base64Encoder())
sealed_box = public.SealedBox(public_key)
encrypted = sealed_box.encrypt(value.encode("utf-8"))
payload = json.dumps({
    "encrypted_value": base64.b64encode(encrypted).decode("utf-8"),
    "key_id": pk["key_id"],
}).encode("utf-8")

put_req = urllib.request.Request(
    f"https://api.github.com/repos/{owner}/{repo}/actions/secrets/{name}",
    data=payload,
    headers={**headers, "Content-Type": "application/json"},
    method="PUT",
)
with urllib.request.urlopen(put_req) as resp:
    resp.read()
print(f"  set {name} on {owner}/{repo}")
PY
}

PRIVATE_KEY="$(cat "${KEY_PATH}")"

for REPO in Rahil-Gallery-Server Rahil-Gallery-Client; do
  echo "==> Configuring secrets on ${GITHUB_OWNER}/${REPO}"
  set_github_secret "${REPO}" "SSH_PRIVATE_KEY" "${PRIVATE_KEY}"
  set_github_secret "${REPO}" "SSH_HOST" "${HOST}"
  set_github_secret "${REPO}" "SSH_USER" "${USER}"
done

set_github_secret "Rahil-Gallery-Server" "SERVER_DEPLOY_PATH" "${SERVER_PATH}"
set_github_secret "Rahil-Gallery-Client" "CLIENT_DEPLOY_PATH" "${CLIENT_PATH}"

echo
echo "Done. CI/CD secrets are configured."
echo "Next on VPS (once): gunzip -c rahil-docker-images.tar.gz | docker load"
echo "Then push to master — GitHub Actions will deploy automatically."
