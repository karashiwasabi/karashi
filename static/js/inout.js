// File: static/js/inout.js (最終修正版)
import { initHeader, resetHeader } from './inout_header.js';
import { initDetailsTable, clearDetailsTable } from './inout_details_table.js';

/**
 * Initializes all In/Out screen functionality.
 */
export async function initInOut() {
  initDetailsTable();
  await initHeader();
}

// ▼▼▼ 修正点: リセット関数をエクスポート ▼▼▼
export function resetInOutView() {
    clearDetailsTable();
    resetHeader();
}