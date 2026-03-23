/**
 * MutationObserver 等待機制：等待 selector、文字、URL 條件成立
 */

/**
 * 等待指定 CSS selector 的元素出現或消失
 * @param {string} selector - CSS 選擇器
 * @param {Object} options
 * @param {'visible'|'hidden'} options.state - 等待可見或隱藏（預設 'visible'）
 * @param {number} options.timeout - 逾時毫秒數（預設 30000）
 * @returns {Promise<{ success: boolean, found: boolean }>}
 */
function waitForSelector(selector, options) {
  const state = (options && options.state) ? options.state : 'visible';
  const timeout = (options && options.timeout) ? options.timeout : 30000;

  return new Promise((resolve, reject) => {
    // 先檢查元素是否已符合條件
    const existing = document.querySelector(selector);
    if (state === 'visible' && existing && isVisible(existing)) {
      return resolve({ success: true, found: true });
    }
    if (state === 'hidden' && (!existing || !isVisible(existing))) {
      return resolve({ success: true, found: true });
    }

    // 使用物件保存 timerId，讓 observer callback 可透過 closure 存取
    const timerRef = { id: null };

    // 使用 MutationObserver 監聽 DOM 變化
    const observer = new MutationObserver(() => {
      const el = document.querySelector(selector);
      const matched =
        (state === 'visible' && el && isVisible(el)) ||
        (state === 'hidden' && (!el || !isVisible(el)));

      if (matched) {
        clearTimeout(timerRef.id);
        observer.disconnect();
        resolve({ success: true, found: true });
      }
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['style', 'class', 'aria-hidden', 'hidden'],
    });

    // 逾時後斷開 observer 並 reject
    timerRef.id = setTimeout(() => {
      observer.disconnect();
      reject({
        code: -32002, // TIMEOUT_ERROR
        message: `等待 selector "${selector}" 超時（${timeout}ms）`,
      });
    }, timeout);
  });
}

/**
 * 等待頁面包含指定文字
 * @param {string} text - 要等待的文字
 * @param {Object} options
 * @param {number} options.timeout - 逾時毫秒數（預設 30000）
 * @returns {Promise<{ success: boolean, found: boolean }>}
 */
function waitForText(text, options) {
  const timeout = (options && options.timeout) ? options.timeout : 30000;

  return new Promise((resolve, reject) => {
    // 先檢查文字是否已存在
    if (document.body.innerText.includes(text)) {
      return resolve({ success: true, found: true });
    }

    const timerRef = { id: null };

    const observer = new MutationObserver(() => {
      if (document.body.innerText.includes(text)) {
        clearTimeout(timerRef.id);
        observer.disconnect();
        resolve({ success: true, found: true });
      }
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true,
      characterData: true,
    });

    timerRef.id = setTimeout(() => {
      observer.disconnect();
      reject({
        code: -32002, // TIMEOUT_ERROR
        message: `等待文字 "${text}" 超時（${timeout}ms）`,
      });
    }, timeout);
  });
}

/**
 * 等待 URL 符合指定的 glob 或正則表達式模式
 * @param {string} pattern - 可用 * 作為萬用字元，或以 / 開頭和結尾表示正則
 * @param {Object} options
 * @param {number} options.timeout - 逾時毫秒數（預設 30000）
 * @returns {Promise<{ success: boolean, url: string }>}
 */
function waitForUrl(pattern, options) {
  const timeout = (options && options.timeout) ? options.timeout : 30000;
  const POLL_INTERVAL = 200;

  // 將 pattern 轉為可比對函式
  function matchesUrl(url) {
    // 若 pattern 以 / 開頭和結尾，視為正則表達式
    if (pattern.startsWith('/') && pattern.lastIndexOf('/') > 0) {
      const lastSlash = pattern.lastIndexOf('/');
      const regexBody = pattern.slice(1, lastSlash);
      const flags = pattern.slice(lastSlash + 1);
      try {
        const re = new RegExp(regexBody, flags);
        return re.test(url);
      } catch (_e) {
        return false;
      }
    }

    // 否則視為 glob，將 * 轉換成正則 .*
    const escaped = pattern.replace(/[.+?^${}()|[\]\\]/g, '\\$&').replace(/\*/g, '.*');
    const re = new RegExp(`^${escaped}$`);
    return re.test(url);
  }

  return new Promise((resolve, reject) => {
    // 先檢查當前 URL 是否已符合
    if (matchesUrl(location.href)) {
      return resolve({ success: true, url: location.href });
    }

    let elapsed = 0;

    // polling 方式定期檢查 URL（URL 變化無法用 MutationObserver 監聽）
    const intervalId = setInterval(() => {
      elapsed += POLL_INTERVAL;

      if (matchesUrl(location.href)) {
        clearInterval(intervalId);
        resolve({ success: true, url: location.href });
        return;
      }

      if (elapsed >= timeout) {
        clearInterval(intervalId);
        reject({
          code: -32002, // TIMEOUT_ERROR
          message: `等待 URL 符合 "${pattern}" 超時（${timeout}ms）`,
        });
      }
    }, POLL_INTERVAL);
  });
}
