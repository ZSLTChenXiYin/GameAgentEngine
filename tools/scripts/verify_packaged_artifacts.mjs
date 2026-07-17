#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { execFileSync } from 'node:child_process';

const ROOT = path.resolve(process.cwd(), process.argv[2] || 'dist');
const REQUIRED_SDK_FILES = [
  'sdks/ts-sdk/tsconfig.json',
  'sdks/js-sdk/src/client.js',
  'sdks/java-sdk/src/GameAgentEngineClient.java',
  'sdks/lua-sdk/src/client.lua',
  'sdks/c-sdk/src/game_agent_engine_client.h',
  'sdks/cpp-sdk/src/game_agent_engine_client.hpp',
  'sdks/cs-sdk/GameAgentEngine.SDK.csproj',
  'sdks/gd-sdk/src/game_agent_engine_client.gd',
];

const REQUIRED_RUNTIME_FILES = [
  'GameAgentEngine{ext}',
  'GameAgentDevCli{ext}',
  'GameAgentWorker{ext}',
  'gameagentengine.conf.yaml',
  'README.md',
  'workerhome/demo/demo-world.yaml',
  'workerhome/demo/demo-state.yaml',
  'workerhome/fixtures/runtime_task_dynamic_interfaces.json',
  'web/GameAgentCreator/index.html',
  'web/GameAgentCreator/js/version.js',
];

function fail(message) {
  throw new Error(message);
}

function isDir(p) {
  return fs.existsSync(p) && fs.statSync(p).isDirectory();
}

function isFile(p) {
  return fs.existsSync(p) && fs.statSync(p).isFile();
}

function assertExists(root, relativePath) {
  const full = path.join(root, relativePath);
  if (!fs.existsSync(full)) {
    fail(`missing ${relativePath} under ${root}`);
  }
  return full;
}

function assertContains(file, needle) {
  const text = fs.readFileSync(file, 'utf8');
  if (!text.includes(needle)) {
    fail(`expected ${path.relative(ROOT, file)} to contain ${needle}`);
  }
}

function psQuote(value) {
  return `'${String(value).replace(/'/g, "''")}'`;
}

function extractZip(zipPath, destination) {
  const commands = [
    ['powershell.exe', ['-NoProfile', '-Command', `Expand-Archive -LiteralPath ${psQuote(zipPath)} -DestinationPath ${psQuote(destination)} -Force`]],
    ['pwsh', ['-NoProfile', '-Command', `Expand-Archive -LiteralPath ${psQuote(zipPath)} -DestinationPath ${psQuote(destination)} -Force`]],
    ['python', ['-m', 'zipfile', '-e', zipPath, destination]],
  ];

  let lastError = null;
  for (const [cmd, args] of commands) {
    try {
      execFileSync(cmd, args, { stdio: 'inherit' });
      return;
    } catch (err) {
      lastError = err;
    }
  }

  throw lastError || new Error(`unable to extract ${zipPath}`);
}

function parsePackageName(name) {
  if (!name.startsWith('GameAgentEngine-')) {
    return null;
  }
  const body = name.slice('GameAgentEngine-'.length);
  const versionPos = body.lastIndexOf('-v');
  if (versionPos < 0) {
    return null;
  }
  const platform = body.slice(0, versionPos);
  const version = body.slice(versionPos + 1);
  const dash = platform.lastIndexOf('-');
  if (dash < 0) {
    return null;
  }
  return {
    platform,
    version,
    os: platform.slice(0, dash),
    arch: platform.slice(dash + 1),
  };
}

function listPackageRoots(root) {
  if (isFile(root) && root.endsWith('.zip')) {
    return [root];
  }

  if (!isDir(root)) {
    fail(`root path does not exist: ${root}`);
  }

  const entries = fs.readdirSync(root, { withFileTypes: true });
  const results = [];
  for (const entry of entries) {
    const full = path.join(root, entry.name);
    if (entry.isDirectory() && parsePackageName(entry.name)) {
      results.push(full);
      continue;
    }
    if (entry.isFile() && entry.name.endsWith('.zip') && parsePackageName(entry.name.slice(0, -4))) {
      results.push(full);
    }
  }
  return results.sort();
}

function verifyPackageDir(dir, options = {}) {
  const { requireArchive = true } = options;
  const base = path.basename(dir);
  const meta = parsePackageName(base);
  if (!meta) {
    fail(`unexpected package directory name: ${base}`);
  }

  const exeExt = meta.os === 'windows' ? '.exe' : '';
  console.log(`verifying ${base}`);

  for (const rel of REQUIRED_RUNTIME_FILES) {
    assertExists(dir, rel.replace('{ext}', exeExt));
  }

  for (const rel of REQUIRED_SDK_FILES) {
    assertExists(dir, rel);
  }

  const versionFile = path.join(dir, 'web/GameAgentCreator/js/version.js');
  assertContains(versionFile, `CREATOR_MIN_COMPATIBLE = "${meta.version}"`);

  if (requireArchive) {
    const zipPath = `${dir}.zip`;
    if (!isFile(zipPath)) {
      fail(`missing archive: ${path.basename(zipPath)}`);
    }

    const zipSize = fs.statSync(zipPath).size;
    if (zipSize <= 0) {
      fail(`archive is empty: ${path.basename(zipPath)}`);
    }
  }

  console.log(`ok ${base} (${meta.platform}, ${meta.version})`);
}

function verifyZipFile(zipPath) {
  const base = path.basename(zipPath, '.zip');
  const meta = parsePackageName(base);
  if (!meta) {
    fail(`unexpected archive name: ${path.basename(zipPath)}`);
  }

  const tempRoot = path.join(ROOT, '..', 'tmp', 'package-acceptance', base);
  fs.rmSync(tempRoot, { recursive: true, force: true });
  fs.mkdirSync(tempRoot, { recursive: true });
  try {
    extractZip(zipPath, tempRoot);
    const extracted = path.join(tempRoot, base);
    if (isDir(extracted)) {
      verifyPackageDir(extracted, { requireArchive: false });
      return;
    }

    const marker = path.join(tempRoot, meta.os === 'windows' ? 'GameAgentEngine.exe' : 'GameAgentEngine');
    if (isFile(marker)) {
      verifyPackageDir(tempRoot, { requireArchive: false });
      return;
    }

    fail(`extracted package root not found: ${base}`);
  } finally {
    fs.rmSync(tempRoot, { recursive: true, force: true });
  }
}

function main() {
  const targets = listPackageRoots(ROOT);
  if (targets.length === 0) {
    fail(`no package targets found under ${ROOT}`);
  }

  for (const target of targets) {
    if (target.endsWith('.zip')) {
      verifyZipFile(target);
    } else {
      verifyPackageDir(target);
    }
  }

  console.log(`verified ${targets.length} package target(s)`);
}

main();
