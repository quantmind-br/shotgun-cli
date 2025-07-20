#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

function main() {
  // Determine the binary name based on platform
  let binaryName = 'shotgun-cli';
  if (process.platform === 'win32') {
    binaryName += '.exe';
  } else if (process.platform === 'linux') {
    binaryName += '-linux';
  } else if (process.platform === 'darwin') {
    // Check if ARM64 build is available
    if (process.arch === 'arm64' && fs.existsSync(path.join(__dirname, 'shotgun-cli-macos-arm64'))) {
      binaryName += '-macos-arm64';
    } else {
      binaryName += '-macos';
    }
  }
  
  const binaryPath = path.join(__dirname, binaryName);
  
  // Check if binary exists
  if (!fs.existsSync(binaryPath)) {
    console.error(`Error: Binary not found at ${binaryPath}`);
    console.error('Please run: npm run build');
    process.exit(1);
  }
  
  // Execute the binary with all arguments
  const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: 'inherit'
  });
  
  child.on('close', (code) => {
    process.exit(code);
  });
  
  child.on('error', (err) => {
    console.error('Failed to start shotgun-cli:', err.message);
    process.exit(1);
  });
}

if (require.main === module) {
  main();
}