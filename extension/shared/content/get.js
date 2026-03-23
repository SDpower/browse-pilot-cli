/**
 * 頁面資料擷取：標題、HTML、文字、值、屬性、邊界框
 */

/**
 * 取得頁面標題
 * @returns {{ title: string }}
 */
function getTitle() {
  return { title: document.title };
}

/**
 * 取得頁面或指定元素的 HTML 內容
 * @param {string|undefined} selector - CSS 選擇器（省略時取整個頁面）
 * @returns {{ html: string|null }}
 */
function getHtml(selector) {
  if (selector) {
    const el = document.querySelector(selector);
    return { html: el ? el.outerHTML : null };
  }
  return { html: document.documentElement.outerHTML };
}

/**
 * 取得指定元素的文字內容
 * @param {number} index - 元素索引
 * @returns {{ text: string }}
 */
function getText(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }
  return { text: el.innerText || el.textContent || '' };
}

/**
 * 取得指定元素的 value 屬性
 * @param {number} index - 元素索引
 * @returns {{ value: string|null }}
 */
function getValue(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }
  return { value: el.value !== undefined ? el.value : null };
}

/**
 * 取得指定元素的所有 HTML 屬性
 * @param {number} index - 元素索引
 * @returns {{ attributes: Object }}
 */
function getAttributes(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }

  const attrs = {};
  for (const attr of el.attributes) {
    attrs[attr.name] = attr.value;
  }
  return { attributes: attrs };
}

/**
 * 取得指定元素的邊界框（位置與尺寸）
 * @param {number} index - 元素索引
 * @returns {{ x: number, y: number, width: number, height: number }}
 */
function getBbox(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }
  const rect = el.getBoundingClientRect();
  return {
    x: rect.x,
    y: rect.y,
    width: rect.width,
    height: rect.height,
  };
}
