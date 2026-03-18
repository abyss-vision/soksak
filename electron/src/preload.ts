import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('electronAPI', {
  getServerUrl: (): Promise<string> =>
    ipcRenderer.invoke('getServerUrl'),

  getAppInfo: (): Promise<{ version: string; platform: string }> =>
    ipcRenderer.invoke('getAppInfo'),

  onNotification: (
    callback: (payload: { title: string; body: string }) => void
  ): (() => void) => {
    const listener = (
      _event: Electron.IpcRendererEvent,
      payload: { title: string; body: string }
    ) => callback(payload);
    ipcRenderer.on('notification', listener);
    return () => ipcRenderer.removeListener('notification', listener);
  },
});
