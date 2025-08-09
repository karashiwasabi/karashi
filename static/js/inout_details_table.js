// File: static/js/inout_details_table.js (修正後)
import { initModal, showModal } from './inout_modal.js';
import { transactionTypeMap } from './common_table.js';

let tableBody, addRowBtn, tableContainer;

/**
 * DAT/USAGE画面と共通のテーブルヘッダーを生成する
 * @param {string} tableId - テーブル要素のID
 * @returns {string} - テーブル全体のHTML
 */
function createInoutTableHTML(tableId) {
    const colWidths = [
        "5.83%", "4.5%", "9.15%", "13.77%", "13.77%",
        "2.91%", "5.34%", "7.69%", "5.34%", "7.21%",
        "5.91%", "7.21%", "7.21%", "5.75%"
    ];
    const colgroup = `<colgroup>${colWidths.map(w => `<col style="width:${w};">`).join("")}</colgroup>`;

    // ▼▼▼ 修正点: 最初のヘッダーを「日付」から「操作」に変更 ▼▼▼
    const header = `
        <thead>
            <tr>
                <th rowspan="2">操作</th>
                <th rowspan="2">種別</th>
                <th>YJ</th>
                <th colspan="2">製品名</th>
                <th rowspan="2">個数</th>
                <th>JAN数量</th>
                <th>JAN包装数</th>
                <th>JAN単位</th>
                <th>単価</th>
                <th>税額</th>
                <th>期限</th>
                <th>ロット</th>
                <th>MA</th>
            </tr>
            <tr>
                <th>JAN</th>
                <th>包装</th>
                <th>メーカー</th>
                <th>YJ数量</th>
                <th>YJ包装数</th>
                <th>YJ単位</th>
                <th>金額</th>
                <th>税率</th>
                <th>得意先</th>
                <th>伝票番号</th>
                <th>行</th>
            </tr>
        </thead>
    `;
    return `<table id="${tableId}" class="data-table">${colgroup}${header}<tbody>
        <tr><td colspan="14">ヘッダーで情報を選択後、「明細を追加」ボタンを押してください。</td></tr>
    </tbody></table>`;
}

/**
 * 渡された製品データに基づいて共通レイアウトの行HTMLを生成する
 * @param {object} record - 製品データまたは取引データ
 * @returns {string} - 2つの<tr>要素からなるHTML文字列
 */
function createRowsHTML(record = {}) {
    const rowId = record.lineNumber || `new-${Date.now()}`;
    
    const janQuantity = record.janQuantity ?? 1;
    const nhiPrice = record.nhiPrice || 0;
    const janPackInnerQty = record.janPackInnerQty || 0;
    const yjQuantity = janQuantity * janPackInnerQty;
    const subtotal = yjQuantity * nhiPrice;

    const transactionType = record.flag ? (transactionTypeMap[record.flag] || '') : '';
    
    // ▼▼▼ 修正点: 最初のセルを削除ボタンに変更し、最後のセルをMA表示に戻す ▼▼▼
    const upperRow = `
        <tr data-row-id="${rowId}">
            <td rowspan="2" class="center"><button class="delete-row-btn btn">削除</button></td>
            <td rowspan="2">${transactionType}</td>
            <td class="display-yj-code">${record.yjCode || ''}</td>
            <td colspan="2" class="product-name-cell left" style="cursor: pointer; text-decoration: underline; color: blue;">${record.productName || 'ここをクリックして製品を検索'}</td>
            <td rowspan="2" class="right">${record.datQuantity || ''}</td>
            <td><input type="number" name="janQuantity" value="${janQuantity}" step="any" class="right"></td>
            <td class="right display-jan-pack-unit-qty">${record.janPackUnitQty || ''}</td>
            <td class="display-jan-unit-name">${record.janUnitName || ''}</td>
            <td class="right display-unit-price">${nhiPrice.toFixed(4)}</td>
            <td class="right">${record.taxAmount || ''}</td>
            <td><input type="text" name="expiryDate" value="${record.expiryDate || ''}" placeholder="YYYYMM"></td>
            <td><input type="text" name="lotNumber" value="${record.lotNumber || ''}"></td>
            <td class="left">${record.processFlagMA || ''}</td>
        </tr>`;

    const lowerRow = `
        <tr data-row-id-lower="${rowId}">
            <td class="display-jan-code">${record.productCode || record.janCode || ''}</td>
            <td class="left display-package-spec">${record.formattedPackageSpec || record.packageSpec || ''}</td>
            <td class="left display-maker-name">${record.makerName || ''}</td>
            <td class="right display-yj-quantity">${yjQuantity.toFixed(2)}</td>
            <td class="right display-yj-pack-unit-qty">${record.yjPackUnitQty || ''}</td>
            <td class="display-yj-unit-name">${record.yjUnitName || ''}</td>
            <td class="right display-subtotal">${subtotal.toFixed(2)}</td>
            <td class="right">${record.taxRate != null ? (record.taxRate * 100).toFixed(0) + "%" : ""}</td>
            <td class="left">${record.clientCode || ''}</td>
            <td class="left">${record.receiptNumber || ''}</td>
            <td class="right">${record.lineNumber || ''}</td>
        </tr>`;

    return upperRow + lowerRow;
}

