#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import process from 'node:process';
import { spawn } from 'node:child_process';
import { setTimeout as delay } from 'node:timers/promises';

const ROOT = path.resolve(process.cwd());
const PACKAGE_DIR = path.resolve(process.argv[2] || path.join('dist', 'GameAgentEngine-windows-amd64-v0.5.0'));
const TMP_ROOT = path.join(ROOT, 'tmp', 'f8-acceptance');
const ENGINE_PORT = Number(process.env.F8_ENGINE_PORT || 18180);
const ENGINE_BASE_URL = `http://127.0.0.1:${ENGINE_PORT}`;

function fail(message) {
  throw new Error(message);
}

function ensureFile(filePath) {
  if (!fs.existsSync(filePath) || !fs.statSync(filePath).isFile()) {
    fail(`missing file: ${filePath}`);
  }
}

function ensureDir(dirPath) {
  if (!fs.existsSync(dirPath) || !fs.statSync(dirPath).isDirectory()) {
    fail(`missing directory: ${dirPath}`);
  }
}

function quoted(value) {
  return JSON.stringify(String(value));
}

function execTool(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: options.cwd || ROOT,
      env: options.env || process.env,
      stdio: ['ignore', 'pipe', 'pipe'],
      shell: false,
    });

    let stdout = '';
    let stderr = '';
    child.stdout.on('data', (chunk) => {
      stdout += chunk.toString();
      if (!options.silent) process.stdout.write(chunk);
    });
    child.stderr.on('data', (chunk) => {
      stderr += chunk.toString();
      if (!options.silent) process.stderr.write(chunk);
    });
    child.on('error', reject);
    child.on('close', (code) => {
      if (code === 0) {
        resolve({ stdout, stderr, code });
      } else {
        const err = new Error(`${command} ${args.join(' ')} exited with code ${code}`);
        err.stdout = stdout;
        err.stderr = stderr;
        err.code = code;
        reject(err);
      }
    });
  });
}

async function waitForHealth(url, timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;
  let lastError = null;
  while (Date.now() < deadline) {
    try {
      const resp = await fetch(url, { headers: { 'X-API-Key': 'dev-key' } });
      if (resp.ok) return;
      lastError = new Error(`health status ${resp.status}`);
    } catch (err) {
      lastError = err;
    }
    await delay(500);
  }
  throw lastError || new Error(`health check timed out: ${url}`);
}

function platformExecutableName(base) {
  return process.platform === 'win32' ? `${base}.exe` : base;
}

function rewritePackageConfig(configText, dbPath, port) {
  return configText
    .replace(/port:\s*8080/, `port: ${port}`)
    .replace(/dsn:\s*"gameagentengine\.db"/, `dsn: ${quoted(dbPath)}`)
    .replace(/provider:\s*"openai"/, 'provider: "mock"')
    .replace(/api_key:\s*"sk-xxx"/, 'api_key: ""')
    .replace(/base_url:\s*"https:\/\/api\.deepseek\.com"/, 'base_url: ""');
}

async function startEngine() {
  const engineExe = path.join(PACKAGE_DIR, platformExecutableName('GameAgentEngine'));
  const configTemplate = path.join(PACKAGE_DIR, 'gameagentengine.conf.yaml');
  ensureFile(engineExe);
  ensureFile(configTemplate);

  const configDir = path.join(TMP_ROOT, 'engine');
  fs.rmSync(configDir, { recursive: true, force: true });
  fs.mkdirSync(configDir, { recursive: true });

  const dbPath = path.join(configDir, 'f8-acceptance.db');
  const configPath = path.join(configDir, 'gameagentengine.conf.yaml');
  const configText = fs.readFileSync(configTemplate, 'utf8');
  fs.writeFileSync(configPath, rewritePackageConfig(configText, dbPath, ENGINE_PORT), 'utf8');

  const stdoutPath = path.join(configDir, 'engine.stdout.log');
  const stderrPath = path.join(configDir, 'engine.stderr.log');
  const stdout = fs.createWriteStream(stdoutPath, { flags: 'w' });
  const stderr = fs.createWriteStream(stderrPath, { flags: 'w' });
  const child = spawn(engineExe, ['serve', '--config', configPath], {
    cwd: PACKAGE_DIR,
    env: process.env,
    stdio: ['ignore', 'pipe', 'pipe'],
    shell: false,
  });
  child.stdout.pipe(stdout);
  child.stderr.pipe(stderr);

  const stop = async () => {
    if (child.exitCode !== null || child.killed) return;
    child.kill();
    await new Promise((resolve) => {
      const timer = setTimeout(() => resolve(), 5000);
      child.once('exit', () => {
        clearTimeout(timer);
        resolve();
      });
    });
    stdout.end();
    stderr.end();
  };

  await waitForHealth(`${ENGINE_BASE_URL}/health`);
  return { child, stop, configPath, dbPath, stdoutPath, stderrPath };
}

