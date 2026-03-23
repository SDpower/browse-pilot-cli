/**
 * 截圖 handler（在 background script 中執行）
 * captureVisibleTab 須在 background 呼叫，需要 activeTab 權限
 */

// 相容 Firefox（browser）和 Chrome/Edge（chrome）
const api = typeof browser !== 'undefined' ? browser : chrome;

const ScreenshotHandler = {
  /**
   * 截取畫面
   * @param {Object} params - 選項參數
   * @param {boolean} [params.full] - true 表示全頁截圖，false/undefined 表示可視區域截圖
   * @returns {Promise<Object>}
   */
  async capture(params) {
    if (params && params.full) {
      return this.captureFullPage();
    }
    return this.captureVisible();
  },

  /**
   * 截取當前可視分頁的畫面
   * @returns {Promise<{ data: string, width: null, height: null }>}
   */
  async captureVisible() {
    // captureVisibleTab 回傳格式：「data:image/png;base64,...」
    const dataUrl = await api.tabs.captureVisibleTab(null, { format: 'png' });

    // 擷取 base64 部分（去除 data URI 前綴）
    const base64 = dataUrl.split(',')[1];

    return {
      data: base64,
      // 寬高資訊需從 content script 或 tab 另外取得，目前回傳 null
      width: null,
      height: null,
    };
  },

  /**
   * 全頁截圖：逐段捲動 + captureVisibleTab，回傳所有片段讓 CLI 端（Go image 標準庫）拼接
   *
   * 注意：MV3 service worker 無 DOM / Canvas，無法在 background 端合成圖片，
   * 因此採用「回傳片段陣列」的策略，由 CLI 端負責拼接。
   *
   * 通訊大小說明：
   * - Firefox：透過 WebSocket 回傳，無大小限制
   * - Chrome：透過 port.postMessage（extension → host）回傳，上限 64MB，實際不會超出
   *
   * @returns {Promise<{
   *   full: true,
   *   width: number,
   *   height: number,
   *   segmentHeight: number,
   *   segments: Array<{ segment: number, scrollY: number, data: string }>
   * }>}
   */
  async captureFullPage() {
    // 取得目前作用中的分頁
    const [tab] = await api.tabs.query({ active: true, currentWindow: true });

    // 透過 eval_js 取得頁面完整尺寸與目前捲動位置
    const dimensions = await api.tabs.sendMessage(tab.id, {
      method: 'eval_js',
      params: {
        code: 'JSON.stringify({' +
          'scrollHeight: document.documentElement.scrollHeight,' +
          'clientHeight: document.documentElement.clientHeight,' +
          'scrollWidth: document.documentElement.scrollWidth,' +
          'clientWidth: document.documentElement.clientWidth,' +
          'currentScrollY: window.scrollY' +
          '})',
      },
    });

    const dim = JSON.parse(dimensions.result.result);
    const { scrollHeight, clientHeight } = dim;
    const originalScrollY = dim.currentScrollY;

    // 計算需要截取的段數（向上取整）
    const segments = Math.ceil(scrollHeight / clientHeight);
    const captures = [];

    // 逐段捲動並截圖
    for (let i = 0; i < segments; i++) {
      const scrollY = i * clientHeight;

      // 捲動到指定垂直位置
      await api.tabs.sendMessage(tab.id, {
        method: 'eval_js',
        params: { code: `window.scrollTo(0, ${scrollY})` },
      });

      // 等待 100ms 讓瀏覽器完成重繪
      await new Promise(resolve => setTimeout(resolve, 100));

      // 截取可視區域（此呼叫必須在 background script 中執行）
      const dataUrl = await api.tabs.captureVisibleTab(null, { format: 'png' });

      // 去除 data URI 前綴，只保留 base64 內容
      const base64 = dataUrl.split(',')[1];

      captures.push({
        segment: i,
        scrollY: scrollY,
        data: base64,
      });
    }

    // 恢復原始捲動位置，避免影響使用者操作
    await api.tabs.sendMessage(tab.id, {
      method: 'eval_js',
      params: { code: `window.scrollTo(0, ${originalScrollY})` },
    });

    return {
      full: true,
      width: dim.clientWidth,
      height: scrollHeight,
      segmentHeight: clientHeight,
      segments: captures,
    };
  },
};
