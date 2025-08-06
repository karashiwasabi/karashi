import { initModal, showModal } from './inout_modal.js';

let tableBody, addRowBtn;
let newRowCount = 0;

/**
 * 渡された製品データに基づいて行のHTMLを生成する
 * @param {object} record - 製品データまたは取引データ
 * @returns {string} - 2つの<tr>要素からなるHTML文字列
 */
function createRowsHTML(record = {}) {
    newRowCount++;
    const rowId = record.lineNumber || `new-${newRowCount}`;
    
    const janQuantity = record.janQuantity || 1; // デフォルト値を1に
    const nhiPrice = record.nhiPrice || 0;
    const janPackInnerQty = record.janPackInnerQty || 0;
    
    // 計算ロジック
    const yjQuantity = janQuantity * janPackInnerQty;
    const subtotal = yjQuantity * nhiPrice;

    const upperRow = `
        <tr data-row-id="${rowId}">
            <td class="product-name-cell" style="cursor: pointer; text-decoration: underline; color: blue;">${record.productName || 'ここをクリックして製品を検索'}</td>
            <td class="display-maker-name">${record.makerName || ''}</td>
            <td><input type="number" name="janQuantity" value="${janQuantity}" step="any"></td>
            <td class="display-yj-quantity right">${yjQuantity.toFixed(2)}</td>
            <td class="display-unit-price right">${nhiPrice.toFixed(4)}</td>
            <td class="display-subtotal right">${subtotal.toFixed(2)}</td>
            <td><input type="text" name="expiryDate" value="${record.expiryDate || ''}" placeholder="YYYYMM"></td>
            <td><input type="text" name="lotNumber" value="${record.lotNumber || ''}"></td>
            <td rowspan="2" class="center"><button class="delete-row-btn btn">削除</button></td>
        </tr>`;

    const lowerRow = `
        <tr data-row-id-lower="${rowId}">
            <td colspan="8" class="lower-info-cell">
                <span class="info-label">JAN:</span> <span class="display-jan-code">${record.productCode || record.janCode || ''}</span>
                <span class="info-label">YJ:</span> <span class="display-yj-code">${record.yjCode || ''}</span>
                <span class="info-label">包装:</span> <span class="display-package-spec">${record.formattedPackageSpec || record.packageSpec || ''}</span>
            </td>
        </tr>`;
    
    return upperRow + lowerRow;
}


/**
 * 読み込んだデータから明細テーブルを生成する
 * @param {Array<object>} records
 */
export function populateDetailsTable(records) {
    clearDetailsTable();
    let allRowsHTML = '';
    records.forEach(rec => {
        allRowsHTML += createRowsHTML(rec);
    });
    tableBody.innerHTML = allRowsHTML;
    
    tableBody.querySelectorAll('tr[data-row-id]').forEach((row, index) => {
        if (records[index]) {
            // サーバーから読み込んだレコード全体をproductデータとして保持
            row.dataset.product = JSON.stringify(records[index]);
        }
    });
}

/**
 * テーブル内の全ての明細行をクリアする
 */
export function clearDetailsTable() {
    tableBody.innerHTML = '';
}

/**
 * 全ての明細行からデータを収集してサーバー送信用の配列として返す
 * @returns {Array<object>}
 */
export function getDetailsData() {
    const records = [];
    const rows = tableBody.querySelectorAll('tr[data-row-id]');
    
    rows.forEach(row => {
        const productDataString = row.dataset.product;
        if (!productDataString || productDataString === '{}') return;

        const productData = JSON.parse(productDataString);
        
        // ▼▼▼ ここを修正 ▼▼▼
        // サーバーに送るレコードを作成
        // 製品マスターの全情報と、ユーザーが入力した情報をマージする
        const record = {
            ...productData, // モーダルから選択した製品マスターの全情報
            janQuantity: parseFloat(row.querySelector('input[name="janQuantity"]').value) || 0,
            expiryDate: row.querySelector('input[name="expiryDate"]').value,
            lotNumber: row.querySelector('input[name="lotNumber"]').value,
        };
        records.push(record);
    });
    
    return records;
}

/**
 * 行内の表示を再計算する
 * @param {HTMLTableRowElement} upperRow
 */
function recalculateRow(upperRow) {
    const productDataString = upperRow.dataset.product;
    if (!productDataString) return;

    const product = JSON.parse(productDataString);
    const lowerRow = upperRow.nextElementSibling;
    
    const janQuantity = parseFloat(upperRow.querySelector('[name="janQuantity"]').value) || 0;
    const nhiPrice = parseFloat(product.nhiPrice) || 0;
    const janPackInnerQty = parseFloat(product.janPackInnerQty) || 0;

    const yjQuantity = janQuantity * janPackInnerQty;
    const subtotal = yjQuantity * nhiPrice;

    upperRow.querySelector('.display-yj-quantity').textContent = yjQuantity.toFixed(2);
    upperRow.querySelector('.display-subtotal').textContent = subtotal.toFixed(2);
}

/**
 * 明細セクションの初期化
 */
export function initDetailsTable() {
    tableBody = document.querySelector('#inout-details-table tbody');
    addRowBtn = document.getElementById('addRowBtn');
    
    if(!tableBody || !addRowBtn) return;

    initModal((selectedProduct, activeRow) => {
        // 選択された製品データ(マスターの全情報)を行に保存
        activeRow.dataset.product = JSON.stringify(selectedProduct);
        const lowerRow = activeRow.nextElementSibling;

        // UIを更新
        activeRow.querySelector('.product-name-cell').textContent = selectedProduct.productName;
        activeRow.querySelector('.display-maker-name').textContent = selectedProduct.makerName;
        activeRow.querySelector('.display-unit-price').textContent = (selectedProduct.nhiPrice || 0).toFixed(4);
        
        lowerRow.querySelector('.display-jan-code').textContent = selectedProduct.productCode;
        lowerRow.querySelector('.display-yj-code').textContent = selectedProduct.yjCode;
        lowerRow.querySelector('.display-package-spec').textContent = selectedProduct.formattedPackageSpec;

        const quantityInput = activeRow.querySelector('input[name="janQuantity"]');
        quantityInput.focus();
        quantityInput.select();
        recalculateRow(activeRow);
    });

    addRowBtn.addEventListener('click', () => {
        tableBody.insertAdjacentHTML('beforeend', createRowsHTML());
    });
    
    tableBody.addEventListener('click', (e) => {
        const target = e.target;
        if (target.classList.contains('delete-row-btn')) {
            const upperRow = target.closest('tr');
            const lowerRow = upperRow.nextElementSibling;
            lowerRow.remove();
            upperRow.remove();
        }
        if (target.classList.contains('product-name-cell')) {
            const activeRow = target.closest('tr');
            showModal(activeRow);
        }
    });

    tableBody.addEventListener('input', (e) => {
        if(e.target.name === 'janQuantity') {
            recalculateRow(e.target.closest('tr'));
        }
    });
}