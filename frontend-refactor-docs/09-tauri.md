## 12. Préparation pour l'application standalone

### 12.1 Configuration Tauri

```
// src-tauri/Cargo.toml
[package]
name = "talas"
version = "1.0.0"
description = "Talas - Plateforme audio collaborative"
authors = ["Talas Team"]
license = "MIT"
edition = "2021"

[build-dependencies]
tauri-build = { version = "1.5", features = [] }

[dependencies]
tauri = { version = "1.5", features = [
  "shell-open",
  "window-all",
  "path-all",
  "fs-all",
  "dialog-all",
  "clipboard-all",
  "notification-all",
  "global-shortcut-all",
  "os-all",
  "updater",
  "system-tray"
] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tokio = { version = "1", features = ["full"] }
reqwest = { version = "0.11", features = ["json", "stream"] }
anyhow = "1.0"
thiserror = "1.0"
log = "0.4"
env_logger = "0.10"

[features]
default = ["custom-protocol"]
custom-protocol = ["tauri/custom-protocol"]

// src-tauri/src/main.rs
#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use tauri::{
    CustomMenuItem, Manager, SystemTray, SystemTrayEvent, SystemTrayMenu, SystemTrayMenuItem,
};
use tauri::api::notification::Notification;

mod commands;
mod utils;

use commands::*;

fn main() {
    env_logger::init();

    let quit = CustomMenuItem::new("quit".to_string(), "Quitter");
    let hide = CustomMenuItem::new("hide".to_string(), "Masquer");
    let show = CustomMenuItem::new("show".to_string(), "Afficher");
    
    let tray_menu = SystemTrayMenu::new()
        .add_item(show)
        .add_item(hide)
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(quit);

    let system_tray = SystemTray::new().with_menu(tray_menu);

    tauri::Builder::default()
        .system_tray(system_tray)
        .on_system_tray_event(|app, event| match event {
            SystemTrayEvent::LeftClick {
                position: _,
                size: _,
                ..
            } => {
                let window = app.get_window("main").unwrap();
                window.show().unwrap();
                window.set_focus().unwrap();
            }
            SystemTrayEvent::MenuItemClick { id, .. } => match id.as_str() {
                "quit" => {
                    std::process::exit(0);
                }
                "hide" => {
                    let window = app.get_window("main").unwrap();
                    window.hide().unwrap();
                }
                "show" => {
                    let window = app.get_window("main").unwrap();
                    window.show().unwrap();
                }
                _ => {}
            },
            _ => {}
        })
        .invoke_handler(tauri::generate_handler![
            download_file,
            upload_file,
            get_system_info,
            check_for_updates,
            send_notification,
            open_external,
            get_audio_devices,
            save_user_preferences,
            load_user_preferences,
        ])
        .setup(|app| {
            let window = app.get_window("main").unwrap();
            
            // Set window properties
            window.set_title("Talas").unwrap();
            window.set_resizable(true).unwrap();
            window.set_fullscreenable(true).unwrap();
            
            // Center window on screen
            window.center().unwrap();
            
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

// src-tauri/src/commands.rs
use std::path::PathBuf;
use tauri::api::notification::Notification;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct SystemInfo {
    pub os: String,
    pub arch: String,
    pub version: String,
}

#[tauri::command]
pub async fn download_file(url: String, path: String) -> Result<String, String> {
    let response = reqwest::get(&url)
        .await
        .map_err(|e| e.to_string())?;
    
    let bytes = response.bytes()
        .await
        .map_err(|e| e.to_string())?;
    
    std::fs::write(&path, bytes)
        .map_err(|e| e.to_string())?;
    
    Ok(format!("File downloaded to: {}", path))
}

#[tauri::command]
pub async fn upload_file(path: String, endpoint: String) -> Result<String, String> {
    let file = std::fs::read(&path)
        .map_err(|e| e.to_string())?;
    
    let client = reqwest::Client::new();
    let response = client
        .post(&endpoint)
        .body(file)
        .send()
        .await
        .map_err(|e| e.to_string())?;
    
    let text = response.text()
        .await
        .map_err(|e| e.to_string())?;
    
    Ok(text)
}

#[tauri::command]
pub fn get_system_info() -> SystemInfo {
    SystemInfo {
        os: std::env::consts::OS.to_string(),
        arch: std::env::consts::ARCH.to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
    }
}

#[tauri::command]
pub async fn check_for_updates() -> Result<bool, String> {
    // Implement update checking logic
    Ok(false)
}

#[tauri::command]
pub fn send_notification(title: String, body: String) -> Result<(), String> {
    Notification::new("com.talas.app")
        .title(&title)
        .body(&body)
        .show()
        .map_err(|e| e.to_string())?;
    
    Ok(())
}

#[tauri::command]
pub async fn open_external(url: String) -> Result<(), String> {
    tauri::api::shell::open(
        &tauri::api::shell::Scope::default(),
        &url,
        None
    ).map_err(|e| e.to_string())
}

#[tauri::command]
pub fn get_audio_devices() -> Result<Vec<String>, String> {
    // This would integrate with system audio APIs
    // For now, return mock data
    Ok(vec![
        "Default Audio Device".to_string(),
        "USB Audio Interface".to_string(),
    ])
}

#[tauri::command]
pub fn save_user_preferences(prefs: serde_json::Value) -> Result<(), String> {
    let config_dir = tauri::api::path::config_dir()
        .ok_or_else(|| "Could not find config directory".to_string())?;
    
    let app_config_dir = config_dir.join("talas");
    std::fs::create_dir_all(&app_config_dir)
        .map_err(|e| e.to_string())?;
    
    let prefs_file = app_config_dir.join("preferences.json");
    std::fs::write(prefs_file, serde_json::to_string_pretty(&prefs).unwrap())
        .map_err(|e| e.to_string())?;
    
    Ok(())
}

#[tauri::command]
pub fn load_user_preferences() -> Result<serde_json::Value, String> {
    let config_dir = tauri::api::path::config_dir()
        .ok_or_else(|| "Could not find config directory".to_string())?;
    
    let prefs_file = config_dir.join("talas").join("preferences.json");
    
    if !prefs_file.exists() {
        return Ok(serde_json::json!({}));
    }
    
    let content = std::fs::read_to_string(prefs_file)
        .map_err(|e| e.to_string())?;
    
    let prefs: serde_json::Value = serde_json::from_str(&content)
        .map_err(|e| e.to_string())?;
    
    Ok(prefs)
}

// src-tauri/tauri.conf.json
{
  "build": {
    "beforeDevCommand": "npm run dev",
    "beforeBuildCommand": "npm run build",
    "devPath": "http://localhost:3000",
    "distDir": "../dist",
    "withGlobalTauri": false
  },
  "package": {
    "productName": "Talas",
    "version": "1.0.0"
  },
  "tauri": {
    "allowlist": {
      "all": false,
      "shell": {
        "open": true
      },
      "window": {
        "all": true
      },
      "path": {
        "all": true
      },
      "fs": {
        "all": true,
        "scope": ["$APP", "$RESOURCE", "$DOWNLOAD", "$DOCUMENT"]
      },
      "dialog": {
        "all": true
      },
      "clipboard": {
        "all": true
      },
      "notification": {
        "all": true
      },
      "globalShortcut": {
        "all": true
      },
      "os": {
        "all": true
      }
    },
    "bundle": {
      "active": true,
      "targets": "all",
      "identifier": "com.talas.app",
      "icon": [
        "icons/32x32.png",
        "icons/128x128.png",
        "icons/128x128@2x.png",
        "icons/icon.icns",
        "icons/icon.ico"
      ]
    },
    "security": {
      "csp": "default-src 'self'; img-src 'self' data: https:; script-src 'self'"
    },
    "updater": {
      "active": true,
      "endpoints": [
        "https://api.talas.app/updates/{{target}}/{{current_version}}"
      ],
      "dialog": true,
      "pubkey": "YOUR_PUBLIC_KEY_HERE"
    },
    "windows": [
      {
        "title": "Talas",
        "width": 1200,
        "height": 800,
        "minWidth": 800,
        "minHeight": 600,
        "resizable": true,
        "fullscreen": false,
        "center": true,
        "transparent": false,
        "decorations": true,
        "alwaysOnTop": false,
        "skipTaskbar": false
      }
    ],
    "systemTray": {
      "iconPath": "icons/icon.png",
      "iconAsTemplate": true
    }
  }
}
```

