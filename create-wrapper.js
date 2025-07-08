const fs = require('fs');
const path = require('path');

// This script (create-wrapper.js) will always be run by Node.js.
// The generated wrapper script (bin/shotgun.js) will also be run by Node.js.
// os.platform() inside the *wrapper* will reflect where that Node instance is running.
// - If Node is in WSL, os.platform() is 'linux'.
// - If Node is in Windows (Git Bash, CMD, PS), os.platform() is 'win32'.

const wrapperContent = `#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os'); // os module for the wrapper script itself

// Path to the directory containing this wrapper script (expected to be 'bin/')
const scriptDir = __dirname;

// Determine the name of the actual binary to execute.
// The binary (shotgun or shotgun.exe) is expected to be in the same directory as this wrapper.

const windowsBinaryPath = path.join(scriptDir, 'shotgun.exe');
const unixBinaryPath = path.join(scriptDir, 'shotgun');

let executableBinaryPath;
let isWindowsTargetBinary = false;

if (fs.existsSync(windowsBinaryPath)) {
    // If shotgun.exe exists, we assume it's the intended target.
    executableBinaryPath = windowsBinaryPath;
    isWindowsTargetBinary = true;
    // console.log("Wrapper: Found shotgun.exe, will attempt to execute it."); // For debugging
} else if (fs.existsSync(unixBinaryPath)) {
    // Otherwise, if 'shotgun' (no extension) exists, use that.
    executableBinaryPath = unixBinaryPath;
    // console.log("Wrapper: Found shotgun (unix-like), will attempt to execute it."); // For debugging
} else {
    console.error(\`❌ Target binary not found in \${scriptDir}. Looked for 'shotgun.exe' and 'shotgun'.\`);
    console.error('Please ensure the binary was downloaded or built correctly (e.g., run "npm install" or "npm run build:platform").');
    process.exit(1);
}

// console.log(\`Wrapper: Determined executable path: \${executableBinaryPath}\`); // For debugging
// console.log(\`Wrapper: Running on os.platform(): \${os.platform()}\`); // For debugging

const spawnOptions = {
    stdio: 'inherit',
    shell: false // Default to false for security and direct execution
};

// Special handling for WSL trying to run a .exe
// If the wrapper is running in WSL (os.platform() === 'linux') AND it's trying to run a .exe file
if (os.platform() === 'linux' && isWindowsTargetBinary) {
    // console.log("Wrapper: WSL environment trying to run a Windows .exe."); // For debugging
    // WSL can often execute .exe files directly if they are in the WSL path or have execute permissions.
    // However, sometimes explicitly using 'cmd.exe /c' or ensuring the .exe is +x can help.
    // For now, direct execution is attempted. If issues arise, 'shell: true' or 'cmd.exe /c' might be needed.
    // but that can be complex due to path translations.
    // The 'ENOENT' error handler below provides a hint for this case.
}


const child = spawn(executableBinaryPath, process.argv.slice(2), spawnOptions);

child.on('error', (error) => {
    console.error(\`Failed to start the application '\${executableBinaryPath}':\`, error);
    if (error.code === 'ENOENT') {
        console.error("This usually means the binary is not found at the specified path, or it's not executable.");
        if (os.platform() === 'linux' && isWindowsTargetBinary) {
            console.error("Hint: If you are in a WSL environment trying to run a Windows .exe,");
            console.error("ensure the .exe has execute permissions (e.g., chmod +x path/to/shotgun.exe within WSL),");
            console.error("and that its location is correctly resolved. Sometimes, WSL interop issues can cause this.");
        }
    }
    process.exit(1);
});

child.on('exit', (code, signal) => {
    if (signal) {
        // console.log(\`Application terminated by signal \${signal}\`); // For debugging
        process.exit(1); // Standard practice to exit with 1 on signal termination
    } else {
        process.exit(code === null ? 1 : code);
    }
});
`;

// This script (create-wrapper.js) is expected to be in the project root.
// The wrapper (shotgun.js) should be created in the 'bin/' directory.
const binDir = path.join(__dirname, 'bin');
const wrapperPath = path.join(binDir, 'shotgun.js'); // This is the script listed in package.json "bin"

// Ensure bin directory exists
if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
}

// Write the wrapper script (bin/shotgun.js)
fs.writeFileSync(wrapperPath, wrapperContent);

// Make the wrapper script itself executable
// This is crucial for it to be run directly on Unix-like systems (Linux, macOS, WSL)
try {
    fs.chmodSync(wrapperPath, 0o755);
} catch (e) {
    console.warn(`Could not make wrapper script ${wrapperPath} executable: ${e.message}. You may need to do this manually.`);
}

console.log(`✅ Created/Updated Node.js wrapper script: ${wrapperPath}`);
