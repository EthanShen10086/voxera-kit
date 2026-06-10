#!/usr/bin/env node
/**
 * semver：维护 CHANGELOG.md 中 ## [Unreleased] 块（--refresh 用于 pre-commit）
 */
import {
  CHANGELOG,
  ROOT,
  buildUnreleasedBlock,
  collectSubjects,
  readChangelog,
  replaceUnreleasedSection,
  resolveLogBaseRef,
  writeChangelog,
} from './changelog-lib.mjs';

function parseArgs() {
  const out = { refresh: false, dryRun: false, mode: 'semver' };
  for (const a of process.argv.slice(2)) {
    if (a === '--refresh') out.refresh = true;
    else if (a === '--dry-run') out.dryRun = true;
    else if (a.startsWith('--mode=')) out.mode = a.slice('--mode='.length);
  }
  return out;
}

function main() {
  const args = parseArgs();
  if (process.env.CHANGELOG_SKIP === '1') {
    console.log('[changelog-append] skip CHANGELOG_SKIP=1');
    return;
  }
  if (process.env.CI === 'true' && process.env.CHANGELOG_IN_CI !== '1') {
    console.log('[changelog-append] skip CI (set CHANGELOG_IN_CI=1 to enable)');
    return;
  }

  let content = readChangelog();
  const baseRef = resolveLogBaseRef(content);
  if (!baseRef) {
    console.warn('[changelog-append] no git baseline, skip');
    return;
  }

  const subjects = collectSubjects(baseRef, 'HEAD');
  const block = buildUnreleasedBlock(subjects);
  const next = replaceUnreleasedSection(content, block);

  if (args.dryRun) {
    console.log(block);
    return;
  }

  if (next === content && !args.refresh) {
    console.log('[changelog-append] unchanged, skip');
    return;
  }

  writeChangelog(next);
  console.log(`[changelog-append] updated ${CHANGELOG} (base ${baseRef}, ${subjects.length} items)`);
}

main();
