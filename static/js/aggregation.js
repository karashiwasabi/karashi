// File: static/js/aggregation.js (Final Integrated Version)
import { transactionTypeMap } from './common_table.js';
const view = document.getElementById('aggregation-view');
const runBtn = document.getElementById('run-aggregation-btn');
const printBtn = document.getElementById('print-aggregation-btn');
const outputContainer = document.getElementById('aggregation-output-container');
const coefficientInput = document.getElementById('reorder-coefficient');
const neededOnlyCheckbox = document.getElementById('reorder-needed-filter');
let lastData = []; // フィルター用に最後の結果を保持

const colWidths = [
    "5.83%", "4.5%", "9.15%", "13.77%", "13.77%",
    "2.91%", "5.34%", "7.69%", "5.34%", "7.21%",
    "5.91%", "7.21%", "7.21%", "5.75%"
];
const colgroup = `<colgroup>${colWidths.map(w => `<col style="width:${w};">`).join("")}</colgroup>`;

const safeToFixed = (num, digits = 2) => (typeof num === 'number' ? num.toFixed(digits) : (0).toFixed(digits));
function renderResults(data) {
    lastData = data;
    if (!data || data.length === 0) {
        outputContainer.innerHTML = "<p>対象データが見つかりませんでした。</p>";
        return;
    }

    const isFiltered = neededOnlyCheckbox.checked;
    let html = '';
    data.forEach((yjGroup, yjIndex) => {
        if (isFiltered && !yjGroup.isReorderNeeded) {
            return;
        }

        const yjAlertClass = yjGroup.endingBalance < yjGroup.totalReorderPoint ? 'reorder-yj-alert' : '';

        // ▼▼▼ 大グループヘッダーの生成ロジックを修正 ▼▼▼
        const yjUnit = yjGroup.yjUnitName || '';
        const yjHeaderText = `${yjGroup.yjCode} ${yjGroup.productName}` +
            `期首在庫: ${safeToFixed(yjGroup.startingBalance)}${yjUnit} | ` +
            `変動数量: ${safeToFixed(yjGroup.netChange)}${yjUnit} | ` +
            `期末在庫: ${safeToFixed(yjGroup.endingBalance)}${yjUnit} | ` +
            `発注点: ${safeToFixed(yjGroup.totalReorderPoint)}${yjUnit}`;

        html += `<table class="aggregation-group-table">${colgroup}`;
        html += `
            <thead class="repeating-header">
                <tr class="${yjAlertClass}">
                    <th colspan="14">
                        <div class="agg-header-line1" style="font-weight: bold;">${yjHeaderText}</div>
                    </th>
                </tr>
            </thead>`;
        // ▲▲▲ ここまで修正 ▲▲▲
            
        yjGroup.packageLedgers.forEach((pkg, pkgIndex) => {
            if (isFiltered && !pkg.isReorderNeeded) {
                return;
            }

            const accordionId = `accordion-${yjIndex}-${pkgIndex}`;
            const innerQty = (pkg.transactions && pkg.transactions.length > 0 && pkg.transactions[0].janPackInnerQty) ? pkg.transactions[0].janPackInnerQty : 1;
            const pkgStartingBalanceJAN = pkg.startingBalance / innerQty;
            const pkgNetChangeJAN = pkg.netChange / innerQty;
            const pkgEndingBalanceJAN = pkg.endingBalance / innerQty;
            const pkgAlertClass = pkg.isReorderNeeded ? 'reorder-pkg-alert' : '';
            
            // ▼▼▼ 小グループヘッダーの生成ロジックを修正 ▼▼▼
            const yjUnit = yjGroup.yjUnitName || '';
            const janUnit = pkg.janUnitName || '';
            const keyParts = pkg.packageKey.split('|');

            const specDisplay = `仕様: ${keyParts[0]} ${keyParts[1]} ${keyParts[2]}`;
            const balancesDisplay = `期首在庫: ${safeToFixed(pkgStartingBalanceJAN)}${janUnit} | ` +
                `変動数量: ${safeToFixed(pkgNetChangeJAN)}${janUnit} | ` +
                `期末在庫: ${safeToFixed(pkgEndingBalanceJAN)}${janUnit}`;
            const reorderDisplay = `最大処方:${safeToFixed(pkg.maxUsage)}${yjUnit} | ` +
                `発注点:${safeToFixed(pkg.reorderPoint)}${yjUnit} ${pkg.isReorderNeeded ? '●' : ''}`;
            const pkgHeaderText = `${specDisplay} | ${balancesDisplay} | ${reorderDisplay}`;

            html += `
              <tbody class="accordion-group">
                <tr class="details-header accordion-trigger ${pkgAlertClass}" data-target="#${accordionId}">
                    <th colspan="14" style="text-align:left; background-color: #f7f7f7;">
                       <div class="agg-header-line2">${pkgHeaderText}</div>
                    </th>
                </tr>
              </tbody>
            `;
            // ▲▲▲ ここまで修正 ▲▲▲
            
            html += `<tbody id="${accordionId}" class="accordion-content hidden">`;
            html += `
                <tr class="details-header">
                    <th rowspan="2">日付</th><th rowspan="2">種別</th><th>YJ</th><th colspan="2">製品名</th>
                    <th rowspan="2">個数</th><th>JAN数量</th><th>JAN包装数</th><th>JAN単位</th>
                    <th>単価</th><th>税額</th><th>期限</th><th>ロット</th><th>MA</th>
                </tr>
                <tr class="details-header">
                    <th>JAN</th><th>包装</th><th>メーカー</th><th>YJ数量</th>
                    <th>YJ包装数</th><th>YJ単位</th><th>金額</th><th>税率</th>
                    <th>得意先</th><th>伝票番号</th><th>行</th>
                </tr>
            `;
            pkg.transactions.forEach(t => {
                html += `
                    <tr>
                        <td rowspan="2">${t.transactionDate}</td>
                        <td rowspan="2">${transactionTypeMap[t.flag] || ''}</td>
                        <td>${t.yjCode}</td>
                        <td colspan="2" class="left">${t.productName}</td>
                        <td rowspan="2" class="right">${t.datQuantity}</td>
                        <td class="right">${safeToFixed(t.janQuantity)}</td>
                        <td class="right">${t.janPackUnitQty}</td>
                        <td>${t.janUnitName}</td>
                        <td class="right">${safeToFixed(t.unitPrice, 4)}</td>
                        <td class="right">${t.taxAmount}</td>
                        <td>${t.expiryDate}</td>
                        <td>${t.lotNumber}</td>
                        <td>${t.processFlagMA}</td>
                    </tr>
                    <tr>
                        <td>${t.janCode}</td>
                        <td class="left">${t.packageSpec}</td>
                        <td class="left">${t.makerName}</td>
                        <td class="right">${safeToFixed(t.yjQuantity)}</td>
                        <td class="right">${t.yjPackUnitQty}</td>
                        <td>${t.yjUnitName}</td>
                        <td class="right">${t.subtotal}</td>
                        <td class="right">${t.taxRate != null ? (t.taxRate * 100).toFixed(0) + "%" : ""}</td>
                        <td class="left">${t.clientCode}</td>
                        <td class="left">${t.receiptNumber}</td>
                        <td class="right">${t.lineNumber}</td>
                    </tr>
                `;
            });
            html += `</tbody>`;
        });
        html += '</table>';
    });
    outputContainer.innerHTML = html;
}

