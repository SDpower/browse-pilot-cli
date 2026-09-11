/**
 * 導航操作 handler（在 background script 中執行）
 * 使用 tabs API 進行頁面導航
 */

const NavigationHandler = (() => {
  // 將瀏覽器 API 留在 handler 私有作用域，避免背景腳本共用全域名稱衝突。
  const api = typeof browser !== 'undefined' ? browser : chrome;

  return {
  /**
   * 導航到指定 URL，等待頁面載入完成
   * @param {{ url: string }} params
   * @returns {Promise<{ success: boolean, url: string, title: string }>}
   */
  async navigate(params) {
    const tab = await api.tabs.update({ url: params.url });

    return new Promise((resolve) => {
      function listener(tabId, changeInfo, updatedTab) {
        if (tabId === tab.id && changeInfo.status === 'complete') {
          api.tabs.onUpdated.removeListener(listener);
          resolve({
            success: true,
            url: updatedTab.url || '',
            title: updatedTab.title || '',
          });
        }
      }

      api.tabs.onUpdated.addListener(listener);
    });
  },

  /**
   * 瀏覽器返回上一頁
   * @param {Object} _params - 未使用
   */
  async goBack(_params) {
    await api.tabs.goBack();
    return { success: true };
  },

  /**
   * 瀏覽器前進下一頁
   * @param {Object} _params - 未使用
   */
  async goForward(_params) {
    await api.tabs.goForward();
    return { success: true };
  },

  /**
   * 重新載入當前頁面
   * @param {Object} _params - 未使用
   */
  async reload(_params) {
    await api.tabs.reload();
    return { success: true };
  },
  };
})();