### 12.2 Intégration Frontend avec Tauri

```
// src/shared/services/tauriService.ts
import { invoke } from '@tauri-apps/api/tauri';
import { open, save } from '@tauri-apps/api/dialog';
import { readBinaryFile, writeBinaryFile } from '@tauri-apps/api/fs';
import { sendNotification } from '@tauri-apps/api/notification';
import { appWindow } from '@tauri-apps/api/window';
import { listen } from '@tauri-apps/api/event';

// Type-safe commands
interface TauriCommands {
  download_file: (args: { url: string; path: string }) => Promise<string>;
  upload_file: (args: { path: string; endpoint: string }) => Promise<string>;
  get_system_info: () => Promise<SystemInfo>;
  check_for_updates: () => Promise<boolean>;
  send_notification: (args: { title: string; body: string }) => Promise<void>;
  open_external: (args: { url: string }) => Promise<void>;
  get_audio_devices: () => Promise<string[]>;
  save_user_preferences: (args: { prefs: any }) => Promise<void>;
  load_user_preferences: () => Promise<any>;
}

interface SystemInfo {
  os: string;
  arch: string;
  version: string;
}

class TauriService {
  private isDesktop: boolean;

  constructor() {
    this.isDesktop = window.__TAURI__ !== undefined;
  }

  // Check if running in Tauri
  isTauri(): boolean {
    return this.isDesktop;
  }

  // File operations
  async selectFile(filters?: { name: string; extensions: string[] }[]): Promise<string | null> {
    if (!this.isTauri()) return null;
    
    const selected = await open({
      multiple: false,
      filters: filters || [],
    });
    
    return selected as string | null;
  }

  async selectFiles(filters?: { name: string; extensions: string[] }[]): Promise<string[] | null> {
    if (!this.isTauri()) return null;
    
    const selected = await open({
      multiple: true,
      filters: filters || [],
    });
    
    return selected as string[] | null;
  }

  async saveFile(defaultPath?: string, filters?: { name: string; extensions: string[] }[]): Promise<string | null> {
    if (!this.isTauri()) return null;
    
    const filePath = await save({
      defaultPath,
      filters: filters || [],
    });
    
    return filePath;
  }

  async readFile(path: string): Promise<Uint8Array> {
    if (!this.isTauri()) throw new Error('Not in Tauri environment');
    return await readBinaryFile(path);
  }

  async writeFile(path: string, contents: Uint8Array): Promise<void> {
    if (!this.isTauri()) throw new Error('Not in Tauri environment');
    await writeBinaryFile(path, contents);
  }

  // System operations
  async getSystemInfo(): Promise<SystemInfo> {
    if (!this.isTauri()) {
      return {
        os: navigator.platform,
        arch: 'unknown',
        version: '1.0.0',
      };
    }
    
    return await invoke<SystemInfo>('get_system_info');
  }

  async checkForUpdates(): Promise<boolean> {
    if (!this.isTauri()) return false;
    return await invoke<boolean>('check_for_updates');
  }

  // Notifications
  async notify(title: string, body: string): Promise<void> {
    if (!this.isTauri()) {
      // Fallback to browser notifications
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(title, { body });
      }
      return;
    }
    
    await invoke('send_notification', { title, body });
  }

  // External links
  async openExternal(url: string): Promise<void> {
    if (!this.isTauri()) {
      window.open(url, '_blank');
      return;
    }
    
    await invoke('open_external', { url });
  }

  // Audio devices
  async getAudioDevices(): Promise<string[]> {
    if (!this.isTauri()) {
      // Fallback to Web Audio API
      try {
        const devices = await navigator.mediaDevices.enumerateDevices();
        return devices
          .filter(device => device.kind === 'audiooutput')
          .map(device => device.label || 'Unknown Device');
      } catch {
        return ['Default Audio Device'];
      }
    }
    
    return await invoke<string[]>('get_audio_devices');
  }

  // Preferences
  async savePreferences(prefs: any): Promise<void> {
    if (!this.isTauri()) {
      localStorage.setItem('talas_preferences', JSON.stringify(prefs));
      return;
    }
    
    await invoke('save_user_preferences', { prefs });
  }

  async loadPreferences(): Promise<any> {
    if (!this.isTauri()) {
      const stored = localStorage.getItem('talas_preferences');
      return stored ? JSON.parse(stored) : {};
    }
    
    return await invoke('load_user_preferences');
  }

  // Window operations
  async minimizeWindow(): Promise<void> {
    if (!this.isTauri()) return;
    await appWindow.minimize();
  }

  async maximizeWindow(): Promise<void> {
    if (!this.isTauri()) return;
    await appWindow.maximize();
  }

  async closeWindow(): Promise<void> {
    if (!this.isTauri()) return;
    await appWindow.close();
  }

  async setFullscreen(fullscreen: boolean): Promise<void> {
    if (!this.isTauri()) return;
    await appWindow.setFullscreen(fullscreen);
  }

  // Event listeners
  async onFileDropped(handler: (paths: string[]) => void): Promise<() => void> {
    if (!this.isTauri()) return () => {};
    
    const unlisten = await listen<{ paths: string[] }>('tauri://file-drop', (event) => {
      handler(event.payload.paths);
    });
    
    return unlisten;
  }
}

export const tauriService = new TauriService();

// src/app/providers/TauriProvider.tsx
import { createContext, useContext, useEffect, useState } from 'react';
import { tauriService } from '@/shared/services/tauriService';

interface TauriContextValue {
  isTauri: boolean;
  systemInfo: SystemInfo | null;
}

const TauriContext = createContext<TauriContextValue>({
  isTauri: false,
  systemInfo: null,
});

export const useTauri = () => useContext(TauriContext);

export const TauriProvider = ({ children }: { children: React.ReactNode }) => {
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const isTauri = tauriService.isTauri();

  useEffect(() => {
    if (isTauri) {
      tauriService.getSystemInfo().then(setSystemInfo);
      
      // Setup file drop handler
      tauriService.onFileDropped((paths) => {
        console.log('Files dropped:', paths);
        // Handle file drops globally
      });
    }
  }, [isTauri]);

  return (
    <TauriContext.Provider value={{ isTauri, systemInfo }}>
      {children}
    </TauriContext.Provider>
  );
};

// src/features/tracks/hooks/useAudioUpload.ts (Tauri integration)
import { useState } from 'react';
import { tauriService } from '@/shared/services/tauriService';
import { audioService } from '../services/audioService';

export const useAudioUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const selectAndUploadAudio = async (metadata: any) => {
    try {
      setUploading(true);
      
      // Use Tauri file picker if available
      let file: File;
      
      if (tauriService.isTauri()) {
        const path = await tauriService.selectFile([
          { name: 'Audio Files', extensions: ['mp3', 'wav', 'flac', 'ogg', 'm4a'] }
        ]);
        
        if (!path) {
          setUploading(false);
          return null;
        }
        
        // Read file from disk
        const fileData = await tauriService.readFile(path);
        const fileName = path.split('/').pop() || 'audio.mp3';
        file = new File([fileData], fileName);
      } else {
        // Fallback to browser file input
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'audio/*';
        
        const filePromise = new Promise<File | null>((resolve) => {
          input.onchange = (e) => {
            const target = e.target as HTMLInputElement;
            resolve(target.files?.[0] || null);
          };
          input.oncancel = () => resolve(null);
        });
        
        input.click();
        const selectedFile = await filePromise;
        
        if (!selectedFile) {
          setUploading(false);
          return null;
        }
        
        file = selectedFile;
      }
      
      // Upload to server
      const result = await audioService.uploadTrack(
        file,
        metadata,
        (progress) => setProgress(progress)
      );
      
      return result;
    } catch (error) {
      console.error('Upload error:', error);
      throw error;
    } finally {
      setUploading(false);
      setProgress(0);
    }
  };

  return {
    selectAndUploadAudio,
    uploading,
    progress,
  };
};

// package.json scripts update
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "tauri": "tauri",
    "tauri:dev": "tauri dev",
    "tauri:build": "tauri build",
    "tauri:build:windows": "tauri build --target x86_64-pc-windows-msvc",
    "tauri:build:mac": "tauri build --target universal-apple-darwin",
    "tauri:build:linux": "tauri build --target x86_64-unknown-linux-gnu",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:coverage": "vitest --coverage",
    "lint": "eslint src --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "format": "prettier --write \"src/**/*.{ts,tsx,css,md}\"",
    "type-check": "tsc --noEmit",
    "analyze": "vite-bundle-visualizer"
  }
}
```