export function resetAggregationView() {
    if (outputContainer) {
        outputContainer.innerHTML = `<p>フィルター条件を指定して「集計実行」を押してください。</p>`;
    }
    lastData = [];
}

export function initAggregation() {
    if (!view) return;
    const startDateInput = document.getElementById('startDate');
    const endDateInput = document.getElementById('endDate');
    const kanaNameInput = document.getElementById('kanaName');
    const drugTypeCheckboxes = document.querySelectorAll('input[name="drugType"]');
    const today = new Date();
    const threeMonthsAgo = new Date();
    threeMonthsAgo.setMonth(today.getMonth() - 3);
    endDateInput.value = today.toISOString().slice(0, 10);
    startDateInput.value = threeMonthsAgo.toISOString().slice(0, 10);

    resetAggregationView();
    if (printBtn) {
        printBtn.addEventListener('click', () => window.print());
    }

    runBtn.addEventListener('click', async () => {
        window.showLoading();
        
        const params = new URLSearchParams();
        params.append('startDate', startDateInput.value.replace(/-/g, ''));
        params.append('endDate', endDateInput.value.replace(/-/g, ''));
        params.append('coefficient', coefficientInput.value);

        if (kanaNameInput.value) {
            params.append('kanaName', kanaNameInput.value);
        }
        const selectedTypes = Array.from(drugTypeCheckboxes)
            .filter(cb => cb.checked)
            .map(cb => cb.value);
        if (selectedTypes.length > 0) {
            params.append('drugTypes', selectedTypes.join(','));
        }
        
        try {
            const res = await fetch(`/api/aggregation?${params.toString()}`);
            if (!res.ok) throw new Error('集計に失敗しました');
            const data = await res.json();
            renderResults(data);
        } catch (err) {
            outputContainer.innerHTML = `<p style="color:red;">${err.message}</p>`;
        } finally {
            window.hideLoading();
        }
    });
    neededOnlyCheckbox.addEventListener('change', () => {
        renderResults(lastData);
    });
    outputContainer.addEventListener('click', (e) => {
        const trigger = e.target.closest('.accordion-trigger');
        if (!trigger) return;

        const targetId = trigger.dataset.target;
        const content = document.querySelector(targetId);
        if (content) {
            content.classList.toggle('hidden');
            trigger.classList.toggle('expanded');
        }
    });
}