async function runDevCliSmoke() {
  const devcliExe = path.join(PACKAGE_DIR, platformExecutableName('GameAgentDevCli'));
  ensureFile(devcliExe);
  const status = await execTool(devcliExe, ['--server', ENGINE_BASE_URL, '--key', 'dev-key', 'status'], { cwd: PACKAGE_DIR });
  if (!status.stdout.includes('"status": "ok"')) {
    fail('DevCli status output did not include status ok');
  }
  const version = await execTool(devcliExe, ['--server', ENGINE_BASE_URL, '--key', 'dev-key', 'version'], { cwd: PACKAGE_DIR });
  if (!version.stdout.includes('GameAgentDevCli') || !version.stdout.includes('Engine')) {
    fail('DevCli version output missing version lines');
  }
}

async function runGoSdkSmoke() {
  const smokeDir = path.join(TMP_ROOT, 'go-sdk');
  fs.rmSync(smokeDir, { recursive: true, force: true });
  fs.mkdirSync(smokeDir, { recursive: true });
  const smokeFile = path.join(smokeDir, 'main.go');
  const smokeSource = `package main

import (
  "fmt"
  "os"

  "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func main() {
  if len(os.Args) < 3 {
    panic("usage: go run main.go <base_url> <api_key>")
  }
  client := sdk.NewClient(os.Args[1], os.Args[2])
  if err := client.Health(); err != nil { panic(err) }
  ver, min, err := client.GetVersion()
  if err != nil { panic(err) }
  worlds, err := client.GetWorlds()
  if err != nil { panic(err) }
  fmt.Printf("sdk-smoke ok version=%s min=%s worlds=%d\\n", ver, min, len(worlds))
}
`;
  fs.writeFileSync(smokeFile, smokeSource, 'utf8');
  const result = await execTool('go', ['run', smokeFile, ENGINE_BASE_URL, 'dev-key'], { cwd: ROOT });
  if (!result.stdout.includes('sdk-smoke ok')) {
    fail('Go SDK smoke output missing success marker');
  }
}

async function runWorkerToolingSmoke() {
  const workerExe = path.join(PACKAGE_DIR, platformExecutableName('GameAgentWorker'));
  const testsDir = path.join(PACKAGE_DIR, 'workerhome', 'fixtures');
  ensureFile(workerExe);
  ensureDir(testsDir);
  const outFile = path.join(TMP_ROOT, 'tooling-smoke.json');
  fs.rmSync(outFile, { force: true });
  await execTool(workerExe, [
    'test',
    'tooling-smoke',
    '--engine-exe', path.join(PACKAGE_DIR, platformExecutableName('GameAgentEngine')),
    '--devcli-exe', path.join(PACKAGE_DIR, platformExecutableName('GameAgentDevCli')),
    '--tests-dir', testsDir,
    '--out', outFile,
    '--json',
  ], { cwd: PACKAGE_DIR });
  ensureFile(outFile);
  const payload = JSON.parse(fs.readFileSync(outFile, 'utf8'));
  if (!payload || !Array.isArray(payload.checks) || payload.checks.length === 0) {
    fail('tooling-smoke result missing checks');
  }
  if (!payload.pending_task_id) {
    fail('tooling-smoke result missing pending_task_id');
  }
}

async function main() {
  ensureDir(PACKAGE_DIR);
  fs.rmSync(TMP_ROOT, { recursive: true, force: true });
  fs.mkdirSync(TMP_ROOT, { recursive: true });

  const verifier = path.join(ROOT, 'tools', 'scripts', 'verify_packaged_artifacts.mjs');
  await execTool('node', [verifier, PACKAGE_DIR], { cwd: ROOT });

  const engine = await startEngine();
  try {
    await runDevCliSmoke();
    await runGoSdkSmoke();
    await runWorkerToolingSmoke();
  } finally {
    await engine.stop();
  }

  console.log(`F8 acceptance complete for ${PACKAGE_DIR}`);
}

main().catch((err) => {
  console.error(err && err.stack ? err.stack : err);
  process.exitCode = 1;
});
