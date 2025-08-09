// File: static/js/inventory.js
// ▼▼▼ 呼び出す関数を共通のものに変更 ▼▼▼
import { createUploadTableHTML, renderUploadTableRows } from './common_table.js';
let view, fileInput, outputContainer;

// ▼▼▼ 専用のrenderResults関数は不要なため削除 ▼▼▼

export function resetInventoryView() {
    if(outputContainer) {
        outputContainer.innerHTML = createUploadTableHTML('inventory-output-table');
        const tbody = document.querySelector('#inventory-output-table tbody');
        if(tbody) tbody.innerHTML = `<tr><td colspan="14">ファイルを選択してください。</td></tr>`;
    }
}

export function initInventoryView() {
    view = document.getElementById('inventory-view');
    if (!view) return;

    fileInput = document.getElementById('inventoryFileInput');
    outputContainer = document.getElementById('inventory-output-container');
    
    resetInventoryView();
    fileInput.addEventListener('change', async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        const tbody = document.querySelector('#inventory-output-table tbody');
        tbody.innerHTML = `<tr><td colspan="14" style="text-align:center;">アップロード処理中...</td></tr>`;

        const formData = new FormData();
        formData.append('file', file);

        try {
            window.showLoading();
            const res = await fetch('/api/inventory/upload', {
                method: 'POST',
                 body: formData,
            });
            const data = await res.json();
            if (!res.ok) {
                const errorText = data.message || await res.text();
                throw new Error(errorText);
            }
            window.showNotification(data.message || '棚卸ファイルを受け付けました', 'success');
            
            // ▼▼▼ 呼び出す関数を共通の renderUploadTableRows に変更 ▼▼▼
            renderUploadTableRows('inventory-output-table', data.details);
        } catch (err) {
            window.showNotification(`エラー: ${err.message}`, 'error');
            tbody.innerHTML = `<tr><td colspan="14" style="color:red; text-align:center;">エラー: ${err.message}</td></tr>`;
        } finally {
            window.hideLoading();
            fileInput.value = '';
        }
    });
}