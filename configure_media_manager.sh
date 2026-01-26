#!/usr/bin/env bash
#
# Media Library Manager Configuration Script
#
# This script sets environment variables for the media library manager.
# It supports both temporary (current session) and permanent configuration.
#
# Usage:
#   chmod +x configure_media_manager.sh
#
#   # For temporary session variables (must source, not execute):
#   source ./configure_media_manager.sh
#
#   # For permanent storage (can execute normally):
#   ./configure_media_manager.sh
#
# Run the Go application with all configuration flags:
#   go run . \
#     -torrent-path="/opt/qbit/downloads" \
#     -movie-path="/opt/jellyfin/media/movies" \
#     -show-path="/opt/jellyfin/media/shows" \
#     -manager-path="/opt/media_manager" \
#     -dry-run
#
# Or with environment variables set:
#   export ENIACORE_TORRENT_PATH="/opt/qbit/downloads"
#   export ENIACORE_MOVIE_PATH="/opt/jellyfin/media/movies"
#   export ENIACORE_SHOW_PATH="/opt/jellyfin/media/shows"
#   export ENIACORE_MANAGER_PATH="/opt/media_manager"
#   export ENIACORE_DRY_RUN="true"
#   go run .


#!/usr/bin/env bash
#
# Media Library Manager Configuration
#
# Source this script to configure environment variables:
#   source ./configure_media_manager.sh

# Function to prompt with default value (works in bash and zsh)
prompt_with_default() {
    local var_name="$1"
    local prompt_text="$2"
    local default_val="$3"
    local input

    if [ -n "$ZSH_VERSION" ]; then
        vared -p "$prompt_text [$default_val]: " -c input
    else
        read -e -p "$prompt_text [$default_val]: " -i "$default_val" input
    fi

    eval "$var_name=\"${input:-$default_val}\""
}

prompt_with_default ENIACORE_TORRENT_PATH "Torrent path" "/opt/qbit/downloads"
prompt_with_default ENIACORE_MOVIE_PATH "Movie path" "/opt/jellyfin/media/movies"
prompt_with_default ENIACORE_SHOW_PATH "Show path" "/opt/jellyfin/media/shows"
prompt_with_default ENIACORE_MANAGER_PATH "Manager path" "/opt/media_manager"
prompt_with_default ENIACORE_DRY_RUN "Dry run (true/false)" "true"

export ENIACORE_TORRENT_PATH
export ENIACORE_MOVIE_PATH
export ENIACORE_SHOW_PATH
export ENIACORE_MANAGER_PATH
export ENIACORE_DRY_RUN

echo "Media Manager environment configured"
