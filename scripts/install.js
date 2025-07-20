#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const os = require('os');

function main() {
  console.log('Installing shotgun-cli...');
  
  // Determine the binary name based on platform
  let binaryName = 'shotgun-cli';
  if (process.platform === 'win32') {
    binaryName += '.exe';
  }
  
  const binPath = path.join(__dirname, '..', 'bin', binaryName);
  
  // Check if binary exists
  if (!fs.existsSync(binPath)) {
    console.log('Binary not found, building from source...');
    
    // Try to build the binary
    const { execSync } = require('child_process');
    try {
      const buildPath = path.join(__dirname, '..');
      process.chdir(buildPath);
      
      // Build the binary
      const buildCmd = `go build -o bin/${binaryName} .`;
      console.log(`Running: ${buildCmd}`);
      execSync(buildCmd, { stdio: 'inherit' });
      
      console.log('✓ Binary built successfully');
    } catch (error) {
      console.error('Failed to build binary:', error.message);
      console.log('Please ensure Go is installed and try building manually:');
      console.log('  go build -o bin/shotgun-cli .');
      process.exit(1);
    }
  } else {
    console.log('✓ Binary found');
  }
  
  // Make binary executable on Unix systems
  if (process.platform !== 'win32') {
    try {
      fs.chmodSync(binPath, '755');
      console.log('✓ Binary permissions set');
    } catch (error) {
      console.warn('Warning: Could not set executable permissions');
    }
  }
  
  console.log('✓ shotgun-cli installed successfully!');
  console.log('');
  console.log('Usage:');
  console.log('  shotgun-cli --version');
  console.log('  shotgun-cli --help');
  console.log('  shotgun-cli           # Start interactive mode');
}

if (require.main === module) {
  main();
}