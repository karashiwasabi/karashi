// File: static/js/inout_details_table.js
import { initModal } from './inout_modal.js';
import { transactionTypeMap } from './common_table.js';

let tableBody, addRowBtn, modal;
const NEW_ROW_ID = 'new-row-';
let newRowCount = 0;

function createNewRow() {
    newRowCount++;
    const newRow = document.createElement('tr');
    const newRow2 = document.createElement('tr');
    newRow.dataset.rowId = `${NEW_ROW_ID}${newRowCount}`;
    newRow2.dataset.rowId = `${NEW_ROW_ID}${newRowCount}-2`;

    const html1 = `
        <td rowspan="2" class="date-cell"></td>
        <td rowspan="2" class="type-cell"></td>
        <td class="yj-code-cell"></td>
        <td colspan="2" class="product-name-cell"><input type="text" class="product-search-input" placeholder="製品名で検索..."></td>
        <td rowspan="2"><input type="number" name="datQuantity"></td>
        <td class="jan-qty-cell"></td>
        <td class="jan-pack-qty-cell"></td>
        <td class="jan-unit-cell"></td>
        <td><input type="number" name="unitPrice" step="any"></td>
        <td class="tax-amount-cell"></td>
        <td><input type="text" name="expiryDate"></td>
        <td><input type="text" name="lotNumber"></td>
        <td rowspan="2"><button class="delete-row-btn">削除</button></td>
    `;
    const html2 = `
        <td class="jan-code-cell"></td>
        <td class="package-spec-cell"></td>
        <td class="maker-name-cell"></td>
        <td class="yj-qty-cell"></td>
        <td class="yj-pack-qty-cell"></td>
        <td class="yj-unit-cell"></td>
        <td class="subtotal-cell"></td>
        <td class="tax-rate-cell"></td>
        <td class="client-code-cell"></td>
        <td class="receipt-number-cell"></td>
        <td class="line-number-cell"></td>
    `;
    newRow.innerHTML = html1;
    newRow2.innerHTML = html2;
    tableBody.appendChild(newRow);
    tableBody.appendChild(newRow2);
}

export function clearDetailsTable() {
    tableBody.innerHTML = '';
}

export function populateDetailsTable(records) {
    clearDetailsTable();
    let html = "";
    records.forEach(rec => {
        html += `
            <tr data-row-id="${rec.lineNumber}">
                <td rowspan="2" class="date-cell">${rec.transactionDate}</td>
                <td rowspan="2" class="type-cell">${transactionTypeMap[rec.flag] || ''}</td>
                <td class="yj-code-cell">${rec.yjCode || ''}</td>
                <td colspan="2" class="product-name-cell">${rec.productName || ''}</td>
                <td rowspan="2"><input type="number" name="datQuantity" value="${rec.datQuantity || ''}"></td>
                <td class="jan-qty-cell">${rec.janQuantity?.toFixed(2) || ''}</td>
                <td class="jan-pack-qty-cell">${rec.janPackInnerQty || ''}</td>
                <td class="jan-unit-cell">${rec.janUnitName || ''}</td>
                <td><input type="number" name="unitPrice" value="${rec.unitPrice || ''}" step="any"></td>
                <td class="tax-amount-cell">${rec.taxAmount?.toFixed(2) || ''}</td>
                <td><input type="text" name="expiryDate" value="${rec.expiryDate || ''}"></td>
                <td><input type="text" name="lotNumber" value="${rec.lotNumber || ''}"></td>
                <td rowspan="2"><button class="delete-row-btn">削除</button></td>
            </tr>
            <tr data-row-id="${rec.lineNumber}-2">
                <td class="jan-code-cell">${rec.janCode || ''}</td>
                <td class="package-spec-cell">${rec.packageSpec || ''}</td>
                <td class="maker-name-cell">${rec.makerName || ''}</td>
                <td class="yj-qty-cell">${rec.yjQuantity?.toFixed(2) || ''}</td>
                <td class="yj-pack-qty-cell">${rec.yjPackUnitQty || ''}</td>
                <td class="yj-unit-cell">${rec.yjUnitName || ''}</td>
                <td class="subtotal-cell">${rec.subtotal?.toFixed(2) || ''}</td>
                <td class="tax-rate-cell">${rec.taxRate != null ? (rec.taxRate * 100).toFixed(0) + "%" : ""}</td>
                <td class="client-code-cell">${rec.clientCode || ''}</td>
                <td class="receipt-number-cell">${rec.receiptNumber || ''}</td>
                <td class="line-number-cell">${rec.lineNumber || ''}</td>
            </tr>
        `;
    });
    tableBody.innerHTML = html;
}

