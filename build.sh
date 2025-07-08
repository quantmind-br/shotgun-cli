#!/bin/bash

echo "Building shotgun-cli for Linux/WSL..."

# Detect if running in WSL
IS_WSL=false
if grep -qi microsoft /proc/version || [ -n "$WSL_DISTRO_NAME" ] || [ -n "$WSL_INTEROP" ]; then
    IS_WSL=true
    echo "WSL environment detected."
fi

# Build the appropriate binary
if [ "$IS_WSL" = true ]; then
    # For WSL, we might prefer the Windows .exe if it's meant to be called from Windows side
    # or a Linux binary if it's for use purely within WSL.
    # Assuming a Linux binary for WSL for now.
    echo "Building Linux binary (can be used in WSL)..."
    npm run build:unix
    if [ $? -ne 0 ]; then
        echo "Build failed. Exiting."
        exit 1
    fi
else
    echo "Building Linux binary..."
    npm run build:unix
    if [ $? -ne 0 ]; then
        echo "Build failed. Exiting."
        exit 1
    fi
fi

echo "Creating wrapper script..."
node create-wrapper.js
if [ $? -ne 0 ]; then
    echo "Wrapper creation failed. Exiting."
    exit 1
fi

echo "Installing globally..."
npm install -g .
if [ $? -ne 0 ]; then
    echo "Global installation failed. Exiting."
    exit 1
fi

echo ""
echo "Done! You can now use 'shotgun' or 'shotgun-cli' commands."

if [ "$IS_WSL" = true ]; then
    echo "Ensure your npm global bin directory is in your WSL PATH."
    echo "Typically: export PATH=$(npm prefix -g)/bin:\$PATH"
fi
