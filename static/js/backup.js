// File: static/js/backup.js (Corrected)

/**
 * Initializes the event listeners for the backup buttons.
 */
export function initBackupButtons() {
    // Client buttons
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

    // *** ADDED: Product buttons ***
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

/**
 * Generic file upload handler.
 * @param {Event} event - The file input change event.
 * @param {string} url - The API endpoint to upload the file to.
 */
async function handleFileUpload(event, url) {
    const fileInput = event.target;
    const file = fileInput.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
        const res = await fetch(url, {
            method: 'POST',
            body: formData,
        });

        const resData = await res.json();
        if (!res.ok) {
            throw new Error(resData.message || 'インポートに失敗しました。');
        }
        alert(resData.message);
        window.location.reload();

    } catch (err) {
        console.error(err);
        alert(`エラー: ${err.message}`);
    } finally {
        fileInput.value = '';
    }
}