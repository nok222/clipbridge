# cp-server

A lightweight, silent Go HTTP server that receives clipboard text from your iOS device (via iOS Shortcuts) and sets it directly to your Windows system clipboard, allowing you to immediately `Ctrl + V` on your PC.

---

## Quick Reference

- **Port**: `8765`
- **Default Endpoints**:
  - `POST http://<YOUR_PC_IP>:8765/clipboard`
  - `GET http://<YOUR_PC_IP>:8765/clipboard`
  - `GET http://<YOUR_PC_IP>:8765/health`

### Local IP Addresses
Check your Wi-Fi IP address on Windows (`ipconfig` or PowerShell):
- Full endpoint for your phone: `http://<YOUR_PC_IP>:8765/clipboard`

---

## iOS Shortcut Setup

1. Open the **Shortcuts** app on iOS.
2. Create a new Shortcut (e.g. named **"Send to PC"**).
3. Add actions:
   - **Get Clipboard** (Action 1)
   - **Get Contents of URL** (Action 2):
     - **URL**: `http://<YOUR_PC_IP>:8765/clipboard`
     - **Method**: `POST`
     - **Headers**:
       - `Content-Type`: `application/json`
     - **Request Body**: `JSON`
       - Key: `text`
       - Type: `Text`
       - Value: Tap and select **Clipboard**
4. Save the Shortcut.
5. (Optional) Add it to Back Tap (Settings > Accessibility > Touch > Back Tap) or the Action Button / Share Sheet / Home Screen widget for instant one-tap sync!

---

## Running the Server

### Option A: Run manually in console
```powershell
.\cp-server.exe
```
Or with custom port:
```powershell
.\cp-server.exe -port 8765
```

### Option B: Auto-start silently on Windows Login
Run the included PowerShell installer:
```powershell
powershell -ExecutionPolicy Bypass -File .\install-startup.ps1
```
This builds `cp-server-bg.exe` (headless, no cmd window) and adds a shortcut to Windows Startup (`shell:startup`).

To uninstall / stop autostart:
```powershell
powershell -ExecutionPolicy Bypass -File .\uninstall-startup.ps1
```

---

## Testing from Terminal / PowerShell

### Send text to PC clipboard:
```powershell
Invoke-RestMethod -Uri "http://localhost:8765/clipboard" -Method Post -ContentType "application/json" -Body '{"text": "Copied from test!"}'
```

### Read PC clipboard:
```powershell
Invoke-RestMethod -Uri "http://localhost:8765/clipboard" -Method Get
```
