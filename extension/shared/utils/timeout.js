/**
 * 逾時工具函式
 */

// 預設逾時時間（毫秒）
const DEFAULT_TIMEOUT_MS = 30000;

/**
 * 為 Promise 加上逾時限制
 * @param {Promise} promise - 要執行的非同步操作
 * @param {number} ms - 逾時毫秒數（預設 30000ms）
 * @param {string} errorMessage - 逾時時的錯誤訊息
 * @returns {Promise} 若在期限內完成則回傳原始結果，否則 reject TimeoutError
 */
function withTimeout(promise, ms, errorMessage) {
  const timeout = ms !== undefined ? ms : DEFAULT_TIMEOUT_MS;
  const message = errorMessage || `操作逾時（${timeout}ms）`;

  let timerId;

  const timeoutPromise = new Promise((_, reject) => {
    timerId = setTimeout(() => {
      reject({
        code: -32002, // ErrorCodes.TIMEOUT_ERROR
        message,
      });
    }, timeout);
  });

  return Promise.race([promise, timeoutPromise]).finally(() => {
    // 清除計時器，避免記憶體洩漏
    clearTimeout(timerId);
  });
}
