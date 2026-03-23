/**
 * Session 管理 handler：回報擴充功能連線狀態與基本資訊
 */

// 相容 Firefox（browser）和 Chrome/Edge（chrome）
const api = typeof browser !== 'undefined' ? browser : chrome;

const SessionHandler = {
  /**
   * 取得擴充功能目前的狀態與版本資訊
   * @returns {Promise<{ connected: boolean, browser: string, version: string, extensionId: string }>}
   */
  async getStatus() {
    const manifest = api.runtime.getManifest();
    return {
      connected: true,
      // 以 browser 全域變數判斷瀏覽器類型
      browser: typeof browser !== 'undefined' ? 'firefox' : 'chrome',
      version: manifest.version,
      extensionId: api.runtime.id,
    };
  },
};
