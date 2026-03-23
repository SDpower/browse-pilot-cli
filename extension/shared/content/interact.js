/**
 * 頁面互動操作：點擊、輸入、按鍵、選擇、懸停等
 */

/**
 * 點擊指定索引的元素
 * @param {number} index - 元素索引
 */
function clickElement(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }
  // 先捲動到元素可視區域
  el.scrollIntoView({ block: 'center', behavior: 'instant' });
  el.click();
  return { success: true };
}

/**
 * 點擊畫面上的座標位置
 * @param {number} x - X 座標
 * @param {number} y - Y 座標
 */
function clickCoordinates(x, y) {
  const el = document.elementFromPoint(x, y);
  if (el) {
    el.click();
  }
  return { success: true };
}

/**
 * 對當前聚焦元素模擬文字輸入（使用 InputEvent）
 * @param {string} text - 要輸入的文字
 */
function typeText(text) {
  const el = document.activeElement;
  if (!el) {
    return { success: true };
  }

  // 使用 insertText InputEvent 逐字觸發
  el.dispatchEvent(new InputEvent('input', {
    bubbles: true,
    cancelable: true,
    inputType: 'insertText',
    data: text,
  }));

  // 若元素支援 value，直接附加文字
  if (el.value !== undefined) {
    el.value += text;
    el.dispatchEvent(new Event('input', { bubbles: true }));
    el.dispatchEvent(new Event('change', { bubbles: true }));
  }

  return { success: true };
}

/**
 * 清空並輸入文字到指定元素
 * @param {number} index - 元素索引
 * @param {string} text - 要輸入的文字
 */
function inputText(index, text) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }

  el.focus();

  // 清空原有內容
  el.value = '';
  el.dispatchEvent(new Event('input', { bubbles: true }));

  // 設定新值並觸發事件
  el.value = text;
  el.dispatchEvent(new Event('input', { bubbles: true }));
  el.dispatchEvent(new Event('change', { bubbles: true }));

  return { success: true };
}

/**
 * 解析並傳送按鍵事件
 * 支援格式：「Enter」、「Tab」、「Ctrl+a」、「Shift+Tab」等
 * @param {string} keys - 按鍵字串
 */
function sendKeys(keys) {
  // 解析修飾鍵與主鍵
  const parts = keys.split('+');
  const mainKey = parts[parts.length - 1];
  const modifiers = parts.slice(0, -1).map((m) => m.toLowerCase());

  const eventInit = {
    bubbles: true,
    cancelable: true,
    key: mainKey,
    code: mainKey,
    ctrlKey: modifiers.includes('ctrl'),
    altKey: modifiers.includes('alt'),
    shiftKey: modifiers.includes('shift'),
    metaKey: modifiers.includes('meta') || modifiers.includes('cmd'),
  };

  const target = document.activeElement || document.body;

  // 觸發完整的按鍵事件序列
  target.dispatchEvent(new KeyboardEvent('keydown', eventInit));
  target.dispatchEvent(new KeyboardEvent('keypress', eventInit));
  target.dispatchEvent(new KeyboardEvent('keyup', eventInit));

  return { success: true };
}

/**
 * 設定 select 元素的選取值
 * @param {number} index - 元素索引
 * @param {string} value - 要選取的 option value
 */
function selectOption(index, value) {
  const el = getElementByIndex(index);
  if (!el || el.tagName.toLowerCase() !== 'select') {
    throw { code: -32003, message: '元素不是 select 或索引不存在' };
  }

  el.value = value;
  el.dispatchEvent(new Event('change', { bubbles: true }));

  return { success: true };
}

/**
 * 模擬滑鼠懸停到指定元素
 * @param {number} index - 元素索引
 */
function hoverElement(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }

  // 觸發滑鼠進入事件
  el.dispatchEvent(new MouseEvent('mouseenter', { bubbles: true }));
  el.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));

  return { success: true };
}

/**
 * 對指定元素執行雙擊
 * @param {number} index - 元素索引
 */
function dblclickElement(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }

  el.scrollIntoView({ block: 'center', behavior: 'instant' });
  el.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }));

  return { success: true };
}

/**
 * 對指定元素執行右鍵點擊（觸發 contextmenu 事件）
 * @param {number} index - 元素索引
 */
function rightclickElement(index) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }

  el.scrollIntoView({ block: 'center', behavior: 'instant' });
  el.dispatchEvent(new MouseEvent('contextmenu', { bubbles: true }));

  return { success: true };
}

/**
 * 上傳檔案至 file input 元素
 * 注意：由於瀏覽器安全限制，content script 無法直接設定 file input 的 files 屬性，
 * 實際檔案選取需由 background script 透過 Native Messaging 處理。
 * 此函式負責驗證元素類型並回傳元素資訊。
 * @param {number} index - 元素索引
 * @param {string} filePath - 欲上傳的檔案絕對路徑
 */
function uploadFile(index, filePath) {
  const el = getElementByIndex(index);
  if (!el) {
    throw { code: -32003, message: `元素索引 ${index} 不存在` };
  }
  if (el.tagName.toLowerCase() !== 'input' || el.type !== 'file') {
    throw { code: -32602, message: '元素不是 file input' };
  }
  // 回傳元素資訊供 background script 使用
  return {
    success: true,
    selector: el.id ? '#' + el.id : null,
    tagName: el.tagName,
    path: filePath,
  };
}
