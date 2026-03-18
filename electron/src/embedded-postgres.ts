import { app } from 'electron';
import { ChildProcess, spawn, execFile } from 'child_process';
import * as path from 'path';
import * as fs from 'fs';
import * as https from 'https';

const PG_PORT = 5433;
const PG_USER = 'abyss';
const PG_DB = 'abyss_view';

export class EmbeddedPostgres {
  private process: ChildProcess | null = null;
  private pgBinDir: string = '';
  private dataDir: string = '';

  async start(): Promise<string> {
    const userData = app.getPath('userData');
    this.dataDir = path.join(userData, 'pgdata');
    this.pgBinDir = await this.ensureBinary();

    if (!fs.existsSync(path.join(this.dataDir, 'PG_VERSION'))) {
      await this.initDb();
    }

    await this.startProcess();
    await this.ensureDatabase();

    return `postgresql://${PG_USER}@localhost:${PG_PORT}/${PG_DB}`;
  }

  private async ensureBinary(): Promise<string> {
    const userData = app.getPath('userData');
    const pgDir = path.join(userData, 'postgres');

    if (app.isPackaged) {
      return path.join(process.resourcesPath, 'postgres', 'bin');
    }

    if (fs.existsSync(path.join(pgDir, 'bin', 'postgres'))) {
      return path.join(pgDir, 'bin');
    }

    await this.downloadPostgres(pgDir);
    return path.join(pgDir, 'bin');
  }

  private downloadPostgres(targetDir: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const platform = process.platform;
      const arch = process.arch;
      const version = '16.2.0';
      const archStr = arch === 'arm64' ? 'arm_64' : 'x86_64';
      const platformStr =
        platform === 'darwin' ? 'darwin' : platform === 'win32' ? 'windows' : 'linux';
      const url = `https://github.com/theseus-rs/postgresql-binaries/releases/download/${version}/postgresql-${version}-${archStr}-${platformStr}.tar.gz`;

      console.log(`Downloading PostgreSQL from ${url}`);
      fs.mkdirSync(targetDir, { recursive: true });

      const tarPath = path.join(targetDir, 'postgres.tar.gz');
      const file = fs.createWriteStream(tarPath);

      https.get(url, (response) => {
        if (response.statusCode !== 200) {
          reject(new Error(`Download failed: ${response.statusCode}`));
          return;
        }
        response.pipe(file);
        file.on('finish', () => {
          file.close();
          execFile('tar', ['-xzf', tarPath, '-C', targetDir], (err) => {
            if (err) {
              reject(err);
              return;
            }
            fs.unlinkSync(tarPath);
            resolve();
          });
        });
      }).on('error', reject);
    });
  }

  private initDb(): Promise<void> {
    return new Promise((resolve, reject) => {
      fs.mkdirSync(this.dataDir, { recursive: true });
      const initdb = path.join(this.pgBinDir, 'initdb');
      execFile(
        initdb,
        ['-D', this.dataDir, '--username', PG_USER, '--auth', 'trust'],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    });
  }

  private startProcess(): Promise<void> {
    return new Promise((resolve, reject) => {
      const postgres = path.join(this.pgBinDir, 'postgres');
      this.process = spawn(postgres, ['-D', this.dataDir, '-p', String(PG_PORT)], {
        stdio: ['ignore', 'pipe', 'pipe'],
      });

      this.process.stderr?.on('data', (data: Buffer) => {
        const msg = data.toString();
        if (msg.includes('database system is ready')) resolve();
        if (msg.includes('FATAL') || msg.includes('could not bind')) {
          reject(new Error(msg.trim()));
        }
      });

      this.process.on('exit', (code) => {
        if (code !== 0) reject(new Error(`postgres exited with code ${code}`));
      });

      setTimeout(() => reject(new Error('Postgres start timed out')), 15000);
    });
  }

  private ensureDatabase(): Promise<void> {
    return new Promise((resolve, reject) => {
      const createdb = path.join(this.pgBinDir, 'createdb');
      execFile(
        createdb,
        ['-p', String(PG_PORT), '-U', PG_USER, PG_DB],
        (err) => {
          if (!err || (err.message && err.message.includes('already exists'))) {
            resolve();
          } else {
            reject(err);
          }
        }
      );
    });
  }

  async stop(): Promise<void> {
    if (!this.process) return;
    const proc = this.process;
    this.process = null;

    await new Promise<void>((resolve) => {
      const kill = setTimeout(() => {
        proc.kill('SIGKILL');
        resolve();
      }, 5000);

      proc.once('exit', () => {
        clearTimeout(kill);
        resolve();
      });

      proc.kill('SIGTERM');
    });
  }
}
