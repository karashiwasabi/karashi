// File: static/js/app.js (最終修正版)
import { initDatUpload, resetDatUploadView } from './dat.js';
import { initUsageUpload, resetUsageUploadView } from './usage.js';
import { initInOut, resetInOutView } from './inout.js';
import { initBackupButtons } from './backup.js';
import { initMasterEdit, resetMasterEditView } from './master_edit.js';
import { initAggregation, resetAggregationView } from './aggregation.js';
import { initUpdateMaster } from './update_master.js';

document.addEventListener('DOMContentLoaded', () => {
  // 各モジュールの初期化
  initDatUpload();
  initUsageUpload();
  initInOut();
  initBackupButtons();
  initMasterEdit();
  initAggregation();
  initUpdateMaster();

  // --- ビュー切り替えロジック ---
  const uploadView = document.getElementById('upload-view');
  const inOutView = document.getElementById('in-out-view');
  const masterEditView = document.getElementById('master-edit-view');
  const aggregationView = document.getElementById('aggregation-view');

  const allViews = [uploadView, inOutView, masterEditView, aggregationView];

  function showView(viewToShow) {
    allViews.forEach(v => v.classList.add('hidden'));
    viewToShow.classList.remove('hidden');
  }

  // ▼▼▼ 修正点: 各ボタンのイベントリスナーを修正 ▼▼▼
  document.getElementById('inOutViewBtn').addEventListener('click', () => {
    resetInOutView();
    showView(inOutView);
  });

  document.getElementById('masterEditViewBtn').addEventListener('click', () => {
    resetMasterEditView();
    showView(masterEditView);
  });
  
  document.getElementById('aggregationBtn').addEventListener('click', () => {
    resetAggregationView();
    showView(aggregationView);
  });
  
  document.getElementById('datBtn').addEventListener('click', () => {
    resetDatUploadView();
    showView(uploadView);
  });

  document.getElementById('usageBtn').addEventListener('click', () => {
    resetUsageUploadView();
    showView(uploadView);
  });

  // ▼▼▼ 修正点: 初期表示の命令を削除 ▼▼▼
  // showView(inOutView);
});