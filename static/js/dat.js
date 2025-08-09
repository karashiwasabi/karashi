// File: static/js/dat.js
import { createUploadTableHTML, renderUploadTableRows } from './common_table.js';

const datBtn = document.getElementById('datBtn');
const datInput = document.getElementById('datFileInput');
const uploadContainer = document.getElementById('upload-output-container');

export function resetDatUploadView() {
    uploadContainer.innerHTML = createUploadTableHTML('upload-output-table');
}

export function initDatUpload() {
    if (!datBtn || !datInput || !uploadContainer) return;
    // ★★★ 修正点: このファイル内のクリックイベントリスナーを削除 ★★★
    // datBtn.addEventListener('click', () => {
    //     datInput.click();
    // });
    datInput.addEventListener('change', async e => {
        if (!e.target.files.length) return;
        const tbody = document.querySelector('#upload-output-table tbody');
        tbody.innerHTML = `<tr><td colspan="14" style="text-align:center;">アップロード処理中...</td></tr>`;
        try {
            const formData = new FormData();
            for (const f of e.target.files) formData.append('file', f);
            const res = await fetch('/api/dat/upload', { method: 'POST', body: formData });
            if (!res.ok) throw new Error(res.status);
            const data = await res.json();
            renderUploadTableRows('upload-output-table', data.records);
        } catch (err) {
            tbody.innerHTML = `<tr><td colspan="14" style="color:red; text-align:center;">処理失敗: ${err.message}</td></tr>`;
        }
    });
}