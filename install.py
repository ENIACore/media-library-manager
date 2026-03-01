#!/usr/bin/env python3
"""
Media Library Manager Installation Script

This script downloads and installs the latest mlm binary for your system.
Supports Linux and macOS (both amd64 and arm64 architectures).

Usage:
    curl -fsSL https://raw.githubusercontent.com/ENIACore/media-library-manager/main/install.py -o /tmp/mlm-install.py && sudo python3 /tmp/mlm-install.py; rm -f /tmp/mlm-install.py
"""

import os
import sys
import platform
import urllib.request
import json
import stat
import shutil
from pathlib import Path


# Configuration
GITHUB_REPO = "ENIACore/media-library-manager"
BINARY_NAME = "mlm"
INSTALL_DIR = "/usr/local/bin"
GITHUB_API_URL = f"https://api.github.com/repos/{GITHUB_REPO}/releases/latest"


class Colors:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'


def print_header(message):
    """Print a colored header message"""
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'=' * 60}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{message.center(60)}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'=' * 60}{Colors.ENDC}\n")


def print_success(message):
    """Print a success message"""
    print(f"{Colors.OKGREEN}✓{Colors.ENDC} {message}")


def print_error(message):
    """Print an error message"""
    print(f"{Colors.FAIL}✗{Colors.ENDC} {message}", file=sys.stderr)


def print_info(message):
    """Print an info message"""
    print(f"{Colors.OKCYAN}ℹ{Colors.ENDC} {message}")


def print_warning(message):
    """Print a warning message"""
    print(f"{Colors.WARNING}⚠{Colors.ENDC} {message}")


def detect_platform():
    """
    Detect the current platform and architecture.

    Returns:
        tuple: (os_name, architecture) e.g., ('linux', 'amd64')
    """
    system = platform.system().lower()
    machine = platform.machine().lower()

    # Normalize OS name
    if system == "darwin":
        os_name = "darwin"
    elif system == "linux":
        os_name = "linux"
    else:
        print_error(f"Unsupported operating system: {system}")
        print_info("Supported platforms: Linux, macOS")
        sys.exit(1)

    # Normalize architecture
    if machine in ["x86_64", "amd64"]:
        arch = "amd64"
    elif machine in ["aarch64", "arm64"]:
        arch = "arm64"
    else:
        print_error(f"Unsupported architecture: {machine}")
        print_info("Supported architectures: amd64, arm64")
        sys.exit(1)

    return os_name, arch


def check_root():
    """Check if script is running with root privileges"""
    if os.geteuid() != 0:
        print_error("This script must be run with sudo privileges")
        print_info(f"Please run: sudo python3 {sys.argv[0]}")
        sys.exit(1)


def get_latest_release():
    """
    Fetch the latest release information from GitHub.

    Returns:
        dict: Release information including tag name and assets
    """
    try:
        print_info("Fetching latest release information...")

        req = urllib.request.Request(GITHUB_API_URL)
        req.add_header('Accept', 'application/vnd.github.v3+json')

        with urllib.request.urlopen(req, timeout=10) as response:
            data = json.loads(response.read().decode())
            print_success(f"Found version: {data['tag_name']}")
            return data

    except urllib.error.HTTPError as e:
        if e.code == 404:
            print_error("No releases found for this repository")
            print_info("Please ensure releases are published on GitHub")
        else:
            print_error(f"HTTP error while fetching release: {e.code}")
        sys.exit(1)
    except urllib.error.URLError as e:
        print_error(f"Network error: {e.reason}")
        print_info("Please check your internet connection")
        sys.exit(1)
    except Exception as e:
        print_error(f"Failed to fetch release information: {e}")
        sys.exit(1)


def download_binary(download_url, destination):
    """
    Download binary from URL to destination.

    Args:
        download_url: URL to download from
        destination: Path to save the file
    """
    try:
        print_info(f"Downloading {BINARY_NAME}...")

        # Download with progress indication
        def report_progress(block_num, block_size, total_size):
            downloaded = block_num * block_size
            if total_size > 0:
                percent = min(100, (downloaded * 100) // total_size)
                bar_length = 40
                filled = int(bar_length * percent / 100)
                bar = '█' * filled + '░' * (bar_length - filled)
                print(f"\r{Colors.OKCYAN}↓{Colors.ENDC} Progress: [{bar}] {percent}%", end='', flush=True)

        urllib.request.urlretrieve(download_url, destination, reporthook=report_progress)
        print()  # New line after progress bar
        print_success(f"Downloaded to {destination}")

    except Exception as e:
        print_error(f"Failed to download binary: {e}")
        sys.exit(1)


def install_binary(source_path, install_path):
    """
    Install binary to the system path.

    Args:
        source_path: Path to downloaded binary
        install_path: Path to install binary to
    """
    try:
        print_info(f"Installing {BINARY_NAME} to {install_path}...")

        # Copy binary
        shutil.copy2(source_path, install_path)

        # Make executable
        st = os.stat(install_path)
        os.chmod(install_path, st.st_mode | stat.S_IEXEC | stat.S_IXGRP | stat.S_IXOTH)

        print_success(f"Installed to {install_path}")

    except Exception as e:
        print_error(f"Failed to install binary: {e}")
        sys.exit(1)


def main():
    """Main installation process"""
    print_header("Media Library Manager Installation")

    # Check for root privileges
    check_root()

    # Detect platform
    os_name, arch = detect_platform()
    print_info(f"Detected platform: {os_name}/{arch}")

    # Get latest release
    release = get_latest_release()
    version = release['tag_name']

    # Find matching asset
    binary_filename = f"{BINARY_NAME}-{os_name}-{arch}"
    download_url = None

    for asset in release.get('assets', []):
        if asset['name'] == binary_filename:
            download_url = asset['browser_download_url']
            break

    if not download_url:
        print_error(f"No binary found for {os_name}/{arch}")
        print_info("Available assets:")
        for asset in release.get('assets', []):
            print(f"  - {asset['name']}")
        sys.exit(1)

    # Download binary
    temp_dir = Path("/tmp")
    temp_binary = temp_dir / binary_filename
    download_binary(download_url, temp_binary)

    # Install binary
    install_path = Path(INSTALL_DIR) / BINARY_NAME
    install_binary(temp_binary, install_path)

    # Clean up
    temp_binary.unlink()

    # Print success message
    print_header("Installation Complete!")
    print_success(f"{BINARY_NAME} {version} has been installed successfully")
    print_info(f"Binary location: {install_path}")
    print()
    print(f"{Colors.BOLD}Next steps:{Colors.ENDC}")
    print(f"  1. Run '{Colors.OKGREEN}{BINARY_NAME} -help{Colors.ENDC}' to see available options")
    print(f"  2. Use '{Colors.OKGREEN}./configure_media_manager.sh{Colors.ENDC}' to set up configuration")
    print(f"  3. Run '{Colors.OKGREEN}{BINARY_NAME}{Colors.ENDC}' to start managing your media library")
    print()
    print(f"{Colors.OKCYAN}Documentation:{Colors.ENDC} https://github.com/{GITHUB_REPO}")
    print()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print()
        print_warning("Installation cancelled by user")
        sys.exit(130)
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        sys.exit(1)
