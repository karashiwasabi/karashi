// File: static/js/app.js
import { initDatUpload, resetDatUploadView } from './dat.js';
import { initUsageUpload, resetUsageUploadView } from './usage.js';
import { initInOut, resetInOutView } from './inout.js';
import { initBackupButtons } from './backup.js';
import { initMasterEdit, resetMasterEditView } from './master_edit.js';
import { initAggregation, resetAggregationView } from './aggregation.js';
import { initUpdateMaster } from './update_master.js';
import { initReprocessButton } from './reprocess.js';
import { initInventoryView, resetInventoryView } from './inventory.js';
// ▼▼▼ 以下を追記 ▼▼▼
import { initMonthEndView, resetMonthEndView } from './month_end.js';

// --- グローバルUI要素の管理 ---
const loadingOverlay = document.getElementById('loading-overlay');
const notificationBox = document.getElementById('notification-box');
const backToTopBtn = document.getElementById('back-to-top');

window.showLoading = () => loadingOverlay.classList.remove('hidden');
window.hideLoading = () => loadingOverlay.classList.add('hidden');
window.showNotification = (message, type = 'success') => {
    notificationBox.textContent = message;
    notificationBox.className = 'hidden';
    void notificationBox.offsetWidth;
    notificationBox.className = '';
    notificationBox.classList.add(type, 'show');
    setTimeout(() => {
        notificationBox.classList.remove('show');
    }, 3000);
};

document.addEventListener('DOMContentLoaded', () => {
  // --- 初期化フラグ ---
  let isDatInitialized = false;
  let isUsageInitialized = false;
  let isInOutInitialized = false;
  let isMasterEditInitialized = false;
  let isAggregationInitialized = false;
  let isInventoryInitialized = false;
  let isMonthEndInitialized = false; // ▼▼▼ 追記
  let isSampleInitialized = false;

  // --- 状態管理 ---
  let lastClickedButtonId = null;

  // --- ビューを持たないボタンの初期化 ---
  initBackupButtons();
  initUpdateMaster();
  initReprocessButton();

  // --- DOM要素 ---
  const mainHeader = document.getElementById('main-header');
  const menuButtons = mainHeader.querySelectorAll('.btn');
  const allViews = document.querySelectorAll('main > div[id$="-view"]');
  const viewMap = {
      'inOutViewBtn': 'in-out-view',
      'masterEditViewBtn': 'master-edit-view',
      'datBtn': 'upload-view',
      'usageBtn': 'upload-view',
      'aggregationBtn': 'aggregation-view',
      'monthEndViewBtn': 'month-end-view', // ▼▼▼ 追記
      'inventoryBtn': 'inventory-view',
      'sampleBtn': 'sample-view',
  };

  /**
   * 指定されたボタンIDに基づいてビューを切り替える関数
   * @param {string} buttonId 押されたボタンのID
   */
  function switchView(buttonId) {
    const targetViewId = viewMap[buttonId];
    if (!targetViewId) return;

    // 1. 全てのボタンを非アクティブ化
    menuButtons.forEach(btn => {
        btn.classList.remove('active');
    });
    // 2. クリックされたボタンのみアクティブ化
    const clickedButton = document.getElementById(buttonId);
    if (clickedButton) {
        clickedButton.classList.add('active');
    }

    // 3. 全てのビューを非表示
    allViews.forEach(v => {
        v.classList.add('hidden');
    });
    // 4. 対象のビューのみ表示
    const targetView = document.getElementById(targetViewId);
    if(targetView) {
        targetView.classList.remove('hidden');
    }

    // 5. 各ビューの初期化/リセット処理
    switch (targetViewId) {
        case 'in-out-view':
            if (!isInOutInitialized) { initInOut(); isInOutInitialized = true; }
            else { resetInOutView(); }
            break;
        case 'master-edit-view':
            if (!isMasterEditInitialized) { initMasterEdit(); isMasterEditInitialized = true; }
            else { resetMasterEditView(); }
            break;
        case 'aggregation-view':
            if (!isAggregationInitialized) { initAggregation(); isAggregationInitialized = true; }
            else { resetAggregationView(); }
            break;
        case 'inventory-view':
            if (!isInventoryInitialized) { initInventoryView(); isInventoryInitialized = true; }
            else { resetInventoryView(); }
            document.getElementById('inventoryFileInput').click();
            break;
        case 'month-end-view': // ▼▼▼ 追記
            if (!isMonthEndInitialized) { initMonthEndView(); isMonthEndInitialized = true; }
            resetMonthEndView();
            break;
        case 'sample-view':
            if (!isSampleInitialized) {
                if (targetView) {
                    targetView.innerHTML = '<p style="padding: 20px; font-size: 18px;">サンプル</p>';
                }
                isSampleInitialized = true;
            }
            break;
        case 'upload-view':
            if (buttonId === 'datBtn') {
                if (!isDatInitialized) { initDatUpload(); isDatInitialized = true; }
                resetDatUploadView();
                document.getElementById('datFileInput').click();
            } else if (buttonId === 'usageBtn') {
                if (!isUsageInitialized) { initUsageUpload(); isUsageInitialized = true; }
                resetUsageUploadView();
                document.getElementById('usageFileInput').click();
            }
            break;
    }
  }

  // --- イベントリスナー ---
  mainHeader.addEventListener('click', (e) => {
    const button = e.target;
    if (button.matches('.btn')) {
        if (viewMap[button.id]) {
            lastClickedButtonId = button.id;
            switchView(lastClickedButtonId);
        }
    }
  });

  window.addEventListener('scroll', () => {
    if (window.pageYOffset > 300) {
        backToTopBtn.classList.remove('hidden');
    } else {
        backToTopBtn.classList.add('hidden');
    }
  });

  backToTopBtn.addEventListener('click', () => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });

  // --- 初期表示 ---
  lastClickedButtonId = 'inOutViewBtn';
  switchView(lastClickedButtonId);
});