export function getDetailsData() {
    const records = [];
    const rows = tableBody.querySelectorAll('tr:nth-child(odd)');
    rows.forEach(row1 => {
        const row2 = row1.nextElementSibling;
        const record = {
            yjCode: row1.querySelector('.yj-code-cell').textContent,
            janCode: row2.querySelector('.jan-code-cell').textContent,
            productName: row1.querySelector('.product-name-cell').textContent || row1.querySelector('.product-name-cell input')?.value,
            datQuantity: parseFloat(row1.querySelector('input[name="datQuantity"]').value) || 0,
            unitPrice: parseFloat(row1.querySelector('input[name="unitPrice"]').value) || 0,
            expiryDate: row1.querySelector('input[name="expiryDate"]').value,
            lotNumber: row1.querySelector('input[name="lotNumber"]').value,
            lineNumber: row2.querySelector('.line-number-cell').textContent,
        };
        records.push(record);
    });
    return records;
}

export function initDetailsTable() {
    tableBody = document.querySelector('#details-table tbody');
    addRowBtn = document.getElementById('addRowBtn');
    
    if(!tableBody || !addRowBtn) return;
    
    modal = initModal((selectedProduct) => {
        const activeInput = document.querySelector('.product-search-input.active');
        if (activeInput) {
            const row1 = activeInput.closest('tr');
            const row2 = row1.nextElementSibling;
            
            row1.querySelector('.product-name-cell').innerHTML = selectedProduct.productName; // Use innerHTML to remove input
            row1.querySelector('.yj-code-cell').textContent = selectedProduct.yjCode;
            row1.querySelector('input[name="unitPrice"]').value = selectedProduct.nhiPrice || 0;
            row1.querySelector('.jan-unit-cell').textContent = selectedProduct.janUnitName || '';
            
            row2.querySelector('.jan-code-cell').textContent = selectedProduct.productCode;
            row2.querySelector('.package-spec-cell').textContent = selectedProduct.formattedPackageSpec;
            row2.querySelector('.maker-name-cell').textContent = selectedProduct.makerName;
            
            row1.querySelector('input[name="datQuantity"]').focus();
        }
    });

    addRowBtn.addEventListener('click', createNewRow);

    tableBody.addEventListener('focusin', (e) => {
        if (e.target.classList.contains('product-search-input')) {
            document.querySelectorAll('.product-search-input.active').forEach(el => el.classList.remove('active'));
            e.target.classList.add('active');
            e.target.classList.add('highlight-search'); // ★★★ 追加: ハイライトクラス
            modal.show();
        }
    });
    
    // ★★★ 追加: フォーカスが外れたらハイライトを消す ★★★
    tableBody.addEventListener('focusout', (e) => {
        if (e.target.classList.contains('product-search-input')) {
            e.target.classList.remove('highlight-search');
        }
    });

    tableBody.addEventListener('click', (e) => {
        if (e.target.classList.contains('delete-row-btn')) {
            const row1 = e.target.closest('tr');
            const row2 = row1.nextElementSibling;
            row1.remove();
            if (row2) row2.remove();
        }
    });
}