/**
 * 読み込んだデータから明細テーブルを生成する
 * @param {Array<object>} records
 */
export function populateDetailsTable(records) {
    if (!records || records.length === 0) {
        clearDetailsTable();
        return;
    }
    tableBody.innerHTML = records.map(createRowsHTML).join('');
    
    tableBody.querySelectorAll('tr[data-row-id]').forEach((row, index) => {
        if (records[index]) {
            row.dataset.product = JSON.stringify(records[index]);
        }
    });
}

/**
 * テーブル内の全ての明細行をクリアする
 */
export function clearDetailsTable() {
    if(tableBody) {
        tableBody.innerHTML = `<tr><td colspan="14">ヘッダーで情報を選択後、「明細を追加」ボタンを押してください。</td></tr>`;
    }
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
        
        const record = {
            ...productData,
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
    if (!lowerRow) return;
    
    const janQuantity = parseFloat(upperRow.querySelector('[name="janQuantity"]').value) || 0;
    const nhiPrice = parseFloat(product.nhiPrice) || 0;
    const janPackInnerQty = parseFloat(product.janPackInnerQty) || 0;

    const yjQuantity = janQuantity * janPackInnerQty;
    const subtotal = yjQuantity * nhiPrice;

    lowerRow.querySelector('.display-yj-quantity').textContent = yjQuantity.toFixed(2);
    lowerRow.querySelector('.display-subtotal').textContent = subtotal.toFixed(2);
}

/**
 * 明細セクションの初期化
 */
export function initDetailsTable() {
    tableContainer = document.getElementById('inout-details-container');
    addRowBtn = document.getElementById('addRowBtn');
    if(!tableContainer || !addRowBtn) return;
    
    tableContainer.innerHTML = createInoutTableHTML('inout-details-table');
    tableBody = document.querySelector('#inout-details-table tbody');

    initModal((selectedProduct, activeRow) => {
        activeRow.dataset.product = JSON.stringify(selectedProduct);
        const lowerRow = activeRow.nextElementSibling;

        // --- 上段 ---
        activeRow.querySelector('.product-name-cell').textContent = selectedProduct.productName;
        activeRow.querySelector('.display-yj-code').textContent = selectedProduct.yjCode;
        activeRow.querySelector('.display-unit-price').textContent = (selectedProduct.nhiPrice || 0).toFixed(4);
        activeRow.querySelector('.display-jan-pack-unit-qty').textContent = selectedProduct.janPackUnitQty || '';
        activeRow.querySelector('.display-jan-unit-name').textContent = selectedProduct.janUnitName || '';
        
        // --- 下段 ---
        lowerRow.querySelector('.display-jan-code').textContent = selectedProduct.productCode;
        lowerRow.querySelector('.display-package-spec').textContent = selectedProduct.formattedPackageSpec || selectedProduct.packageSpec || '';
        lowerRow.querySelector('.display-maker-name').textContent = selectedProduct.makerName;
        lowerRow.querySelector('.display-yj-pack-unit-qty').textContent = selectedProduct.yjPackUnitQty || '';
        lowerRow.querySelector('.display-yj-unit-name').textContent = selectedProduct.yjUnitName || '';
        
        const quantityInput = activeRow.querySelector('input[name="janQuantity"]');
        quantityInput.focus();
        quantityInput.select();
        recalculateRow(activeRow);
    });

    addRowBtn.addEventListener('click', () => {
        if (tableBody.querySelector('td[colspan="14"]')) {
            tableBody.innerHTML = '';
        }
        tableBody.insertAdjacentHTML('beforeend', createRowsHTML());
    });

    tableBody.addEventListener('click', (e) => {
        const target = e.target;
        if (target.classList.contains('delete-row-btn')) {
            const upperRow = target.closest('tr');
            const lowerRow = upperRow.nextElementSibling;
            if(lowerRow) lowerRow.remove();
            upperRow.remove();
            if (tableBody.children.length === 0) {
                clearDetailsTable();
            }
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