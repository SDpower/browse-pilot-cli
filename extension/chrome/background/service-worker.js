// Chrome/Edge MV3 Service Worker
// 透過 Native Messaging 與 CLI 端通訊
//
// 注意：MV3 module service worker 無法使用 importScripts()
// Build script 會將 shared/utils/ 與 shared/handlers/ 的程式碼
// concat 到此檔案前方，產生最終的 dist/chrome/background/service-worker.js
//
// 因此此檔案假設以下全域變數已由 build script 注入：
//   Router, NavigationHandler, TabsHandler, CookiesHandler,
//   ScreenshotHandler, SessionHandler, ErrorCodes, createErrorResponse

/* global Router, NavigationHandler, TabsHandler, CookiesHandler,
          ScreenshotHandler, SessionHandler, ErrorCodes, createErrorResponse */

const NM_HOST = 'com.browse_pilot.host';
let nmPort = null;
let isConnected = false;
let reconnectTimer = null;

// 建立 Native Messaging 連線
function connectNative() {
  console.log('[Browse Pilot] 連線至 Native Messaging host:', NM_HOST);

  try {
    nmPort = chrome.runtime.connectNative(NM_HOST);
  } catch (e) {
    console.error('[Browse Pilot] Native Messaging 連線失敗:', e.message);
    scheduleReconnect();
    return;
  }

  isConnected = true;
  chrome.runtime.sendMessage({ type: 'connection_status', connected: true }).catch(() => {});

  nmPort.onMessage.addListener(async (request) => {
    try {
      console.log(`[Browse Pilot] 收到指令: ${request.method}`);
      const response = await Router.dispatch(request);
      nmPort.postMessage(response);
    } catch (e) {
      console.error('[Browse Pilot] 處理指令失敗:', e);
      nmPort.postMessage(createErrorResponse(
        request?.id ?? null,
        ErrorCodes.EXTENSION_ERROR,
        e.message
      ));
    }
  });

  nmPort.onDisconnect.addListener(() => {
    console.log('[Browse Pilot] Native Messaging 已斷線');
    const lastError = chrome.runtime.lastError;
    if (lastError) {
      console.error('[Browse Pilot] 斷線原因:', lastError.message);
    }
    isConnected = false;
    nmPort = null;
    chrome.runtime.sendMessage({ type: 'connection_status', connected: false }).catch(() => {});
    scheduleReconnect();
  });
}

function scheduleReconnect() {
  if (reconnectTimer) clearTimeout(reconnectTimer);
  reconnectTimer = setTimeout(() => {
    console.log('[Browse Pilot] 嘗試重新連線...');
    connectNative();
  }, 3000);
}

// Keepalive 策略：使用 chrome.alarms 防止 service worker 因 idle timeout 而終止
// Chrome MV3 service worker 在閒置 30 秒後會被終止
// 設定約 24 秒的 alarm 定期喚醒，確保連線持續
chrome.alarms.create('keepalive', { periodInMinutes: 0.4 });

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === 'keepalive') {
    // alarm 事件本身即可重置 service worker idle timer
    if (!isConnected && !nmPort) {
      connectNative();
    }
  }
});

// 在 Router 上註冊所有指令 handler
// Background 端直接處理需要 tabs/cookies 等 API 權限的指令
// Content script 相關的指令透過 chrome.tabs.sendMessage 轉發
function registerHandlers() {
  // Background 端直接處理的指令
  Router.register('navigate', (params) => NavigationHandler.navigate(params));
  Router.register('go_back', (params) => NavigationHandler.goBack(params));
  Router.register('go_forward', (params) => NavigationHandler.goForward(params));
  Router.register('reload', (params) => NavigationHandler.reload(params));
  Router.register('get_tabs', () => TabsHandler.getTabs());
  Router.register('switch_tab', (params) => TabsHandler.switchTab(params));
  Router.register('close_tab', (params) => TabsHandler.closeTab(params));
  Router.register('get_cookies', (params) => CookiesHandler.getCookies(params));
  Router.register('set_cookie', (params) => CookiesHandler.setCookie(params));
  Router.register('clear_cookies', (params) => CookiesHandler.clearCookies(params));
  Router.register('screenshot', (params) => ScreenshotHandler.capture(params));
  Router.register('get_status', () => SessionHandler.getStatus());

  // 需要轉發至 content script 的指令
  const contentMethods = [
    'get_state', 'click', 'type_text', 'input_text', 'send_keys',
    'select_option', 'hover', 'dblclick', 'rightclick', 'upload_file',
    'wait_selector', 'wait_text', 'wait_url',
    'get_title', 'get_html', 'get_text', 'get_value',
    'get_attributes', 'get_bbox', 'eval_js', 'scroll'
  ];

  for (const method of contentMethods) {
    Router.register(method, async (params) => {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      if (!tab) throw { code: ErrorCodes.TAB_NOT_FOUND, message: '無活躍分頁' };

      const response = await chrome.tabs.sendMessage(tab.id, { method, params });
      if (response.error) throw response.error;
      return response.result;
    });
  }
}

// 處理來自 popup 的訊息
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'get_connection_status') {
    sendResponse({ connected: isConnected, port: null, browser: 'chrome' });
    return false;
  }
  return false;
});

// 啟動：註冊 handler → 建立 Native Messaging 連線
registerHandlers();
connectNative();
