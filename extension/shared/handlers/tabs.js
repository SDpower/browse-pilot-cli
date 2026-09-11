/**
 * 分頁管理 handler（在 background script 中執行）
 */

const TabsHandler = (() => {
  // 將瀏覽器 API 留在 handler 私有作用域，避免背景腳本共用全域名稱衝突。
  const api = typeof browser !== 'undefined' ? browser : chrome;

  return {
  /**
   * 取得所有分頁清單
   * @returns {Promise<{ tabs: Array }>}
   */
  async getTabs() {
    const tabs = await api.tabs.query({});
    return {
      tabs: tabs.map((t, i) => ({
        index: i,
        id: t.id,
        url: t.url,
        title: t.title,
        active: t.active,
      })),
    };
  },

  /**
   * 切換到指定索引的分頁
   * @param {{ index: number }} params
   */
  async switchTab(params) {
    const tabs = await api.tabs.query({});

    if (params.index < 0 || params.index >= tabs.length) {
      throw { code: -32004, message: `分頁索引 ${params.index} 不存在` };
    }

    await api.tabs.update(tabs[params.index].id, { active: true });
    return { success: true, tabId: tabs[params.index].id };
  },

  /**
   * 關閉指定索引的分頁（省略索引時關閉當前分頁）
   * @param {{ index?: number }} params
   */
  async closeTab(params) {
    if (params.index !== undefined) {
      const tabs = await api.tabs.query({});

      if (params.index < 0 || params.index >= tabs.length) {
        throw { code: -32004, message: `分頁索引 ${params.index} 不存在` };
      }

      await api.tabs.remove(tabs[params.index].id);
    } else {
      // 關閉當前分頁
      const [tab] = await api.tabs.query({ active: true, currentWindow: true });
      await api.tabs.remove(tab.id);
    }

    return { success: true };
  },

  };
})();
