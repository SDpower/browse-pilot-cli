/**
 * Cookie 管理 handler（在 background script 中執行）
 * Content script 無法存取 cookies API，須由 background 代理
 */

// 相容 Firefox（browser）和 Chrome/Edge（chrome）
const api = typeof browser !== 'undefined' ? browser : chrome;

const CookiesHandler = {
  /**
   * 取得 Cookie 清單
   * @param {{ url?: string }} params
   * @returns {Promise<{ cookies: Array }>}
   */
  async getCookies(params) {
    const query = {};
    if (params.url) {
      query.url = params.url;
    }
    const cookies = await api.cookies.getAll(query);
    return { cookies };
  },

  /**
   * 設定單一 Cookie
   * @param {{ url?: string, domain?: string, name: string, value: string,
   *           secure?: boolean, sameSite?: string, expires?: number }} params
   */
  async setCookie(params) {
    await api.cookies.set({
      url: params.url || `https://${params.domain}`,
      name: params.name,
      value: params.value,
      domain: params.domain,
      secure: params.secure,
      sameSite: params.sameSite,
      // expires 是 Unix 時間戳（毫秒），需轉換為秒
      expirationDate: params.expires ? Math.floor(params.expires / 1000) : undefined,
    });
    return { success: true };
  },

  /**
   * 清除指定 URL 的所有 Cookie（省略 URL 時清除全部）
   * @param {{ url?: string }} params
   * @returns {Promise<{ success: boolean, count: number }>}
   */
  async clearCookies(params) {
    const query = {};
    if (params.url) {
      query.url = params.url;
    }
    const cookies = await api.cookies.getAll(query);

    for (const cookie of cookies) {
      // 根據 cookie 的 secure 屬性決定使用 http 或 https
      const protocol = cookie.secure ? 'https' : 'http';
      const domain = cookie.domain.startsWith('.') ? cookie.domain.slice(1) : cookie.domain;
      const url = `${protocol}://${domain}${cookie.path}`;
      await api.cookies.remove({ url, name: cookie.name });
    }

    return { success: true, count: cookies.length };
  },
};
