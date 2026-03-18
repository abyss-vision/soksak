import { ipcMain, shell, Notification, BrowserWindow } from 'electron';
import { app } from 'electron';
import { ServerManager } from './server-manager';

export function registerIpcHandlers(serverManager: ServerManager): void {
  ipcMain.handle('getServerUrl', () => serverManager.getServerUrl());

  ipcMain.handle('getAppInfo', () => ({
    version: app.getVersion(),
    platform: process.platform,
  }));

  ipcMain.handle('showNotification', (_event, title: string, body: string) => {
    const notification = new Notification({ title, body });
    notification.show();

    const win = BrowserWindow.getFocusedWindow();
    if (win) {
      win.webContents.send('notification', { title, body });
    }
  });

  ipcMain.handle('openExternal', async (_event, url: string) => {
    const parsedUrl = new URL(url);
    const allowedProtocols = ['https:', 'http:'];
    if (!allowedProtocols.includes(parsedUrl.protocol)) {
      throw new Error(`Protocol not allowed: ${parsedUrl.protocol}`);
    }
    await shell.openExternal(url);
  });
}
