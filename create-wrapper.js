const fs = require('fs');
const path = require('path');

const wrapperContent = `#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os');

// Determine the correct binary based on the platform
const isWindows = os.platform() === 'win32';
const binaryName = isWindows ? 'shotgun.exe' : 'shotgun';
const binaryPath = path.join(__dirname, binaryName);

// Check if the binary exists
if (!fs.existsSync(binaryPath)) {
    console.error(\`❌ Binary not found: \${binaryPath}\`);
    console.error('Please run: npm run build');
    process.exit(1);
}

// Spawn the binary with all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: 'inherit',
    shell: false
});

child.on('error', (error) => {
    console.error('Failed to start the application:', error);
    process.exit(1);
});

child.on('exit', (code) => {
    process.exit(code);
});`;

// Ensure bin directory exists
if (!fs.existsSync('bin')) {
    fs.mkdirSync('bin', { recursive: true });
}

// Write the wrapper script
fs.writeFileSync('bin/shotgun.js', wrapperContent);
console.log('✅ Created wrapper script: bin/shotgun.js');