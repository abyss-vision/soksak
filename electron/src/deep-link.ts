import { app, BrowserWindow } from 'electron';

const PROTOCOL = 'soksak';

export function registerDeepLink(): void {
  if (process.defaultApp) {
    if (process.argv.length >= 2) {
      app.setAsDefaultProtocolClient(PROTOCOL, process.execPath, [
        process.argv[1],
      ]);
    }
  } else {
    app.setAsDefaultProtocolClient(PROTOCOL);
  }

  app.on('open-url', (_event, url) => {
    handleDeepLink(url);
  });

  // Windows: second-instance deep link
  app.on('second-instance', (_event, argv) => {
    const url = argv.find((arg) => arg.startsWith(`${PROTOCOL}://`));
    if (url) handleDeepLink(url);

    const win = BrowserWindow.getAllWindows()[0];
    if (win) {
      if (win.isMinimized()) win.restore();
      win.focus();
    }
  });
}

function handleDeepLink(url: string): void {
  console.log('Deep link received:', url);

  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    console.error('Invalid deep link URL:', url);
    return;
  }

  if (parsed.protocol !== `${PROTOCOL}:`) return;

  const route = parsed.hostname + parsed.pathname;
  routeToView(route, parsed.searchParams);
}

function routeToView(route: string, params: URLSearchParams): void {
  const win = BrowserWindow.getAllWindows()[0];
  if (!win) return;

  const query = params.toString() ? `?${params.toString()}` : '';

  switch (route) {
    case 'issue':
    case 'issue/': {
      const id = params.get('id');
      if (id) win.webContents.send('navigate', `/issues/${id}${query}`);
      break;
    }
    case 'agent':
    case 'agent/': {
      const id = params.get('id');
      if (id) win.webContents.send('navigate', `/agents/${id}${query}`);
      break;
    }
    default:
      win.webContents.send('navigate', `/${route}${query}`);
      break;
  }
}
