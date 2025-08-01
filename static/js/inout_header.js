// File: static/js/inout_header.js (Corrected)
import { setupDateDropdown, setupClientDropdown } from './common_table.js';

const NEW_ENTRY_VALUE = '--new--';

let clientSelect, receiptSelect, saveBtn, deleteBtn, headerDateInput, headerTypeSelect;
let newClientName = null;
let currentLoadedReceipt = null; // Variable to track the loaded slip

async function initializeClientDropdown() {
    clientSelect.innerHTML = `<option value="">選択してください</option>`;
    await setupClientDropdown(clientSelect);
    
    const newOption = document.createElement('option');
    newOption.value = NEW_ENTRY_VALUE;
    newOption.textContent = '--- 新規作成 ---';
    clientSelect.appendChild(newOption);
}

export async function initHeader(getDetailsData, clearDetailsTable, populateDetailsTable) {
    clientSelect = document.getElementById('in-out-client');
    receiptSelect = document.getElementById('in-out-receipt');
    saveBtn = document.getElementById('saveBtn');
    deleteBtn = document.getElementById('deleteBtn');
    headerDateInput = document.getElementById('in-out-date');
    headerTypeSelect = document.getElementById('in-out-type');

    if (!clientSelect || !receiptSelect || !saveBtn || !deleteBtn) return;
    deleteBtn.disabled = true;

    setupDateDropdown(headerDateInput);
    await initializeClientDropdown();
    receiptSelect.innerHTML = `
        <option value="">日付を選択してください</option>
        <option value="${NEW_ENTRY_VALUE}">--- 新規作成 ---</option>
    `;
    
    headerDateInput.addEventListener('change', async () => {
        const date = headerDateInput.value.replace(/-/g, '');
        if (!date) return;

        try {
            const res = await fetch(`/api/receipts?date=${date}`);
            if (!res.ok) throw new Error('伝票の取得に失敗');
            const receiptNumbers = await res.json();
            
            receiptSelect.innerHTML = `
                <option value="">選択してください</option>
                <option value="${NEW_ENTRY_VALUE}">--- 新規作成 ---</option>
            `;

            if (receiptNumbers && receiptNumbers.length > 0) {
                receiptNumbers.forEach(num => {
                    const opt = document.createElement('option');
                    opt.value = num;
                    opt.textContent = num;
                    receiptSelect.appendChild(opt);
                });
            }
        } catch (err) {
            console.error(err);
            receiptSelect.innerHTML = `
                <option value="">選択してください</option>
                <option value="${NEW_ENTRY_VALUE}">--- 新規作成 ---</option>
            `;
        }
    });

    clientSelect.addEventListener('change', () => {
        const selectedValue = clientSelect.value;
        if (selectedValue === NEW_ENTRY_VALUE) {
            const name = prompt('新しい得意先名を入力してください:');
            if (name && name.trim()) {
                newClientName = name.trim();
                const opt = document.createElement('option');
                opt.value = `new:${newClientName}`;
                opt.textContent = `[新規] ${newClientName}`;
                opt.selected = true;
                clientSelect.appendChild(opt);
            } else {
                clientSelect.value = '';
            }
        } else if (!selectedValue.startsWith('new:')) {
            newClientName = null;
        }
    });

    receiptSelect.addEventListener('change', async () => {
        const selectedValue = receiptSelect.value;
        deleteBtn.disabled = (selectedValue === NEW_ENTRY_VALUE || selectedValue === "");

        if (selectedValue === NEW_ENTRY_VALUE || selectedValue === "") {
            clearDetailsTable();
            currentLoadedReceipt = null; // Clear loaded receipt state
        } else {
            try {
                const res = await fetch(`/api/transaction/${selectedValue}`);
                if (!res.ok) throw new Error('明細の読込に失敗');
                const records = await res.json();

                if (records && records.length > 0) {
                    currentLoadedReceipt = selectedValue; // Set the loaded receipt number
                    const clientCode = records[0].clientCode;
                    clientSelect.value = clientCode;
                    newClientName = null;
                }
                
                populateDetailsTable(records);
            } catch (err) {
                console.error(err);
                alert(err.message);
            }
        }
    });
    
    saveBtn.addEventListener('click', async () => {
        let clientCode = clientSelect.value;
        let clientNameToSave = '';
        let isNewClient = false;

        if (newClientName && clientCode.startsWith('new:')) {
            clientNameToSave = newClientName;
            isNewClient = true;
            clientCode = '';
        } else {
            if (!clientCode || clientCode === NEW_ENTRY_VALUE) {
                alert('得意先を選択または新規作成してください。');
                return;
            }
        }
        
        const records = getDetailsData();
        if (records.length === 0) {
            alert('保存する明細データがありません。');
            return;
        }
        
        const payload = {
            isNewClient: isNewClient,
            clientCode: clientCode,
            clientName: clientNameToSave,
            transactionDate: headerDateInput.value.replace(/-/g, ''),
            transactionType: headerTypeSelect.value,
            records: records,
            originalReceiptNumber: currentLoadedReceipt // Send the original receipt number
        };

        try {
            const res = await fetch('/api/inout/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            const resData = await res.json();
            if (!res.ok) {
                throw new Error(resData.message || `保存に失敗しました (HTTP ${res.status})`);
            }

            alert(`データを保存しました。\n伝票番号: ${resData.receiptNumber}`);
            
            // Reset UI
            clearDetailsTable();
            await initializeClientDropdown();
            newClientName = null;
            currentLoadedReceipt = null;
            headerDateInput.dispatchEvent(new Event('change'));
            receiptSelect.value = NEW_ENTRY_VALUE;
            deleteBtn.disabled = true;

        } catch (err) {
            console.error(err);
            alert(err.message);
        }
    });

    deleteBtn.addEventListener('click', async () => {
        const receiptNumber = receiptSelect.value;
        if (!receiptNumber || receiptNumber === NEW_ENTRY_VALUE) {
            alert("削除対象の伝票が選択されていません。");
            return;
        }

        if (!confirm(`伝票番号 [${receiptNumber}] を完全に削除します。よろしいですか？`)) {
            return;
        }

        try {
            const res = await fetch(`/api/transaction/delete/${receiptNumber}`, {
                method: 'DELETE',
            });
            
            const errData = await res.json().catch(() => null);
            if (!res.ok) {
                throw new Error(errData?.message || '削除に失敗しました。');
            }

            alert(`伝票 [${receiptNumber}] を削除しました。`);
            
            clearDetailsTable();
            await initializeClientDropdown();
            receiptSelect.innerHTML = `
                <option value="">日付を選択してください</option>
                <option value="${NEW_ENTRY_VALUE}" selected>--- 新規作成 ---</option>
            `;
            newClientName = null;
            deleteBtn.disabled = true;
            headerDateInput.dispatchEvent(new Event('change'));

        } catch(err) {
            console.error(err);
            alert(err.message);
        }
    });

    headerDateInput.dispatchEvent(new Event('change'));
}