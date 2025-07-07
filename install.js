#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const GITHUB_REPO = 'your-username/shotgun-cli'; // Replace with actual repository
const VERSION = require('./package.json').version;

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;
  
  const platformMap = {
    'darwin': 'darwin',
    'linux': 'linux',
    'win32': 'windows'
  };
  
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  
  const mappedPlatform = platformMap[platform];
  const mappedArch = archMap[arch];
  
  if (!mappedPlatform || !mappedArch) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }
  
  return {
    platform: mappedPlatform,
    arch: mappedArch,
    ext: platform === 'win32' ? '.exe' : ''
  };
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        return downloadFile(response.headers.location, dest)
          .then(resolve)
          .catch(reject);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: ${response.statusCode}`));
        return;
      }
      
      response.pipe(file);
      
      file.on('finish', () => {
        file.close();
        resolve();
      });
      
      file.on('error', (err) => {
        fs.unlink(dest, () => {}); // Delete the file on error
        reject(err);
      });
    }).on('error', (err) => {
      reject(err);
    });
  });
}

async function installBinary() {
  try {
    console.log('Installing shotgun-cli binary...');
    
    const { platform, arch, ext } = getPlatformInfo();
    const binaryName = `shotgun-cli-${platform}-${arch}${ext}`;
    const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${binaryName}`;
    
    // Ensure bin directory exists
    const binDir = path.join(__dirname, 'bin');
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    const binaryPath = path.join(binDir, `shotgun-cli${ext}`);
    
    console.log(`Downloading from: ${downloadUrl}`);
    await downloadFile(downloadUrl, binaryPath);
    
    // Make binary executable on Unix-like systems
    if (platform !== 'windows') {
      try {
        fs.chmodSync(binaryPath, 0o755);
      } catch (err) {
        console.warn('Warning: Could not make binary executable:', err.message);
      }
    }
    
    console.log('✅ shotgun-cli installed successfully!');
    console.log(`Binary location: ${binaryPath}`);
    console.log('You can now use: npx shotgun-cli');
    
  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    console.error('\nFallback: You can download the binary manually from:');
    console.error(`https://github.com/${GITHUB_REPO}/releases/latest`);
    process.exit(1);
  }
}

// Run installation
if (require.main === module) {
  installBinary();
}

module.exports = { installBinary, getPlatformInfo };