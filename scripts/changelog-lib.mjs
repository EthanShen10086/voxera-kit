import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

export const scriptDir = dirname(fileURLToPath(import.meta.url));
export const ROOT = join(scriptDir, '..');
export const CHANGELOG = join(ROOT, 'CHANGELOG.md');

export function sh(cmd, args, opts = {}) {
  return execFileSync(cmd, args, {
    encoding: 'utf8',
    cwd: ROOT,
    stdio: ['pipe', 'pipe', 'pipe'],
    ...opts,
  }).trim();
}

export function resolveRef(ref) {
  try {
    sh('git', ['rev-parse', '--verify', ref]);
    return ref;
  } catch {
    return null;
  }
}

export function resolveBaseRef(base) {
  if (resolveRef(base)) return base;
  if (!base.includes('/')) {
    const originBranch = `origin/${base}`;
    if (resolveRef(originBranch)) return originBranch;
  }
  return null;
}

export function currentBranch() {
  const fromEnv =
    process.env.CI_COMMIT_REF_NAME ||
    process.env.GITHUB_REF_NAME ||
    process.env.BRANCH_NAME ||
    process.env.GIT_BRANCH;
  if (fromEnv) return fromEnv.replace(/^refs\/heads\//, '');
  try {
    return sh('git', ['rev-parse', '--abbrev-ref', 'HEAD']);
  } catch {
    return '';
  }
}

export function defaultReleaseBranch() {
  if (resolveRef('origin/main')) return 'main';
  if (resolveRef('origin/master')) return 'master';
  if (resolveRef('main')) return 'main';
  return 'master';
}

export function todayStr() {
  if (process.env.CHANGELOG_DATE && /^\d{4}-\d{2}-\d{2}$/.test(process.env.CHANGELOG_DATE)) {
    return process.env.CHANGELOG_DATE;
  }
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

const NOISE = /^(chore|style|ci|build|test|docs)(\([^)]*\))?:/i;

export function shouldSkipSubject(line) {
  const s = line.trim();
  if (!s) return true;
  if (NOISE.test(s)) return true;
  if (/^(eslint|prettier|lint|oxlint|oxfmt|swiftlint|swiftformat)\b/i.test(s)) return true;
  if (/^merge\b/i.test(s)) return true;
  if (/^revert(\s|\()/i.test(s)) return true;
  return false;
}

export function getLastTag() {
  try {
    const tag = sh('git', ['describe', '--tags', '--abbrev=0']);
    return tag || null;
  } catch {
    return null;
  }
}

export function collectCommitSubjects(fromRef, toRef = 'HEAD') {
  try {
    const range = fromRef ? `${fromRef}..${toRef}` : toRef;
    const raw = sh('git', ['log', range, '--no-merges', '--format=%s%n%b---COMMIT---']);
    if (!raw) return [];
    const seen = new Set();
    const out = [];
    for (const chunk of raw.split('---COMMIT---')) {
      const lines = chunk.trim().split('\n');
      const subject = (lines[0] || '').trim();
      const body = lines.slice(1).join('\n');
      if (!subject || shouldSkipSubject(subject) || seen.has(subject)) continue;
      seen.add(subject);
      out.push({ subject, body, full: `${subject}\n${body}`.trim() });
    }
    return out;
  } catch {
    return [];
  }
}

export function collectSubjects(fromRef, toRef = 'HEAD') {
  return collectCommitSubjects(fromRef, toRef).map(c => c.subject);
}

export function parseLatestReleasedVersion(content) {
  const re = /^## \[(\d+\.\d+\.\d+)\]\s*-\s*(\d{4}-\d{2}-\d{2})/gm;
  let match;
  let latest = null;
  while ((match = re.exec(content)) !== null) {
    latest = { version: match[1], date: match[2] };
  }
  return latest;
}

export function parseUnreleasedBlock(content) {
  const lines = content.split('\n');
  const start = lines.findIndex(l => /^## \[Unreleased\]/i.test(l.trim()));
  if (start === -1) return null;
  let end = lines.length;
  for (let i = start + 1; i < lines.length; i++) {
    if (/^## \[\d+\.\d+\.\d+\]/.test(lines[i].trim())) {
      end = i;
      break;
    }
  }
  return {
    start,
    end,
    body: lines.slice(start + 1, end).join('\n').trim(),
  };
}

export function buildUnreleasedBlock(subjects) {
  const header = '## [Unreleased]';
  if (subjects.length === 0) {
    return `${header}\n\n**变更摘要**（自动生成）\n\n- （相对上一版本无有效提交，或均为工程类提交）\n`;
  }
  const bullets = subjects.map(s => `- ${s}`).join('\n');
  return `${header}\n\n**变更摘要**（自动生成，发版前可删减无关项并改写）\n\n${bullets}\n`;
}

export function replaceUnreleasedSection(content, unreleasedBlock) {
  const block = parseUnreleasedBlock(content);
  const lines = content.split('\n');
  if (block) {
    const before = lines.slice(0, block.start).join('\n');
    const after = lines.slice(block.end).join('\n');
    const merged = `${before.replace(/\n+$/, '\n')}\n${unreleasedBlock.replace(/\n+$/, '\n')}\n${after.replace(/^\n+/, '')}`;
    return merged.trimEnd() + '\n';
  }
  const latest = parseLatestReleasedVersion(content);
  if (latest) {
    const re = new RegExp(`^(## \\[${latest.version.replace(/\./g, '\\.')}\\].*)$`, 'm');
    const idx = content.search(re);
    if (idx !== -1) {
      return `${content.slice(0, idx).trimEnd()}\n\n${unreleasedBlock}\n${content.slice(idx)}`;
    }
  }
  const sep = content.indexOf('\n---\n');
  if (sep !== -1) {
    return `${content.slice(0, sep + 5).trimEnd()}\n\n${unreleasedBlock}\n${content.slice(sep + 5).replace(/^\n+/, '')}`;
  }
  return `${content.trimEnd()}\n\n${unreleasedBlock}\n`;
}

export function finalizeReleaseSection(unreleasedBody, version, date) {
  const body = unreleasedBody.trim() || '**变更摘要**\n- （无详细说明）';
  return `## [${version}] - ${date}\n\n${body}\n`;
}

export function parseVersion(version) {
  const m = String(version)
    .replace(/^v/, '')
    .match(/^(\d+)\.(\d+)\.(\d+)/);
  if (!m) return null;
  return { major: +m[1], minor: +m[2], patch: +m[3] };
}

export function formatVersion(v) {
  return `${v.major}.${v.minor}.${v.patch}`;
}

export function bumpVersion(version, type) {
  const v = parseVersion(version);
  if (!v) throw new Error(`Invalid version: ${version}`);
  if (type === 'major') return formatVersion({ major: v.major + 1, minor: 0, patch: 0 });
  if (type === 'minor') return formatVersion({ major: v.major, minor: v.minor + 1, patch: 0 });
  if (type === 'patch') return formatVersion({ major: v.major, minor: v.minor, patch: v.patch + 1 });
  throw new Error(`Unknown bump type: ${type}`);
}

export function inferBumpType(commits) {
  let bump = null;
  for (const c of commits) {
    const text = c.full;
    if (/^(\w+)(\(.*\))?!:/.test(c.subject) || /BREAKING CHANGE/i.test(text)) {
      return 'major';
    }
    if (/^feat(\(.*\))?:/.test(c.subject)) {
      bump = bump === 'patch' ? 'minor' : bump || 'minor';
    } else if (/^(fix|perf|refactor)(\(.*\))?:/.test(c.subject)) {
      bump = bump || 'patch';
    }
  }
  return bump;
}

export function getCurrentVersionFromTagOrChangelog(content) {
  const tag = getLastTag();
  if (tag) {
    const v = parseVersion(tag);
    if (v) return formatVersion(v);
  }
  const latest = parseLatestReleasedVersion(content);
  return latest?.version || '0.0.0';
}

export function resolveLogBaseRef(content) {
  const tag = getLastTag();
  if (tag && resolveRef(tag)) return tag;
  const latest = parseLatestReleasedVersion(content);
  if (latest?.date) {
    try {
      const hash = sh('git', ['rev-list', '-1', '--before', `${latest.date}T23:59:59`, 'HEAD']);
      if (hash) return hash;
    } catch {
      /* fall through */
    }
  }
  for (const ref of ['origin/main', 'origin/master', 'main', 'master']) {
    if (resolveRef(ref)) return ref;
  }
  return null;
}

export function readChangelog() {
  if (!existsSync(CHANGELOG)) {
    return '# Changelog\n\n';
  }
  return readFileSync(CHANGELOG, 'utf8');
}

export function writeChangelog(content) {
  writeFileSync(CHANGELOG, content, 'utf8');
}

export function readPackageVersion() {
  const pkgPath = join(ROOT, 'package.json');
  if (!existsSync(pkgPath)) return null;
  try {
    const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'));
    return pkg.version || null;
  } catch {
    return null;
  }
}

export function writePackageVersion(version) {
  const pkgPath = join(ROOT, 'package.json');
  if (!existsSync(pkgPath)) return false;
  const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'));
  if (!pkg.version) return false;
  pkg.version = version;
  writeFileSync(pkgPath, `${JSON.stringify(pkg, null, 2)}\n`, 'utf8');
  return true;
}

export function assertReleaseBranch() {
  const branch = currentBranch();
  const allowed = new Set(['main', 'master']);
  if (!allowed.has(branch)) {
    throw new Error(`Release must run on main/master (current: ${branch})`);
  }
}

export function assertCleanWorkingTree(allowFiles = []) {
  const allow = new Set(allowFiles);
  const status = sh('git', ['status', '--porcelain']);
  if (!status) return;
  const dirty = status
    .split('\n')
    .filter(Boolean)
    .filter(line => {
      const file = line.slice(3).trim();
      return !allow.has(file);
    });
  if (dirty.length && process.env.RELEASE_IN_CI !== '1') {
    throw new Error(`Working tree not clean:\n${dirty.join('\n')}`);
  }
}
