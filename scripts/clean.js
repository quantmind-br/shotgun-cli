#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function removeFile(filePath) {
  try {
    if (fs.existsSync(filePath)) {
      fs.unlinkSync(filePath);
      console.log(`🗑️  Removed: ${path.relative(process.cwd(), filePath)}`);
      return true;
    }
    return false;
  } catch (error) {
    console.error(`❌ Failed to remove ${filePath}: ${error.message}`);
    return false;
  }
}

function removeDirectory(dirPath) {
  try {
    if (fs.existsSync(dirPath)) {
      fs.rmSync(dirPath, { recursive: true, force: true });
      console.log(`🗑️  Removed directory: ${path.relative(process.cwd(), dirPath)}`);
      return true;
    }
    return false;
  } catch (error) {
    console.error(`❌ Failed to remove directory ${dirPath}: ${error.message}`);
    return false;
  }
}

function cleanBuildArtifacts() {
  console.log('🧹 Cleaning build artifacts...\n');
  
  const binDir = path.join(process.cwd(), 'bin');
  let removedCount = 0;
  
  // List of binary files to remove
  const binaryFiles = [
    'shotgun-cli',
    'shotgun-cli.exe',
    'shotgun-cli-linux',
    'shotgun-cli-macos',
    'shotgun-cli-macos-arm64',
    'shotgun-cli.exe~'  // Windows backup files
  ];
  
  // Remove specific binary files
  binaryFiles.forEach(fileName => {
    const filePath = path.join(binDir, fileName);
    if (removeFile(filePath)) {
      removedCount++;
    }
  });
  
  // Remove any other build artifacts
  const rootBinaries = [
    path.join(process.cwd(), 'shotgun-cli'),
    path.join(process.cwd(), 'shotgun-cli.exe')
  ];
  
  rootBinaries.forEach(filePath => {
    if (removeFile(filePath)) {
      removedCount++;
    }
  });
  
  // Check if bin directory is empty (except for wrapper)
  if (fs.existsSync(binDir)) {
    const remainingFiles = fs.readdirSync(binDir)
      .filter(file => file !== 'shotgun-cli-wrapper.js');
    
    if (remainingFiles.length === 0) {
      console.log('✅ All build artifacts cleaned');
    } else {
      console.log(`📝 Remaining files in bin/: ${remainingFiles.join(', ')}`);
    }
  }
  
  if (removedCount === 0) {
    console.log('✨ No build artifacts found to clean');
  } else {
    console.log(`\n✅ Cleaned ${removedCount} build artifact(s)`);
  }
}

function main() {
  const arg = process.argv[2];
  
  if (arg === '--help' || arg === '-h') {
    console.log(`
🧹 shotgun-cli Clean Script

USAGE:
  npm run clean                 Clean build artifacts
  node scripts/clean.js         Clean build artifacts
  
DESCRIPTION:
  Removes all built binaries and build artifacts while preserving
  the wrapper script and other important files.
  
FILES CLEANED:
  • bin/shotgun-cli*            All platform binaries
  • shotgun-cli, shotgun-cli.exe  Root directory binaries
  • *.exe~ backup files         Windows backup files
`);
    return;
  }
  
  cleanBuildArtifacts();
}

if (require.main === module) {
  main();
}

module.exports = { cleanBuildArtifacts };