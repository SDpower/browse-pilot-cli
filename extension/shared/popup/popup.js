/**
 * Popup UI 邏輯：向 background script 查詢連線狀態並更新畫面
 */

// 相容 Firefox（browser）和 Chrome/Edge（chrome）
const popupApi = typeof browser !== 'undefined' ? browser : chrome;

/**
 * 更新狀態指示器的樣式與文字
 * @param {boolean} connected - 是否已連線
 */
function updateStatusUI(connected) {
  const statusEl = document.getElementById('status');
  const statusTextEl = document.getElementById('status-text');

  if (connected) {
    statusEl.classList.remove('disconnected');
    statusEl.classList.add('connected');
    statusTextEl.textContent = '已連線';
  } else {
    statusEl.classList.remove('connected');
    statusEl.classList.add('disconnected');
    statusTextEl.textContent = '未連線';
  }
}

/**
 * 向 background script 取得連線狀態並更新 UI
 */
async function fetchStatus() {
  try {
    const response = await popupApi.runtime.sendMessage({ method: 'session_status' });

    if (response && response.result) {
      const { connected, browser: browserType, version } = response.result;

      updateStatusUI(connected);
      document.getElementById('browser-type').textContent = browserType || '-';
      document.getElementById('version').textContent = version || '-';
    } else {
      updateStatusUI(false);
    }
  } catch (_err) {
    // background script 無法回應時顯示未連線狀態
    updateStatusUI(false);
  }
}

/**
 * 向 background script 查詢 WebSocket 連接埠資訊
 */
async function fetchPort() {
  try {
    const response = await popupApi.runtime.sendMessage({ method: 'get_port' });

    if (response && response.result && response.result.port) {
      document.getElementById('port').textContent = String(response.result.port);
    } else {
      document.getElementById('port').textContent = '-';
    }
  } catch (_err) {
    document.getElementById('port').textContent = '-';
  }
}

// DOM 載入完成後執行
document.addEventListener('DOMContentLoaded', () => {
  fetchStatus();
  fetchPort();
});
