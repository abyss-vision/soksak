import { app, BrowserWindow } from 'electron';
import * as path from 'path';
import { ServerManager } from './server-manager';
import { registerIpcHandlers } from './ipc';

let mainWindow: BrowserWindow | null = null;
const serverManager = new ServerManager();

function createWindow(): BrowserWindow {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  win.once('ready-to-show', () => win.show());

  win.on('closed', () => {
    mainWindow = null;
  });

  return win;
}

app.whenReady().then(async () => {
  registerIpcHandlers(serverManager);

  try {
    await serverManager.start();
  } catch (err) {
    console.error('Failed to start server:', err);
    app.quit();
    return;
  }

  mainWindow = createWindow();
  const serverUrl = serverManager.getServerUrl();
  await mainWindow.loadURL(serverUrl);
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', async () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    mainWindow = createWindow();
    await mainWindow.loadURL(serverManager.getServerUrl());
  }
});

app.on('before-quit', async () => {
  await serverManager.stop();
});
