if [[ "$TRAVIS_PULL_REQUEST" == "false" ]]; then
  exit 0
fi

if ! git diff --name-only HEAD.."${TRAVIS_BRANCH}" | grep CHANGELOG.md; then
  echo "[FAIL] No diff between HEAD and ${TRAVIS_BRANCH}"
  echo "[FAIL] PR has no CHANGELOG entry. Please update the CHANGELOG!"
  exit 1
fi

exit 0
