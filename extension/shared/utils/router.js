/**
 * 指令路由器：將 JSON-RPC method 對應到 handler 函式
 */

const Router = {
  // 已註冊的 handler 對應表
  _handlers: {},

  /**
   * 註冊 method 對應的 handler
   * @param {string} method - 方法名稱
   * @param {Function} handler - 處理函式，接收 params 並回傳結果（可為 Promise）
   */
  register(method, handler) {
    this._handlers[method] = handler;
  },

  /**
   * 根據 request 找到對應 handler 並執行
   * @param {Object} request - JSON-RPC 請求物件 { id, method, params }
   * @returns {Promise<Object>} JSON-RPC 回應 { id, result } 或 { id, error }
   */
  async dispatch(request) {
    const { id, method, params } = request;

    const handler = this._handlers[method];
    if (!handler) {
      // 找不到對應的 handler
      return {
        id,
        error: {
          code: -32601, // METHOD_NOT_FOUND
          message: `找不到方法: ${method}`,
          data: null,
        },
      };
    }

    try {
      const result = await handler(params || {});
      return { id, result };
    } catch (err) {
      // 已結構化的 RPCError 必須完整保留 code、message 與 data。
      // 一般例外則統一標記為 EXTENSION_ERROR，交由 MCP 層產生安全訊息。
      const code = Number.isInteger(err?.code) ? err.code : -32000;
      const message = typeof err?.message === 'string' ? err.message : String(err);
      const data = err && Object.prototype.hasOwnProperty.call(err, 'data')
        ? err.data
        : null;
      return { id, error: { code, message, data } };
    }
  },
};
