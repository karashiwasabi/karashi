// File: static/js/usage.js (最終修正版)
import { createUploadTableHTML, renderUploadTableRows } from './common_table.js';

const usageBtn = document.getElementById('usageBtn');
const usageInput = document.getElementById('usageFileInput');
const uploadContainer = document.getElementById('upload-output-container');

// ▼▼▼ 修正点: リセット関数をエクスポート ▼▼▼
export function resetUsageUploadView() {
    uploadContainer.innerHTML = createUploadTableHTML('upload-output-table');
}

export function initUsageUpload() {
    if(!usageBtn || !usageInput || !uploadContainer) return;

    usageBtn.addEventListener('click', () => {
        usageInput.click();
    });

    usageInput.addEventListener('change', async e => {
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
            tbody.innerHTML = `<tr><td colspan="14" style="color:red; text-align:center;">処理失敗: ${err.message}</td></tr>`;
        }
    });
}