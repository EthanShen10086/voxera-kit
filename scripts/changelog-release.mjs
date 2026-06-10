#!/usr/bin/env node
/**
 * semver release：推断 bump → 固化 [Unreleased] → 打 tag vX.Y.Z
 */
import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { join } from 'node:path';
import {
  CHANGELOG,
  ROOT,
  assertCleanWorkingTree,
  assertReleaseBranch,
  buildUnreleasedBlock,
  bumpVersion,
  collectCommitSubjects,
  finalizeReleaseSection,
  getCurrentVersionFromTagOrChangelog,
  inferBumpType,
  parseLatestReleasedVersion,
  parseUnreleasedBlock,
  readChangelog,
  replaceUnreleasedSection,
  resolveLogBaseRef,
  todayStr,
  writeChangelog,
  writePackageVersion,
} from './changelog-lib.mjs';

function parseArgs() {
  const out = { dryRun: false, bump: null, releaseAs: null, skipTag: false };
  for (const a of process.argv.slice(2)) {
    if (a === '--dry-run') out.dryRun = true;
    else if (a === '--skip-tag') out.skipTag = true;
    else if (a.startsWith('--bump=')) out.bump = a.slice('--bump='.length);
    else if (a.startsWith('--release-as=')) out.releaseAs = a.slice('--release-as='.length);
  }
  return out;
}

function sh(cmd, args) {
  execFileSync(cmd, args, { cwd: ROOT, stdio: 'inherit' });
}

function buildNextContent(content, unreleasedBody, nextVersion, date) {
  const released = finalizeReleaseSection(unreleasedBody, nextVersion, date);
  const freshUnreleased = buildUnreleasedBlock([]);
  const unreleased = parseUnreleasedBlock(content);

  if (unreleased) {
    const lines = content.split('\n');
    const before = lines.slice(0, unreleased.start).join('\n').trimEnd();
    const after = lines.slice(unreleased.end).join('\n').replace(/^\n+/, '');
    return `${before}\n\n${released}\n\n${freshUnreleased}\n${after}`.trimEnd() + '\n';
  }

  const latest = parseLatestReleasedVersion(content);
  if (latest) {
    const marker = `## [${latest.version}]`;
    const idx = content.indexOf(marker);
    if (idx !== -1) {
      const before = content.slice(0, idx).trimEnd();
      const after = content.slice(idx);
      return `${before}\n\n${released}\n\n${freshUnreleased}\n${after}`.trimEnd() + '\n';
    }
  }

  return replaceUnreleasedSection(`${content.trimEnd()}\n\n${released}\n`, freshUnreleased);
}

function main() {
  const args = parseArgs();
  assertReleaseBranch();

  const content = readChangelog();
  const unreleased = parseUnreleasedBlock(content);
  const baseRef = resolveLogBaseRef(content);
  if (!baseRef) throw new Error('Cannot resolve git baseline for release');

  const commits = collectCommitSubjects(baseRef, 'HEAD');
  const bumpType = args.bump || inferBumpType(commits);
  if (!bumpType && !args.releaseAs) {
    if (process.env.RELEASE_IN_CI === '1') {
      console.log('[changelog-release] no releasable commits, skip');
      return;
    }
    throw new Error('No semver bump (only chore/docs/style since last release). Use --bump=patch or --release-as=X.Y.Z');
  }

  const current = getCurrentVersionFromTagOrChangelog(content);
  const nextVersion = args.releaseAs || bumpVersion(current, bumpType);
  const date = todayStr();

  let unreleasedBody = unreleased?.body?.trim();
  if (!unreleasedBody) {
    const bullets = commits.map(c => `- ${c.subject}`).join('\n');
    unreleasedBody = bullets
      ? `**变更摘要**（自动生成，发版前可删减无关项并改写）\n${bullets}`
      : '**变更摘要**\n- （无详细说明）';
  }

  const nextContent = buildNextContent(content, unreleasedBody, nextVersion, date);
  const tagName = `v${nextVersion}`;

  if (args.dryRun) {
    console.log('[changelog-release] dry-run');
    console.log(`  current: ${current}`);
    console.log(`  next:    ${nextVersion} (${args.bump || bumpType || 'release-as'})`);
    console.log(`  tag:     ${tagName}`);
    console.log(`  commits: ${commits.length}`);
    return;
  }

  assertCleanWorkingTree();

  writeChangelog(nextContent);
  const pkgPath = join(ROOT, 'package.json');
  const hasPkgVersion = existsSync(pkgPath);
  if (hasPkgVersion) writePackageVersion(nextVersion);

  sh('git', ['add', CHANGELOG]);
  if (hasPkgVersion) sh('git', ['add', 'package.json']);

  sh('git', ['commit', '-m', `chore(release): ${tagName} [release]`]);

  if (!args.skipTag) {
    sh('git', ['tag', '-a', tagName, '-m', `Release ${tagName}`]);
  }

  console.log(`[changelog-release] released ${tagName}`);
}

main();
