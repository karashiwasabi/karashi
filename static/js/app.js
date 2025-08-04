// File: static/js/app.js
import { initDatUpload, resetDatUploadView } from './dat.js';
import { initUsageUpload, resetUsageUploadView } from './usage.js';
import { initInOut, resetInOutView } from './inout.js';
import { initBackupButtons } from './backup.js';
import { initMasterEdit, resetMasterEditView } from './master_edit.js';
import { initAggregation, resetAggregationView } from './aggregation.js';
import { initUpdateMaster } from './update_master.js';
import { initReprocessButton } from './reprocess.js';

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
  };

  /**
   * 指定されたボタンIDに基づいてビューを切り替える関数
   * @param {string} buttonId 押されたボタンのID
   */
  function switchView(buttonId) {
    const targetViewId = viewMap[buttonId];
    if (!targetViewId) return;

    // 1. ボタンのアクティブ状態を更新 (最後に押されたボタンのみ)
    menuButtons.forEach(btn => {
        if (viewMap[btn.id]) {
            btn.classList.toggle('active', btn.id === buttonId);
        }
    });

    // 2. ビューの表示/非表示を切り替え
    allViews.forEach(v => {
        v.classList.toggle('hidden', v.id !== targetViewId);
    });
    
    // 3. 各ビューの初期化/リセット処理
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
    if (e.target.matches('.btn') && viewMap[e.target.id]) {
        lastClickedButtonId = e.target.id;
        switchView(lastClickedButtonId);
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
