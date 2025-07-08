#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

// Attempt to read GITHUB_REPO from package.json, otherwise use a default or placeholder
let GITHUB_REPO = 'your-username/shotgun-cli'; // Default placeholder
try {
    const pkg = require('./package.json');
    if (pkg.repository && typeof pkg.repository.url === 'string') {
        // Regex to capture 'username/repo' from various GitHub URL formats
        const repoUrlMatch = pkg.repository.url.match(/github\.com[/:]([^/]+\/[^/.]+?)(?:\.git)?$/);
        if (repoUrlMatch && repoUrlMatch[1]) {
            GITHUB_REPO = repoUrlMatch[1];
        }
    }
} catch (e) {
    // Silently ignore errors and use the default GITHUB_REPO
    // console.warn(`Could not read repository URL from package.json. Using default: ${GITHUB_REPO}. Error: ${e.message}`);
}

const VERSION = require('./package.json').version;

function getPlatformInfo() {
  const hostNodePlatform = process.platform; // Platform Node.js is running on (e.g., 'linux' in WSL, 'win32' on Windows)
  const hostNodeArch = process.arch; // Arch Node.js is running on (e.g., 'x64')

  // Check for WSL. These checks determine if Node is running *inside* a WSL environment.
  const isWSL = fs.existsSync('/proc/sys/fs/binfmt_misc/WSLInterop') ||
                process.env.WSL_DISTRO_NAME ||
                (process.env.WSL_INTEROP !== undefined) ||
                (process.env.IS_WSL === 'true') || // Custom flag potentially set by a calling script
                (process.env.WSLENV !== undefined && process.env.WSLENV.includes('WSL_INTEROP'));

  let targetBinaryPlatform; // The platform of the binary we intend to download

  if (isWSL) {
    // If running inside WSL, we want the Linux binary.
    targetBinaryPlatform = 'linux';
    // console.log("WSL environment detected by install.js. Targeting Linux binary."); // For debugging
  } else if (hostNodePlatform === 'win32') {
    // If on native Windows (not WSL), we want the Windows binary.
    targetBinaryPlatform = 'windows';
  } else if (hostNodePlatform === 'darwin') {
    targetBinaryPlatform = 'darwin';
  } else if (hostNodePlatform === 'linux') {
    // If on native Linux (not WSL), we want the Linux binary.
    targetBinaryPlatform = 'linux';
  } else {
    throw new Error(`Unsupported host platform: ${hostNodePlatform}`);
  }

  // Map Node.js architecture names to Go architecture names (used in binary naming)
  const archMap = {
    'x64': 'amd64',
    'arm64': 'arm64'
  };
  const targetBinaryArch = archMap[hostNodeArch];

  if (!targetBinaryArch) {
    throw new Error(`Unsupported architecture: ${hostNodeArch} for target binary platform: ${targetBinaryPlatform}`);
  }

  return {
    // `platform` and `arch` refer to the *target binary*
    platform: targetBinaryPlatform,
    arch: targetBinaryArch,
    // `ext` is determined by the *target binary's platform*
    ext: targetBinaryPlatform === 'windows' ? '.exe' : '',
    isWSL, // Information about the execution environment
    hostPlatform: hostNodePlatform // Information about the execution environment
  };
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);

    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Handle redirect
        // console.log(`Redirected to: ${response.headers.location}`); // For debugging
        return downloadFile(response.headers.location, dest)
          .then(resolve)
          .catch(reject);
      }

      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download ${url}: Status Code ${response.statusCode}`));
        return;
      }

      response.pipe(file);

      file.on('finish', () => {
        file.close(resolve); // Pass resolve to close to ensure it's called after stream closes
      });

      file.on('error', (err) => {
        fs.unlink(dest, () => {}); // Delete the file on error, ignore unlink errors
        reject(err);
      });
    }).on('error', (err) => {
      reject(new Error(`Error during HTTPS request to ${url}: ${err.message}`));
    });
  });
}

async function installBinary() {
  try {
    console.log('Attempting to install shotgun-cli binary...');

    const platformInfo = getPlatformInfo();
    const { platform: targetPlatform, arch: targetArch, ext: targetExt } = platformInfo;

    const binaryReleaseName = `shotgun-cli-${targetPlatform}-${targetArch}${targetExt}`;
    const downloadUrl = `https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${binaryReleaseName}`;

    const binDir = path.join(__dirname, 'bin');
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    // The actual filename in the bin/ directory will be 'shotgun' or 'shotgun.exe'
    const localBinaryFilename = `shotgun${targetExt}`; // This should be 'shotgun' or 'shotgun.exe'
    const binaryPath = path.join(binDir, localBinaryFilename);

    console.log(`Host Platform: ${platformInfo.hostPlatform}, Host Arch: ${process.arch}, WSL: ${platformInfo.isWSL}`);
    console.log(`Targeting Binary: ${binaryReleaseName} (for ${targetPlatform}-${targetArch})`);
    console.log(`Downloading from: ${downloadUrl}`);
    console.log(`Saving to: ${binaryPath}`);

    await downloadFile(downloadUrl, binaryPath);

    if (targetPlatform !== 'windows') {
      try {
        fs.chmodSync(binaryPath, 0o755);
        console.log(`Made ${binaryPath} executable.`);
      } catch (err) {
        console.warn(`Warning: Could not make binary ${binaryPath} executable: ${err.message}`);
      }
    }

    // Create the main wrapper script `bin/shotgun.js` which is defined in package.json "bin"
    // This wrapper will then call the downloaded binary (shotgun or shotgun.exe)
    const createWrapperScriptPath = path.join(__dirname, 'create-wrapper.js');
    if (fs.existsSync(createWrapperScriptPath)) {
        try {
            console.log('Executing create-wrapper.js to generate bin/shotgun.js...');
            execSync(`node "${createWrapperScriptPath}"`, { stdio: 'inherit' });
        } catch (wrapperError) {
            console.warn(`Warning: Failed to create wrapper script via create-wrapper.js: ${wrapperError.message}`);
        }
    } else {
        console.warn("Warning: create-wrapper.js not found. The CLI might not be directly invokable via 'shotgun' or 'shotgun-cli'.");
    }

    console.log('✅ shotgun-cli binary download and setup process completed.');
    console.log(`Binary should be at: ${binaryPath}`);
    console.log('If this script was run via "npm install", the command should now be available.');

  } catch (error) {
    console.error('❌ Installation failed:', error.message);
    const { platform: p, arch: a } = getPlatformInfo(); // Get info again for error message
    if (error.message.includes("Failed to download") && error.message.includes("404")) {
        console.error(`\nA binary for your platform/architecture (${p}-${a}) version v${VERSION} might not be available at the download URL.`);
        console.error(`Please check releases at: https://github.com/${GITHUB_REPO}/releases/tag/v${VERSION}`);
    } else {
        console.error(`\nIt's possible the binary for your platform/architecture (${p}-${a}) is not available, or there was a network/file system issue.`);
    }
    console.error('Fallback: You can try building from source (e.g., "npm run build:platform") or download manually from:');
    console.error(`https://github.com/${GITHUB_REPO}/releases/latest`);
    process.exit(1); // Exit with error code so npm knows postinstall failed
  }
}

// Run installation only if this script is executed directly
if (require.main === module) {
  installBinary();
}

// Export for potential external use or testing
module.exports = { installBinary, getPlatformInfo, GITHUB_REPO, VERSION };
