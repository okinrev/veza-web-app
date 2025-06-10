#!/bin/bash

SESSION_NAME="veza-test-session"
PROJECT_ROOT=~/Documents/TG__Talas_Group/veza-wep-app

# Kill existing session if it exists
tmux has-session -t $SESSION_NAME 2>/dev/null
if [ $? -eq 0 ]; then
  tmux kill-session -t $SESSION_NAME
fi

# Fonction pour charger .env si présent et lancer la commande dans bash
run_with_env() {
  local dir="$1"
  local cmd="$2"
  # On utilise bash -c avec un one-liner pour charger .env si présent, puis exécuter la commande
  echo "cd $dir; if [ -f .env ]; then export \$(grep -v '^#' .env | xargs); fi; $cmd; exec bash"
}

start_container_if_not_running() {
  local container_name="$1"
  if incus info "$container_name" &>/dev/null; then
    local status
    status=$(incus info "$container_name" | grep "^Status:" | awk '{print $2}')
    if [ "$status" != "Running" ]; then
      echo "[+] Starting container $container_name..."
      incus start "$container_name" 2>/dev/null || {
        echo "[!] Warning: Failed to start $container_name or already running."
      }
    else
      echo "[*] Container $container_name is already running."
    fi
  else
    echo "[!] Container $container_name does not exist."
  fi
}



# Create tmux session and windows with the right commands

# Main backend go
tmux new-session -d -s $SESSION_NAME -n "main" "$(run_with_env ~/Documents/TG__Talas_Group/veza-web-app/backend 'go run main.go')"

# chat_server rust
tmux new-window -t $SESSION_NAME:1 -n "chat" "$(run_with_env ~/Documents/TG__Talas_Group/veza-web-app/backend/modules/chat_server 'cargo run')"

# stream_server rust
tmux new-window -t $SESSION_NAME:2 -n "stream" "$(run_with_env ~/Documents/TG__Talas_Group/veza-web-app/backend/modules/stream_server 'cargo run')"

start_container_if_not_running veza-pg-database

tmux new-window -t $SESSION_NAME:3 -n "db" "$(run_with_env ~/Documents/TG__Talas_Group/veza-web-app 'incus shell veza-pg-database')"

# Open a new gnome-terminal and attach tmux session
gnome-terminal -- bash -c "tmux attach -t $SESSION_NAME"

echo "[+] Opening VS Code..."
code "$PROJECT_ROOT"

echo "[+] Opening Firefox on localhost:44103..."
firefox http://localhost:8080 &


