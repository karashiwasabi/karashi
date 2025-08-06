// File: static/js/month_end.js (新規作成)
import { createUploadTableHTML, renderUploadTableRows } from './common_table.js';

let view, monthSelect, runBtn, outputContainer;

// 実行ボタンが押された時の処理
async function executeCalculation() {
    const selectedMonth = monthSelect.value;
    if (!selectedMonth) {
        alert('対象年月を選択してください。');
        return;
    }

    if (!confirm(`${selectedMonth} の月末在庫を計算します。\n(同月の既存データは上書きされます)`)) {
        return;
    }

    outputContainer.innerHTML = createUploadTableHTML('month-end-output-table');
    const tbody = document.querySelector('#month-end-output-table tbody');
    tbody.innerHTML = `<tr><td colspan="14" style="text-align:center;">${selectedMonth}の月末在庫を計算中...</td></tr>`;

    window.showLoading();
    try {
        const res = await fetch('/api/inventory/calculate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ month: selectedMonth }),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.message || '計算に失敗しました。');

        window.showNotification(data.message, 'success');
        renderUploadTableRows('month-end-output-table', data.records);
    } catch (err) {
        window.showNotification(`エラー: ${err.message}`, 'error');
        tbody.innerHTML = `<tr><td colspan="14" style="color:red;text-align:center;">${err.message}</td></tr>`;
    } finally {
        window.hideLoading();
    }
}

// プルダウンメニューを初期化する
async function populateMonthDropdown() {
    monthSelect.innerHTML = `<option value="">読込中...</option>`;
    outputContainer.innerHTML = '';
    try {
        const res = await fetch('/api/inventory/months');
        if (!res.ok) throw new Error('対象月の取得に失敗');
        const months = await res.json();

        if (!months || months.length === 0) {
            monthSelect.innerHTML = `<option value="">対象月なし</option>`;
            return;
        }

        monthSelect.innerHTML = `<option value="">選択してください</option>`;
        months.forEach(month => {
            const opt = document.createElement('option');
            opt.value = month;
            opt.textContent = month;
            monthSelect.appendChild(opt);
        });
    } catch (err) {
        monthSelect.innerHTML = `<option value="">取得失敗</option>`;
        console.error(err);
    }
}

// 月末在庫ビューの初期化
export function initMonthEndView() {
    view = document.getElementById('month-end-view');
    if (!view) return;

    monthSelect = document.getElementById('month-end-select');
    runBtn = document.getElementById('run-month-end-btn');
    outputContainer = document.getElementById('month-end-output-container');

    runBtn.addEventListener('click', executeCalculation);
}

// ビューが表示されるたびに呼ばれるリセット/初期化関数
export function resetMonthEndView() {
    populateMonthDropdown();
}