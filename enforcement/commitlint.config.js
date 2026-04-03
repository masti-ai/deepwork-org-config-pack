// commitlint.config.js — Shared across all YourOrg repos
// Enforces: type(scope): description
// Types: feat, fix, refactor, chore, test, docs, perf, ci
// Scope: bead ID (pa-xxx) or area (dashboard, api, mobile)
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [2, 'always', [
      'feat', 'fix', 'refactor', 'chore', 'test', 'docs', 'perf', 'ci', 'build', 'revert'
    ]],
    'type-case': [2, 'always', 'lower-case'],
    'scope-empty': [1, 'never'],  // warn if no scope
    'subject-case': [2, 'always', 'lower-case'],
    'subject-full-stop': [2, 'never', '.'],
    'subject-max-length': [2, 'always', 72],
    'header-max-length': [2, 'always', 100],
    'body-max-line-length': [1, 'always', 100],  // warn only
  },
  // Allow merge commits and bd backup (from agents)
  ignores: [
    (commit) => commit.startsWith('Merge'),
    (commit) => commit.startsWith('bd:'),
    (commit) => commit.startsWith('auto:'),
  ],
};
