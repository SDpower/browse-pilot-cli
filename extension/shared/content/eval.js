/**
 * 在頁面 context 中執行任意 JavaScript 程式碼
 */

/**
 * 執行 JS 字串並回傳結果
 * 結果會嘗試 JSON 序列化；若無法序列化則轉換為字串
 * @param {string} code - 要執行的 JavaScript 程式碼
 * @returns {{ result: * }}
 */
// eslint-disable-next-line no-unused-vars
function evalJs(code) {
  let rawResult;

  try {
    // eslint-disable-next-line no-eval
    rawResult = eval(code);
  } catch (e) {
    throw { code: -32000, message: e.message };
  }

  // 嘗試 JSON 序列化（排除循環引用等無法序列化的值）
  try {
    const serialized = JSON.parse(JSON.stringify(rawResult));
    return { result: serialized };
  } catch (_e) {
    // 序列化失敗時轉為字串
    return { result: String(rawResult) };
  }
}
