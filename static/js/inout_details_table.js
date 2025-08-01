// File: static/js/inout_details_table.js (Final Version)
import { initModal, showModal } from './inout_modal.js';

const detailsTableBody = document.querySelector('#details-table tbody');
const addRowBtn = document.getElementById('addRowBtn');
let targetRow = null;

/**
 * Calculates totals for a given row.
 * @param {HTMLTableRowElement} firstRow - The first row of the entry.
 */
function calculateRowTotals(firstRow) {
    if (!firstRow) return;
    const secondRow = firstRow.nextElementSibling;

    const quantity = parseFloat(firstRow.cells[5].querySelector('input').value) || 0;
    const unitPrice = parseFloat(firstRow.cells[8].querySelector('input').value) || 0;
    const taxRate = parseFloat(document.getElementById('in-out-taxrate').value) || 0;
    
    const janPackInnerQty = parseFloat(firstRow.dataset.janPackInnerQty) || 0;
    const yjQuantity = quantity * janPackInnerQty;
    secondRow.cells[3].textContent = yjQuantity.toFixed(2);

    const subtotal = quantity * unitPrice;
    const taxAmount = Math.floor(subtotal * (taxRate / 100));

    secondRow.cells[6].textContent = subtotal.toFixed(0); // Amount cell
    firstRow.cells[9].textContent = taxAmount.toFixed(0);  // Tax cell
}

/**
 * Handles product selection from the modal.
 * @param {object} product - The selected product data.
 */
function handleProductSelection(product) {
    if (!targetRow) return;
    const firstRow = targetRow;
    const secondRow = targetRow.nextElementSibling;

    // --- Store all necessary master data in the row's dataset ---
    firstRow.dataset.janPackInnerQty = product.janPackInnerQty || '0';
    firstRow.dataset.kanaName = product.kanaName || '';
    firstRow.dataset.packageForm = product.packageSpec || '';
    firstRow.dataset.flagPoison = product.flagPoison || '0';
    firstRow.dataset.flagDeleterious = product.flagDeleterious || '0';
    firstRow.dataset.flagNarcotic = product.flagNarcotic || '0';
    firstRow.dataset.flagPsychotropic = product.flagPsychotropic || '0';
    firstRow.dataset.flagStimulant = product.flagStimulant || '0';
    firstRow.dataset.flagStimulantRaw = product.flagStimulantRaw || '0';

    // --- Populate visible cells ---
    firstRow.cells[2].textContent = product.yjCode || '';
    firstRow.cells[3].textContent = product.productName || '';
    firstRow.cells[6].textContent = product.janPackUnitQty || '';
    firstRow.cells[7].textContent = product.janUnitName || '';
    firstRow.cells[8].querySelector('input').value = product.nhiPrice ? product.nhiPrice.toFixed(2) : '0.00';

    secondRow.cells[0].textContent = product.productCode || '';
    secondRow.cells[1].textContent = product.formattedPackageSpec || '';
    secondRow.cells[2].textContent = product.makerName || '';
    secondRow.cells[4].textContent = product.yjPackUnitQty || '';
    secondRow.cells[5].textContent = product.yjUnitName || '';
    
    calculateRowTotals(firstRow);
    targetRow = null;
}

/**
 * Adds a new empty row to the table.
 */
function addRow() {
  const newRowsHTML = `
    <tr>
      <td rowspan="2"></td>
      <td rowspan="2"></td>
      <td></td>
      <td colspan="2" class="product-name-cell"></td>
      <td rowspan="2"></td>
      <td><input type="text" class="recalc-trigger jan-quantity-input"></td>
      <td></td>
      <td></td>
      <td><input type="text" class="recalc-trigger unit-price-input"></td>
      <td></td>
      <td><input type="date"></td>
      <td><input type="text"></td>
      <td><button class="delete-row-btn">削除</button></td>
    </tr>
    <tr>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
      <td></td>
    </tr>
  `;
  detailsTableBody.insertAdjacentHTML('beforeend', newRowsHTML);
}

/**
 * Initializes the details table functionality.
 */
export function initDetailsTable() {
    if (!detailsTableBody || !addRowBtn) return;
    initModal(handleProductSelection);
    addRowBtn.addEventListener('click', addRow);
    
    detailsTableBody.addEventListener('click', (event) => {
        if (event.target.classList.contains('delete-row-btn')) {
            const firstRow = event.target.closest('tr');
            const secondRow = firstRow.nextElementSibling;
            firstRow.remove();
            if (secondRow) secondRow.remove();
            return;
        }
        const productNameCell = event.target.closest('.product-name-cell');
        if (productNameCell) {
            targetRow = productNameCell.closest('tr');
            showModal();
        }
    });

    detailsTableBody.addEventListener('input', (event) => {
        if (event.target.classList.contains('recalc-trigger')) {
            const firstRow = event.target.closest('tr');
            calculateRowTotals(firstRow);
        }
    });

    document.getElementById('in-out-taxrate').addEventListener('input', () => {
        const rows = detailsTableBody.querySelectorAll('tr:nth-child(odd)');
        rows.forEach(calculateRowTotals);
    });
}

/**
 * Gathers and returns data from the details table.
 * @returns {Array<object>}
 */
