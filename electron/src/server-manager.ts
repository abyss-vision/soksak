import { ChildProcess, spawn } from 'child_process';
import * as http from 'http';
import * as path from 'path';
import { app } from 'electron';

const HEALTH_CHECK_INTERVAL_MS = 2000;
const HEALTH_CHECK_TIMEOUT_MS = 1000;
const SHUTDOWN_WAIT_MS = 5000;
const MAX_RETRIES = 3;

export class ServerManager {
  private process: ChildProcess | null = null;
  private port: number = 3200;
  private retryCount: number = 0;
  private healthCheckTimer: NodeJS.Timeout | null = null;
  private isShuttingDown: boolean = false;

  getServerUrl(): string {
    return `http://localhost:${this.port}`;
  }

  async start(): Promise<void> {
    const binaryPath = this.getBinaryPath();
    this.spawnProcess(binaryPath);
    await this.waitForHealth();
  }

  private getBinaryPath(): string {
    if (app.isPackaged) {
      return path.join(process.resourcesPath, 'soksak-server');
    }
    return path.join(app.getAppPath(), '..', 'server', 'soksak-server');
  }

  private spawnProcess(binaryPath: string): void {
    this.process = spawn(binaryPath, [], {
      env: {
        ...process.env,
        PORT: String(this.port),
      },
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    this.process.stdout?.on('data', (data: Buffer) => {
      console.log('[server]', data.toString().trim());
    });

    this.process.stderr?.on('data', (data: Buffer) => {
      console.error('[server:err]', data.toString().trim());
    });

    this.process.on('exit', (code, signal) => {
      if (this.isShuttingDown) return;
      console.warn(`Server exited (code=${code}, signal=${signal}), retrying...`);
      this.handleCrash(binaryPath);
    });
  }

  private handleCrash(binaryPath: string): void {
    if (this.retryCount >= MAX_RETRIES) {
      console.error('Server exceeded max retries, giving up.');
      return;
    }
    this.retryCount++;
    console.log(`Restart attempt ${this.retryCount}/${MAX_RETRIES}...`);
    setTimeout(() => this.spawnProcess(binaryPath), 1000 * this.retryCount);
  }

  waitForHealth(): Promise<void> {
    return new Promise((resolve, reject) => {
      const start = Date.now();
      const maxWait = 30000;

      const poll = () => {
        if (Date.now() - start > maxWait) {
          reject(new Error('Server health check timed out'));
          return;
        }
        this.checkHealth()
          .then(() => resolve())
          .catch(() => {
            this.healthCheckTimer = setTimeout(poll, HEALTH_CHECK_INTERVAL_MS);
          });
      };
      poll();
    });
  }

  private checkHealth(): Promise<void> {
    return new Promise((resolve, reject) => {
      const req = http.get(
        `http://localhost:${this.port}/api/health`,
        { timeout: HEALTH_CHECK_TIMEOUT_MS },
        (res) => {
          if (res.statusCode === 200) {
            resolve();
          } else {
            reject(new Error(`Health check returned ${res.statusCode}`));
          }
        }
      );
      req.on('error', reject);
      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Health check request timed out'));
      });
    });
  }

  async stop(): Promise<void> {
    this.isShuttingDown = true;

    if (this.healthCheckTimer) {
      clearTimeout(this.healthCheckTimer);
      this.healthCheckTimer = null;
    }

    if (!this.process) return;

    const proc = this.process;
    this.process = null;

    await new Promise<void>((resolve) => {
      const killTimer = setTimeout(() => {
        proc.kill('SIGKILL');
        resolve();
      }, SHUTDOWN_WAIT_MS);

      proc.once('exit', () => {
        clearTimeout(killTimer);
        resolve();
      });

      proc.kill('SIGTERM');
    });
  }
}
