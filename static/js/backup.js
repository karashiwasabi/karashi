// File: static/js/backup.js

async function handleFileUpload(event, url) {
    const fileInput = event.target;
    const file = fileInput.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);
    
    window.showLoading(); // ローディング開始

    try {
        const res = await fetch(url, {
            method: 'POST',
            body: formData,
        });

        const resData = await res.json();
        if (!res.ok) {
            throw new Error(resData.message || 'インポートに失敗しました。');
        }
        window.showNotification(resData.message, 'success'); // ★★★ 修正: alert -> showNotification
        setTimeout(() => window.location.reload(), 1000); // 通知が見えるように少し待つ

    } catch (err) {
        console.error(err);
        window.showNotification(`エラー: ${err.message}`, 'error'); // ★★★ 修正: alert -> showNotification
    } finally {
        window.hideLoading(); // ローディング終了
        fileInput.value = '';
    }
}

export function initBackupButtons() {
    const exportClientsBtn = document.getElementById('exportClientsBtn');
    const importClientsBtn = document.getElementById('importClientsBtn');
    const importClientsInput = document.getElementById('importClientsInput');

    if (exportClientsBtn && importClientsBtn && importClientsInput) {
        exportClientsBtn.addEventListener('click', () => {
            window.location.href = '/api/clients/export';
        });
        importClientsBtn.addEventListener('click', () => {
            importClientsInput.click();
        });
        importClientsInput.addEventListener('change', (event) => {
            handleFileUpload(event, '/api/clients/import');
        });
    }

    const exportProductsBtn = document.getElementById('exportProductsBtn');
    const importProductsBtn = document.getElementById('importProductsBtn');
    const importProductsInput = document.getElementById('importProductsInput');
    
    if (exportProductsBtn && importProductsBtn && importProductsInput) {
        exportProductsBtn.addEventListener('click', () => {
            window.location.href = '/api/products/export';
        });
        importProductsBtn.addEventListener('click', () => {
            importProductsInput.click();
        });
        importProductsInput.addEventListener('change', (event) => {
            handleFileUpload(event, '/api/products/import');
        });
    }
}
