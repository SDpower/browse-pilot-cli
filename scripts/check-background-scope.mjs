#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import vm from 'node:vm';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const projectDir = path.dirname(scriptDir);
const distDir = path.join(projectDir, 'dist');

function read(file) {
  return fs.readFileSync(file, 'utf8');
}

function checkClassicBackground(browserName) {
  const browserDir = path.join(distDir, browserName);
  const manifest = JSON.parse(read(path.join(browserDir, 'manifest.json')));
  const scripts = manifest.background?.scripts;
  if (!Array.isArray(scripts) || scripts.length === 0) {
    throw new Error(`${browserName} manifest 未定義 background.scripts`);
  }

  const source = scripts
    .map((relativePath) => `\n// ${relativePath}\n${read(path.join(browserDir, relativePath))}`)
    .join('\n');
  new vm.Script(source, { filename: `${browserName}-background.js` });

  return { browserDir, scripts };
}

function checkServiceWorker(browserName) {
  const browserDir = path.join(distDir, browserName);
  const manifest = JSON.parse(read(path.join(browserDir, 'manifest.json')));
  const relativePath = manifest.background?.service_worker;
  if (!relativePath) {
    throw new Error(`${browserName} manifest 未定義 background.service_worker`);
  }

  new vm.Script(read(path.join(browserDir, relativePath)), {
    filename: `${browserName}-${relativePath}`,
  });
}

async function checkSharedHandlerBindings(browserDir, scripts) {
  const sharedSource = scripts
    .filter((relativePath) => relativePath.startsWith('shared/'))
    .map((relativePath) => read(path.join(browserDir, relativePath)))
    .join('\n');
  const exposeBindings = `
    globalThis.__browsePilotBindings = {
      Router,
      NavigationHandler,
      TabsHandler,
      CookiesHandler,
      ScreenshotHandler,
      SessionHandler,
    };
  `;
  const context = vm.createContext({ browser: {} });
  new vm.Script(sharedSource + exposeBindings, {
    filename: 'shared-background-bindings.js',
  }).runInContext(context);

  const bindings = context.__browsePilotBindings;
  for (const handlerName of [
    'NavigationHandler',
    'TabsHandler',
    'CookiesHandler',
    'ScreenshotHandler',
    'SessionHandler',
  ]) {
    if (!bindings?.[handlerName]) {
      throw new Error(`${handlerName} 未公開至背景腳本作用域`);
    }
  }

  bindings.Router.register('__error_data_check__', async () => {
    throw { code: -32602, message: '測試錯誤', data: false };
  });
  const response = await bindings.Router.dispatch({
    id: 'scope-check',
    method: '__error_data_check__',
  });
  if (response.error?.data !== false) {
    throw new Error('Router 未完整保留 RPCError data');
  }
}

async function main() {
  const firefox = checkClassicBackground('firefox');
  checkServiceWorker('chrome');
  checkServiceWorker('edge');
  await checkSharedHandlerBindings(firefox.browserDir, firefox.scripts);
  console.log('✓ Firefox、Chrome、Edge 背景腳本同作用域語法檢查通過');
}

main().catch((error) => {
  console.error(`背景腳本同作用域語法檢查失敗：${error.message}`);
  process.exitCode = 1;
});
