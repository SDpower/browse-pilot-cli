// Firefox MV2 Persistent Background Script
// WebSocket client 連線至 CLI 端的 WebSocket server

/* global browser, Router, NavigationHandler, TabsHandler, CookiesHandler,
          ScreenshotHandler, SessionHandler, ErrorCodes, createErrorResponse */

const DEFAULT_WS_PORT = 9222;

let ws = null;
let wsPort = DEFAULT_WS_PORT;
let reconnectTimer = null;
let isConnected = false;

// 從 storage 讀取設定
async function loadConfig() {
  try {
    const result = await browser.storage.local.get(['wsPort']);
    if (result.wsPort) wsPort = result.wsPort;
  } catch (e) {
    console.warn('[Browse Pilot] 讀取設定失敗:', e.message);
  }
}

// 建立 WebSocket 連線
function connect() {
  if (ws && ws.readyState === WebSocket.OPEN) return;

  const url = `ws://127.0.0.1:${wsPort}`;
  console.log(`[Browse Pilot] 連線至 ${url}`);

  try {
    ws = new WebSocket(url);
  } catch (e) {
    console.error('[Browse Pilot] WebSocket 建立失敗:', e.message);
    scheduleReconnect();
    return;
  }

  ws.onopen = () => {
    console.log('[Browse Pilot] WebSocket 已連線');
    isConnected = true;
    clearReconnectTimer();
    // 通知 popup 連線狀態
    browser.runtime.sendMessage({ type: 'connection_status', connected: true }).catch(() => {});
  };

  ws.onmessage = async (event) => {
    try {
      const request = JSON.parse(event.data);
      console.log(`[Browse Pilot] 收到指令: ${request.method}`);

      const response = await Router.dispatch(request);
      ws.send(JSON.stringify(response));
    } catch (e) {
      console.error('[Browse Pilot] 處理指令失敗:', e);
      const errorResp = createErrorResponse(null, ErrorCodes.EXTENSION_ERROR, e.message);
      ws.send(JSON.stringify(errorResp));
    }
  };

  ws.onclose = () => {
    console.log('[Browse Pilot] WebSocket 已斷線');
    isConnected = false;
    ws = null;
    browser.runtime.sendMessage({ type: 'connection_status', connected: false }).catch(() => {});
    scheduleReconnect();
  };

  ws.onerror = () => {
    console.error('[Browse Pilot] WebSocket 錯誤');
    // onclose 會自動觸發，不需重複處理
  };
}

function scheduleReconnect() {
  clearReconnectTimer();
  reconnectTimer = setTimeout(() => {
    console.log('[Browse Pilot] 嘗試重新連線...');
    connect();
  }, 3000);
}

function clearReconnectTimer() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

// 在 Router 上註冊所有指令 handler
// Background 端直接處理需要 tabs/cookies 等 API 權限的指令
// Content script 相關的指令透過 browser.tabs.sendMessage 轉發
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
      const [tab] = await browser.tabs.query({ active: true, currentWindow: true });
      if (!tab) throw { code: ErrorCodes.TAB_NOT_FOUND, message: '無活躍分頁' };

      const response = await browser.tabs.sendMessage(tab.id, { method, params });
      if (response.error) throw response.error;
      return response.result;
    });
  }
}

// 處理來自 popup 的訊息
browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'get_connection_status') {
    sendResponse({ connected: isConnected, port: wsPort, browser: 'firefox' });
    return false;
  }
  if (message.type === 'set_port') {
    wsPort = message.port;
    browser.storage.local.set({ wsPort: message.port });
    // 關閉舊連線並重新建立
    if (ws) ws.close();
    connect();
    sendResponse({ success: true });
    return false;
  }
  return false;
});

// 啟動：讀取設定 → 註冊 handler → 建立連線
loadConfig().then(() => {
  registerHandlers();
  connect();
});
