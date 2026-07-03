#!/usr/bin/env bash
# Push track-app-go to GitHub. Create the repo first if it does not exist:
#   gh repo create prestonw/track-app-go --public --source=. --remote=origin --push
set -euo pipefail
cd "$(dirname "$0")/.."

if ! git remote get-url origin &>/dev/null; then
  git remote add origin git@github.com:prestonw/track-app-go.git
fi

if command -v gh &>/dev/null && gh auth status &>/dev/null; then
  if ! gh repo view prestonw/track-app-go &>/dev/null; then
    echo "Creating github.com/prestonw/track-app-go ..."
    gh repo create prestonw/track-app-go --public --source=. --remote=origin
  fi
fi

echo "Pushing main..."
git push -u origin main
echo "Done: https://github.com/prestonw/track-app-go"