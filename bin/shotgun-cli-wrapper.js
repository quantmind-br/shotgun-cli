#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Determine the binary name based on platform
let binaryName = 'shotgun-cli';
if (process.platform === 'win32') {
  binaryName += '.exe';
} else if (process.platform === 'linux') {
  binaryName += '-linux';
} else if (process.platform === 'darwin') {
  if (process.arch === 'arm64') {
    binaryName += '-macos-arm64';
  } else {
    binaryName += '-macos';
  }
}

const binaryPath = path.join(__dirname, binaryName);

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error(`Error: Binary not found at ${binaryPath}`);
  console.error('Please run "npm run build" to build the binary for your platform.');
  process.exit(1);
}

// Spawn the Go binary with all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit'
});

// Forward exit code
child.on('exit', (code) => {
  process.exit(code || 0);
});

// Handle process termination
process.on('SIGINT', () => {
  child.kill('SIGINT');
});

process.on('SIGTERM', () => {
  child.kill('SIGTERM');
});