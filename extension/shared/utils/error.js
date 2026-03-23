/**
 * JSON-RPC 2.0 風格的錯誤碼常數與輔助函式
 */

// 錯誤碼定義
const ErrorCodes = {
  // JSON-RPC 標準錯誤碼
  PARSE_ERROR: -32700,
  INVALID_REQUEST: -32600,
  METHOD_NOT_FOUND: -32601,
  INVALID_PARAMS: -32602,

  // 擴充錯誤碼（應用層）
  EXTENSION_ERROR: -32000,
  CONNECTION_ERROR: -32001,
  TIMEOUT_ERROR: -32002,
  ELEMENT_NOT_FOUND: -32003,
  TAB_NOT_FOUND: -32004,
  INJECTION_ERROR: -32005,
  PERMISSION_ERROR: -32006,
  STALE_ELEMENT: -32007,
  BROWSER_NOT_FOUND: -32008,
  NATIVE_MESSAGING_ERROR: -32009,
};

/**
 * 建立錯誤物件
 * @param {number} code - 錯誤碼
 * @param {string} message - 錯誤訊息
 * @param {*} data - 附加資料（可選）
 */
function createError(code, message, data) {
  return { code, message, data: data !== undefined ? data : null };
}

/**
 * 建立成功回應
 * @param {string|number} id - 請求 ID
 * @param {*} result - 回應結果
 */
function createResponse(id, result) {
  return { id, result };
}

/**
 * 建立錯誤回應
 * @param {string|number} id - 請求 ID
 * @param {number} code - 錯誤碼
 * @param {string} message - 錯誤訊息
 * @param {*} data - 附加資料（可選）
 */
function createErrorResponse(id, code, message, data) {
  return { id, error: createError(code, message, data) };
}
