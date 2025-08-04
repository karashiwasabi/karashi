// File: static/js/aggregation.js
import { transactionTypeMap } from './common_table.js';

const view = document.getElementById('aggregation-view');
const runBtn = document.getElementById('run-aggregation-btn');
const printBtn = document.getElementById('print-aggregation-btn');
const outputContainer = document.getElementById('aggregation-output-container');

// ★★★ 追加: 集計テーブルの列幅を定義 ★★★
const colWidths = [
    "5.83%", "4.5%", "9.15%", "13.77%", "13.77%",
    "2.91%", "5.34%", "7.69%", "5.34%", "7.21%",
    "5.91%", "7.21%", "7.21%", "5.75%"
];
const colgroup = `<colgroup>${
    colWidths.map(w => `<col style="width:${w};">`).join("")
}</colgroup>`;

function createAggregationTableHTML(tableId) {
  const header = `
    <thead>
      <tr>
        <th rowspan="2">日付</th><th rowspan="2">種別</th><th>YJ</th><th colspan="2">製品名</th>
        <th rowspan="2">個数</th><th>JAN数量</th><th>JAN包装数</th><th>JAN単位</th>
        <th>単価</th><th>税額</th><th>期限</th><th>ロット</th><th>MA</th>
      </tr>
      <tr>
        <th>JAN</th><th>包装</th><th>メーカー</th><th>YJ数量</th>
        <th>YJ包装数</th><th>YJ単位</th><th>金額</th><th>税率</th>
        <th>得意先</th><th>伝票番号</th><th>行</th>
      </tr>
    </thead>
  `;
  return `<table id="${tableId}" class="data-table">${colgroup}${header}<tbody>
    <tr><td colspan="14">フィルター条件を指定して「集計実行」を押してください。</td></tr>
  </tbody></table>`;
}

export function resetAggregationView() {
    if(outputContainer) {
        outputContainer.innerHTML = createAggregationTableHTML('aggregation-table');
    }
}

const startDateInput = document.getElementById('startDate');
const endDateInput = document.getElementById('endDate');
const kanaNameInput = document.getElementById('kanaName');
const drugTypeCheckboxes = document.querySelectorAll('input[name="drugType"]');
const noMovementCheckbox = document.getElementById('no-movement-filter');

const safeToFixed = (num, digits = 2) => (typeof num === 'number' ? num.toFixed(digits) : (0).toFixed(digits));

function renderResults(data) {
    if (!data || data.length === 0) {
        outputContainer.innerHTML = "<p>対象データが見つかりませんでした。</p>";
        return;
    }

    let html = '';
    data.forEach(yg => {
        yg.packageGroups.forEach(pg => {
            html += `<table class="aggregation-group-table">${colgroup}`; // ★★★ 修正: colgroupを追加

            const line1Parts = [
                `${yg.yjCode} ${yg.productName}`,
                `YJ数量 合計: ${safeToFixed(yg.totalYjQty)}`,
                `処方YJ数量 最大値: ${safeToFixed(yg.maxUsageYjQty)}`
            ];
            const line2Parts = [
                `${pg.packageKey}`,
                `JAN数量 合計: ${safeToFixed(pg.totalJanQty)}`,
                `処方JAN数量 最大値: ${safeToFixed(pg.maxUsageJanQty)}`,
                `YJ数量 合計: ${safeToFixed(pg.totalYjQty)}`,
                `処方YJ数量 最大値: ${safeToFixed(pg.maxUsageYjQty)}`
            ];
            
            html += `
                <thead class="repeating-header">
                    <tr>
                        <th colspan="14">
                            <div class="agg-header-line1">${line1Parts.join(' ')}</div>
                            <div class="agg-header-line2">${line2Parts.join(' ')}</div>
                        </th>
                    </tr>
                </thead>
            `;

            html += '<tbody>';
            if (pg.transactions && pg.transactions.length > 0) {
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
                pg.transactions.forEach(t => {
                    html += `
                        <tr>
                            <td rowspan="2">${t.transactionDate}</td><td rowspan="2">${transactionTypeMap[t.flag] || ''}</td><td>${t.yjCode}</td>
                            <td colspan="2" class="left">${t.productName}</td><td rowspan="2" class="right">${t.datQuantity}</td>
                            <td class="right">${safeToFixed(t.janQuantity)}</td><td class="right">${t.janPackUnitQty}</td><td>${t.janUnitName}</td>
                            <td class="right">${safeToFixed(t.unitPrice)}</td><td class="right">${t.taxAmount}</td><td>${t.expiryDate}</td>
                            <td>${t.lotNumber}</td><td>${t.processFlagMA}</td>
                        </tr>
                        <tr>
                            <td>${t.janCode}</td><td class="left">${t.packageSpec}</td><td class="left">${t.makerName}</td>
                            <td class="right">${safeToFixed(t.yjQuantity)}</td><td class="right">${t.yjPackUnitQty}</td><td>${t.yjUnitName}</td>
                            <td class="right">${t.subtotal}</td><td class="right">${t.taxRate != null ? (t.taxRate * 100).toFixed(0) + "%" : ""}</td>
                            <td class="left">${t.clientCode}</td><td class="left">${t.receiptNumber}</td><td class="right">${t.lineNumber}</td>
                        </tr>
                    `;
                });
            } else {
                 html += '<tr><td colspan="14" style="text-align:center; padding:10px;">このグループの明細はありません。</td></tr>';
            }
            html += '</tbody></table>';
        });
    });
    outputContainer.innerHTML = html;
}

export function initAggregation() {
    if (!view) return;
    
    const today = new Date();
    const fourMonthsAgo = new Date();
    fourMonthsAgo.setMonth(today.getMonth() - 4);
    endDateInput.value = today.toISOString().slice(0, 10);
    startDateInput.value = fourMonthsAgo.toISOString().slice(0, 10);

    resetAggregationView();

    if(printBtn) {
        printBtn.addEventListener('click', () => {
            window.print();
        });
    }

    runBtn.addEventListener('click', async () => {
        window.showLoading(); // ローディング開始
        
        const params = new URLSearchParams();
        params.append('startDate', startDateInput.value.replace(/-/g, ''));
        params.append('endDate', endDateInput.value.replace(/-/g, ''));
        if (kanaNameInput.value) {
            params.append('kanaName', kanaNameInput.value);
        }
        const selectedTypes = Array.from(drugTypeCheckboxes)
            .filter(cb => cb.checked)
            .map(cb => cb.value);
        if (selectedTypes.length > 0) {
            params.append('drugTypes', selectedTypes.join(','));
        }
        if (noMovementCheckbox.checked) {
            params.append('noMovement', 'true');
        }

        try {
            const res = await fetch(`/api/aggregation?${params.toString()}`);
            if (!res.ok) throw new Error('集計に失敗しました');
            const data = await res.json();
            renderResults(data);
        } catch (err) {
            outputContainer.innerHTML = `<p style="color:red;">${err.message}</p>`;
        } finally {
            window.hideLoading(); // ローディング終了
        }
    });
}
