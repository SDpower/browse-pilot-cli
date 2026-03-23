/**
 * 頁面捲動控制
 */

/**
 * 向上或向下捲動頁面
 * @param {'up'|'down'} direction - 捲動方向
 * @param {number|undefined} amount - 捲動距離（像素），省略時使用視窗高度
 * @returns {{ success: boolean, scrollY: number }}
 */
function scrollPage(direction, amount) {
  // 未指定距離時使用視窗高度
  const distance = amount !== undefined ? amount : window.innerHeight;
  const delta = direction === 'up' ? -distance : distance;

  window.scrollBy({ top: delta, behavior: 'instant' });

  return { success: true, scrollY: window.scrollY };
}
