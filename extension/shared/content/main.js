/**
 * Content script 進入點
 * 監聽來自 background script 的訊息並分派到對應的操作函式
 *
 * 載入順序（manifest content_scripts 定義）：
 * state.js → interact.js → wait.js → get.js → eval.js → scroll.js → main.js
 */

// 相容 Firefox（browser）和 Chrome/Edge（chrome）
const api = typeof browser !== 'undefined' ? browser : chrome;

/**
 * 主訊息監聽器
 * 格式：{ method: string, params: Object }
 * 回應格式：{ result: * } 或 { error: { code, message } }
 */
api.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  const { method, params } = message;

  let resultPromise;

  switch (method) {
    case 'get_state':
      resultPromise = Promise.resolve(buildState());
      break;

    case 'click':
      // 支援座標點擊和索引點擊兩種模式
      if (params.x !== undefined && params.y !== undefined) {
        resultPromise = Promise.resolve(clickCoordinates(params.x, params.y));
      } else {
        resultPromise = Promise.resolve(clickElement(params.index));
      }
      break;

    case 'type_text':
      resultPromise = Promise.resolve(typeText(params.text));
      break;

    case 'input_text':
      resultPromise = Promise.resolve(inputText(params.index, params.text));
      break;

    case 'send_keys':
      resultPromise = Promise.resolve(sendKeys(params.keys));
      break;

    case 'select_option':
      resultPromise = Promise.resolve(selectOption(params.index, params.value));
      break;

    case 'hover':
      resultPromise = Promise.resolve(hoverElement(params.index));
      break;

    case 'dblclick':
      resultPromise = Promise.resolve(dblclickElement(params.index));
      break;

    case 'rightclick':
      resultPromise = Promise.resolve(rightclickElement(params.index));
      break;

    case 'upload_file':
      resultPromise = Promise.resolve(uploadFile(params.index, params.path));
      break;

    case 'wait_selector':
      resultPromise = waitForSelector(params.selector, params);
      break;

    case 'wait_text':
      resultPromise = waitForText(params.text, params);
      break;

    case 'wait_url':
      resultPromise = waitForUrl(params.pattern, params);
      break;

    case 'get_title':
      resultPromise = Promise.resolve(getTitle());
      break;

    case 'get_html':
      resultPromise = Promise.resolve(getHtml(params && params.selector));
      break;

    case 'get_text':
      resultPromise = Promise.resolve(getText(params.index));
      break;

    case 'get_value':
      resultPromise = Promise.resolve(getValue(params.index));
      break;

    case 'get_attributes':
      resultPromise = Promise.resolve(getAttributes(params.index));
      break;

    case 'get_bbox':
      resultPromise = Promise.resolve(getBbox(params.index));
      break;

    case 'eval_js':
      resultPromise = Promise.resolve(evalJs(params.code));
      break;

    case 'scroll':
      resultPromise = Promise.resolve(scrollPage(params.direction, params.amount));
      break;

    default:
      // 未知方法：立即回傳錯誤
      sendResponse({ error: { code: -32601, message: `未知方法: ${method}` } });
      return false;
  }

  // 非同步等待結果後回傳
  resultPromise
    .then((result) => sendResponse({ result }))
    .catch((err) => sendResponse({
      error: {
        code: (err && err.code) ? err.code : -32000,
        message: (err && err.message) ? err.message : String(err),
      },
    }));

  // 回傳 true 表示將非同步呼叫 sendResponse
  return true;
});
