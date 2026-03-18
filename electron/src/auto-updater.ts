import { autoUpdater } from 'electron-updater';
import { dialog, BrowserWindow } from 'electron';

export function setupAutoUpdater(): void {
  autoUpdater.autoDownload = false;
  autoUpdater.autoInstallOnAppQuit = true;

  autoUpdater.on('update-available', (info) => {
    const win = BrowserWindow.getFocusedWindow();
    dialog
      .showMessageBox(win ?? BrowserWindow.getAllWindows()[0], {
        type: 'info',
        title: 'Update Available',
        message: `Version ${info.version} is available. Download now?`,
        buttons: ['Download', 'Later'],
        defaultId: 0,
        cancelId: 1,
      })
      .then(({ response }) => {
        if (response === 0) {
          autoUpdater.downloadUpdate();
        }
      });
  });

  autoUpdater.on('update-downloaded', () => {
    const win = BrowserWindow.getFocusedWindow();
    dialog
      .showMessageBox(win ?? BrowserWindow.getAllWindows()[0], {
        type: 'info',
        title: 'Update Ready',
        message: 'Update downloaded. Restart to apply the update now?',
        buttons: ['Restart', 'Later'],
        defaultId: 0,
        cancelId: 1,
      })
      .then(({ response }) => {
        if (response === 0) {
          autoUpdater.quitAndInstall();
        }
      });
  });

  autoUpdater.on('error', (err) => {
    console.error('Auto-updater error:', err);
  });

  autoUpdater.checkForUpdates().catch((err) => {
    console.warn('Update check failed:', err);
  });
}
