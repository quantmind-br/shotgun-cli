#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

// Build configurations for different platforms
const buildConfigs = {
  windows: {
    env: { GOOS: 'windows', GOARCH: 'amd64' },
    output: 'bin/shotgun-cli.exe',
    description: 'Windows (x64)'
  },
  linux: {
    env: { GOOS: 'linux', GOARCH: 'amd64' },
    output: 'bin/shotgun-cli-linux',
    description: 'Linux (x64)'
  },
  macos: {
    env: { GOOS: 'darwin', GOARCH: 'amd64' },
    output: 'bin/shotgun-cli-macos',
    description: 'macOS (x64)'
  },
  'macos-arm': {
    env: { GOOS: 'darwin', GOARCH: 'arm64' },
    output: 'bin/shotgun-cli-macos-arm64',
    description: 'macOS (ARM64)'
  }
};

function ensureBinDirectory() {
  const binDir = path.join(process.cwd(), 'bin');
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
    console.log('📁 Created bin/ directory');
  }
}

function buildForPlatform(platform, config) {
  console.log(`🔨 Building for ${config.description}...`);
  
  try {
    const env = { ...process.env, ...config.env };
    const command = `go build -ldflags="-s -w" -o ${config.output} .`;
    
    execSync(command, { 
      env,
      stdio: 'pipe'
    });
    
    console.log(`✅ Successfully built: ${config.output}`);
    
    // Check file size
    const stats = fs.statSync(config.output);
    const sizeInMB = (stats.size / (1024 * 1024)).toFixed(2);
    console.log(`📦 Binary size: ${sizeInMB} MB`);
    
  } catch (error) {
    console.error(`❌ Failed to build for ${platform}:`);
    console.error(error.message);
    process.exit(1);
  }
}

function main() {
  const target = process.argv[2] || 'current';
  
  console.log('🚀 Starting shotgun-cli build process...\n');
  
  // Ensure bin directory exists
  ensureBinDirectory();
  
  // Determine which builds to run
  let buildTargets = [];
  
  switch (target) {
    case 'all':
      buildTargets = ['windows', 'linux', 'macos', 'macos-arm'];
      break;
    case 'current':
      // Build for current platform
      const platform = process.platform;
      if (platform === 'win32') {
        buildTargets = ['windows'];
      } else if (platform === 'darwin') {
        buildTargets = ['macos'];
      } else {
        buildTargets = ['linux'];
      }
      break;
    case 'windows':
    case 'linux':
    case 'macos':
    case 'macos-arm':
      buildTargets = [target];
      break;
    default:
      console.error('❌ Invalid build target. Use: current, all, windows, linux, macos, or macos-arm');
      process.exit(1);
  }
  
  console.log(`🎯 Building for: ${buildTargets.join(', ')}\n`);
  
  // Run builds
  for (const target of buildTargets) {
    buildForPlatform(target, buildConfigs[target]);
    console.log('');
  }
  
  console.log('🎉 Build process completed successfully!');
  
  // Show available binaries
  console.log('\n📋 Available binaries:');
  const binDir = path.join(process.cwd(), 'bin');
  if (fs.existsSync(binDir)) {
    const files = fs.readdirSync(binDir)
      .filter(file => file.startsWith('shotgun-cli') && !file.endsWith('.js'))
      .sort();
    
    files.forEach(file => {
      const filePath = path.join(binDir, file);
      const stats = fs.statSync(filePath);
      const sizeInMB = (stats.size / (1024 * 1024)).toFixed(2);
      console.log(`  • ${file} (${sizeInMB} MB)`);
    });
  }
}

if (require.main === module) {
  main();
}

module.exports = { buildForPlatform, buildConfigs };