export function getDetailsData() {
    const records = [];
    const rows = detailsTableBody.querySelectorAll('tr:nth-child(odd)');
    rows.forEach((firstRow, index) => {
        const secondRow = firstRow.nextElementSibling;
        const productName = firstRow.cells[3].textContent.trim();
        if (!productName) return;

        const ds = firstRow.dataset; // Dataset for easy access

        const record = {
            lineNumber: (index + 1).toString(),
            yjCode: firstRow.cells[2].textContent,
            productName: productName,
            janQuantity: parseFloat(firstRow.cells[5].querySelector('input').value) || 0,
            janPackUnitQty: parseFloat(firstRow.cells[6].textContent) || 0,
            janUnitName: firstRow.cells[7].textContent,
            unitPrice: parseFloat(firstRow.cells[8].querySelector('input').value) || 0,
            expiryDate: firstRow.cells[10].querySelector('input').value.replace(/-/g, ''),
            lotNumber: firstRow.cells[11].querySelector('input').value,
            janCode: secondRow.cells[0].textContent,
            packageSpec: secondRow.cells[1].textContent,
            makerName: secondRow.cells[2].textContent,
            yjQuantity: parseFloat(secondRow.cells[3].textContent) || 0,
            yjPackUnitQty: parseFloat(secondRow.cells[4].textContent) || 0,
            yjUnitName: secondRow.cells[5].textContent,
            subtotal: parseFloat(secondRow.cells[6].textContent) || 0,
            taxAmount: parseFloat(firstRow.cells[9].textContent) || 0,
            // --- Add the new fields from the dataset ---
            datQuantity: 0, // Set to 0 as requested
            kanaName: ds.kanaName,
            packageForm: ds.packageForm,
            flagPoison: parseInt(ds.flagPoison, 10) || 0,
            flagDeleterious: parseInt(ds.flagDeleterious, 10) || 0,
            flagNarcotic: parseInt(ds.flagNarcotic, 10) || 0,
            flagPsychotropic: parseInt(ds.flagPsychotropic, 10) || 0,
            flagStimulant: parseInt(ds.flagStimulant, 10) || 0,
            flagStimulantRaw: parseInt(ds.flagStimulantRaw, 10) || 0,
        };
        records.push(record);
    });
    return records;
}

/**
 * Clears the details table.
 */
export function clearDetailsTable() {
    detailsTableBody.innerHTML = '';
}

/**
 * Populates the details table with loaded data.
 * @param {Array<object>} records - The transaction records to display.
 */
export function populateDetailsTable(records) {
    clearDetailsTable();
    let newHTML = '';

    records.forEach(rec => {
        const expiryDateFormatted = rec.expiryDate ? `${rec.expiryDate.slice(0, 4)}-${rec.expiryDate.slice(4, 6)}-${rec.expiryDate.slice(6, 8)}` : '';
        const flagMap = {1: "入庫", 2: "出庫"};

        // Recreate the two-row structure for each record, including the data-* attributes
        newHTML += `
            <tr 
              data-jan-pack-inner-qty="${rec.janPackInnerQty || 0}"
              data-kana-name="${rec.kanaName || ''}"
              data-package-form="${rec.packageForm || ''}"
              data-flag-poison="${rec.flagPoison || 0}"
              data-flag-deleterious="${rec.flagDeleterious || 0}"
              data-flag-narcotic="${rec.flagNarcotic || 0}"
              data-flag-psychotropic="${rec.flagPsychotropic || 0}"
              data-flag-stimulant="${rec.flagStimulant || 0}"
              data-flag-stimulant-raw="${rec.flagStimulantRaw || 0}"
            >
                <td rowspan="2">${rec.transactionDate || ''}</td>
                <td rowspan="2">${flagMap[rec.flag] || ''}</td>
                <td>${rec.yjCode || ''}</td>
                <td colspan="2" class="product-name-cell">${rec.productName || ''}</td>
                <td rowspan="2">${rec.datQuantity?.toFixed(2) || ''}</td>
                <td><input type="text" class="recalc-trigger jan-quantity-input" value="${rec.janQuantity || 0}"></td>
                <td>${rec.janPackUnitQty || ''}</td>
                <td>${rec.janUnitName || ''}</td>
                <td><input type="text" class="recalc-trigger unit-price-input" value="${rec.unitPrice?.toFixed(2) || 0}"></td>
                <td>${rec.taxAmount || 0}</td>
                <td><input type="date" value="${expiryDateFormatted}"></td>
                <td><input type="text" value="${rec.lotNumber || ''}"></td>
                <td><button class="delete-row-btn">削除</button></td>
            </tr>
            <tr>
                <td>${rec.janCode || ''}</td>
                <td>${rec.packageSpec || ''}</td>
                <td>${rec.makerName || ''}</td>
                <td>${rec.yjQuantity?.toFixed(2) || 0}</td>
                <td>${rec.yjPackUnitQty || ''}</td>
                <td>${rec.yjUnitName || ''}</td>
                <td>${rec.subtotal || 0}</td>
                <td>${rec.taxRate != null ? (rec.taxRate * 100).toFixed(0) + "%" : ""}</td>
                <td>${rec.clientCode || ''}</td>
                <td>${rec.receiptNumber || ''}</td>
                <td>${rec.lineNumber || ''}</td>
            </tr>
        `;
    });

    detailsTableBody.innerHTML = newHTML;
}