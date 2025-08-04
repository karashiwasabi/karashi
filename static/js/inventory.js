// File: static/js/inventory.js
import { createUploadTableHTML, renderUploadTableRows } from './common_table.js';

let view, uploadBtn, fileInput, outputContainer;

function renderResults(records) {
    if (!records || records.length === 0) {
        outputContainer.innerHTML = `<p>調整対象のデータはありませんでした。</p>`;
        return;
    }
    // YJ単位での調整結果なので、少し列名を変更して表示
    const flagMap = {4: "棚卸増", 5: "棚卸減"};
    const tbody = document.querySelector('#inventory-output-table tbody');
    let html = "";
    records.forEach(rec => {
        html += `
          <tr>
            <td rowspan="2">${rec.transactionDate || ""}</td>
            <td rowspan="2">${flagMap[rec.flag] || ""}</td>
            <td>${rec.yjCode || ""}</td>
            <td class="left" colspan="2">${rec.productName || "(新規マスター)"}</td>
            <td class="right" rowspan="2"></td>
            <td class="right"></td>
            <td class="right"></td>
            <td></td>
            <td class="right"></td>
            <td class="right"></td>
            <td></td>
            <td class="left"></td>
            <td class="left"></td>
          </tr>
          <tr>
            <td>${rec.janCode || ""}</td>
            <td class="left"></td>
            <td class="left"></td>
            <td class="right">${rec.yjQuantity?.toFixed(2) || ""}</td>
            <td class="right"></td>
            <td></td>
            <td class="right"></td>
            <td class="right"></td>
            <td class="left"></td>
            <td class="left">${rec.receiptNumber || ""}</td>
            <td class="right">${rec.lineNumber || ""}</td>
          </tr>
        `;
    });
    tbody.innerHTML = html;
}

export function initInventoryView() {
    view = document.getElementById('inventory-view');
    if (!view) return;

    uploadBtn = document.getElementById('inventoryUploadBtn');
    fileInput = document.getElementById('inventoryFileInput');
    outputContainer = document.getElementById('inventory-output-container');
    
    outputContainer.innerHTML = createUploadTableHTML('inventory-output-table');

    uploadBtn.addEventListener('click', () => fileInput.click());

    fileInput.addEventListener('change', async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        outputContainer.innerHTML = `<p>処理中...</p>`;

        const formData = new FormData();
        formData.append('file', file);

        try {
            const res = await fetch('/api/inventory/upload', {
                method: 'POST',
                body: formData,
            });
            const data = await res.json();
            if (!res.ok) {
                // サーバーからのエラーメッセージを優先して表示
                const errorText = await res.text();
                try {
                    const errorJson = JSON.parse(errorText);
                    throw new Error(errorJson.message || errorText);
                } catch {
                    throw new Error(errorText);
                }
            }
            alert(data.message);
            outputContainer.innerHTML = createUploadTableHTML('inventory-output-table');
            renderResults(data.details);

        } catch (err) {
            outputContainer.innerHTML = `<p style="color:red;">エラー: ${err.message}</p>`;
        } finally {
            fileInput.value = '';
        }
    });
}