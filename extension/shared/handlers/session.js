/**
 * Session 管理 handler：回報擴充功能連線狀態與基本資訊
 */

const SessionHandler = (() => {
  // 將瀏覽器 API 留在 handler 私有作用域，避免背景腳本共用全域名稱衝突。
  const api = typeof browser !== 'undefined' ? browser : chrome;

  return {
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
})();
