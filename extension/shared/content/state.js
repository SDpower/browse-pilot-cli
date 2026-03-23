/**
 * DOM 狀態管理：遍歷可互動元素並建立索引快取
 */

// 元素快取：index -> DOM element（使用 const，內容透過 .clear() 重置而非重新賦值）
const elementCache = new Map();

// 快取版本號，每次 buildState() 呼叫時遞增
let cacheVersion = 0;

/**
 * 判斷元素是否可見（排除隱藏、aria-hidden、尺寸為零的元素）
 * @param {Element} el
 * @returns {boolean}
 */
function isVisible(el) {
  const style = getComputedStyle(el);

  // 檢查 CSS 隱藏屬性
  if (style.display === 'none' || style.visibility === 'hidden') {
    return false;
  }

  // 檢查 ARIA 隱藏屬性
  if (el.getAttribute('aria-hidden') === 'true') {
    return false;
  }

  // 檢查元素實際尺寸
  const rect = el.getBoundingClientRect();
  if (rect.width === 0 && rect.height === 0) {
    return false;
  }

  return true;
}

/**
 * 取得元素的描述名稱
 * 優先順序：aria-label > innerText（截斷 50 字）> placeholder > name > title > alt > id
 * @param {Element} el
 * @returns {string}
 */
function getElementName(el) {
  return el.getAttribute('aria-label')
    || (el.innerText || '').trim().substring(0, 50)
    || el.placeholder
    || el.name
    || el.title
    || el.alt
    || el.id
    || '';
}

/**
 * 產生元素的唯一 CSS selector
 * 優先使用 id，否則組合兩層 parent + nth-child
 * @param {Element} el
 * @returns {string}
 */
function generateSelector(el) {
  // 有 id 時直接使用
  if (el.id) {
    return `#${el.id}`;
  }

  // 向上走最多 2 層，組合 tag + nth-child
  const parts = [];
  let current = el;
  let depth = 0;

  while (current && current !== document.body && depth < 3) {
    const tag = current.tagName.toLowerCase();
    const parent = current.parentElement;

    if (parent) {
      // 計算在同層兄弟中的位置（從 1 開始）
      const siblings = Array.from(parent.children).filter(
        (c) => c.tagName === current.tagName
      );
      if (siblings.length > 1) {
        const index = siblings.indexOf(current) + 1;
        parts.unshift(`${tag}:nth-of-type(${index})`);
      } else {
        parts.unshift(tag);
      }
    } else {
      parts.unshift(tag);
    }

    current = current.parentElement;
    depth++;
  }

  return parts.join(' > ');
}

/**
 * 掃描 DOM，建立可互動元素索引
 * @returns {{ version: number, elements: Array }}
 */
function buildState() {
  elementCache.clear();
  cacheVersion++;
  const elements = [];

  // 選取所有可互動元素
  const interactables = document.querySelectorAll(
    'a[href], button, input, select, textarea, ' +
    '[role="button"], [role="link"], [role="checkbox"], [role="radio"], ' +
    '[role="tab"], [role="menuitem"], [role="switch"], [role="combobox"], ' +
    '[onclick], [tabindex]'
  );

  let index = 0;
  for (const el of interactables) {
    // 排除不可見或已停用的元素
    if (!isVisible(el) || el.disabled) {
      continue;
    }

    elementCache.set(index, el);
    elements.push({
      index,
      tag: el.tagName.toLowerCase(),
      type: el.type || null,
      role: el.getAttribute('role') || null,
      name: getElementName(el),
      value: el.value !== undefined ? el.value : null,
      placeholder: el.placeholder || null,
      selector: generateSelector(el),
      visible: true,
      bbox: el.getBoundingClientRect().toJSON(),
    });
    index++;
  }

  return { version: cacheVersion, elements };
}

/**
 * 根據索引從快取取得元素
 * @param {number} idx
 * @returns {Element|null}
 */
function getElementByIndex(idx) {
  return elementCache.get(idx) || null;
}
