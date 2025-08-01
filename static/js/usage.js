// File: static/js/usage.js
import {
  createUploadTableHTML,
  renderUploadTableRows
} from './common_table.js';

export function initUsageUpload() {
  const btn       = document.getElementById('usageBtn');
  const input     = document.getElementById('usageFileInput');
  const container = document.getElementById('upload-output-container');

  container.innerHTML = createUploadTableHTML('upload-output-table');

  btn.addEventListener('click', () => {
    container.innerHTML = createUploadTableHTML('upload-output-table');
    input.click();
  });

  input.addEventListener('change', async e => {
    if (!e.target.files.length) return;
    const tbody = document.querySelector('#upload-output-table tbody');
    tbody.innerHTML = `<tr><td colspan="14" style="text-align:center;">アップロード処理中...</td></tr>`;

    try {
      const formData = new FormData();
      for (const f of e.target.files) formData.append('file', f);
      const res = await fetch('/api/usage/upload', { method: 'POST', body: formData });
      if (!res.ok) throw new Error(res.status);
      const data = await res.json();
      renderUploadTableRows('upload-output-table', data.records);
    } catch (err) {
      tbody.innerHTML =
        `<tr><td colspan="14" style="color:red; text-align:center;">
           処理失敗: ${err.message}
         </td></tr>`;
    }
  });
}