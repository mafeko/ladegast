// Versioning is derived from Conventional Commits on main:
//   fix:  -> patch, feat: -> minor, "!" / BREAKING CHANGE -> major
// Everything else (chore, docs, ci, test, refactor, ...) does not release.
export default {
  branches: ["main"],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    ["@semantic-release/changelog", { changelogFile: "CHANGELOG.md" }],
    // Writes the version into package.json / package-lock.json (no npm publish).
    ["@semantic-release/npm", { npmPublish: false }],
    [
      "@semantic-release/exec",
      {
        // Keep the frontend manifest in sync with the release version.
        prepareCmd:
          "npm --prefix web version ${nextRelease.version} --no-git-tag-version --allow-same-version",
        // Hand the result to later CI jobs (docker build). No-op outside Actions.
        successCmd:
          'if [ -n "$GITHUB_OUTPUT" ]; then printf "new_release_published=true\\nnew_release_version=${nextRelease.version}\\n" >> "$GITHUB_OUTPUT"; fi',
      },
    ],
    [
      "@semantic-release/git",
      {
        assets: [
          "CHANGELOG.md",
          "package.json",
          "package-lock.json",
          "web/package.json",
          "web/package-lock.json",
        ],
        message:
          "chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}",
      },
    ],
    "@semantic-release/github",
  ],
};
