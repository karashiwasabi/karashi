// File: static/js/app.js
import { initDatUpload } from './dat.js';
import { initUsageUpload } from './usage.js';
import { initInOut } from './inout.js';

document.addEventListener('DOMContentLoaded', () => {
  // 各モジュールの初期化
  initDatUpload();
  initUsageUpload();
  initInOut();

  // ビュー切り替え
  const uploadView = document.getElementById('upload-view');
  const inOutView = document.getElementById('in-out-view');
  document.getElementById('uploadViewBtn').addEventListener('click', () => {
    uploadView.classList.remove('hidden');
    inOutView.classList.add('hidden');
  });
  document.getElementById('inOutViewBtn').addEventListener('click', () => {
    inOutView.classList.remove('hidden');
    uploadView.classList.add('hidden');
  });

  // 初期ビュー
  uploadView.classList.remove('hidden');
  inOutView.classList.add('hidden');
});