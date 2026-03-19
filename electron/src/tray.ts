import { app, Menu, MenuItem, Tray, BrowserWindow, nativeImage } from 'electron';
import * as path from 'path';

export class TrayManager {
  private tray: Tray | null = null;
  private agentCount: number = 0;

  create(): void {
    const iconPath = this.getIconPath();
    const icon = nativeImage.createFromPath(iconPath);
    this.tray = new Tray(icon.isEmpty() ? nativeImage.createEmpty() : icon);
    this.tray.setToolTip('SokSak');
    this.updateMenu();

    this.tray.on('double-click', () => this.showMainWindow());
  }

  updateAgentCount(count: number): void {
    this.agentCount = count;
    this.updateMenu();
    if (this.tray) {
      this.tray.setToolTip(
        count > 0
          ? `SokSak — ${count} agent${count !== 1 ? 's' : ''} running`
          : 'SokSak'
      );
    }
  }

  private getIconPath(): string {
    if (app.isPackaged) {
      return path.join(process.resourcesPath, 'tray-icon.png');
    }
    return path.join(app.getAppPath(), 'assets', 'tray-icon.png');
  }

  private showMainWindow(): void {
    const windows = BrowserWindow.getAllWindows();
    if (windows.length === 0) return;
    const win = windows[0];
    if (win.isMinimized()) win.restore();
    win.focus();
  }

  private updateMenu(): void {
    if (!this.tray) return;

    const statusLabel =
      this.agentCount > 0
        ? `${this.agentCount} agent${this.agentCount !== 1 ? 's' : ''} running`
        : 'No agents running';

    const menu = Menu.buildFromTemplate([
      new MenuItem({ label: 'SokSak', enabled: false }),
      new MenuItem({ type: 'separator' }),
      new MenuItem({ label: statusLabel, enabled: false }),
      new MenuItem({ type: 'separator' }),
      new MenuItem({ label: 'Show', click: () => this.showMainWindow() }),
      new MenuItem({ type: 'separator' }),
      new MenuItem({ label: 'Quit', click: () => app.quit() }),
    ]);

    this.tray.setContextMenu(menu);
  }

  destroy(): void {
    if (this.tray) {
      this.tray.destroy();
      this.tray = null;
    }
  }
}
