// File: static/js/common_table.js

// ▼▼▼ transactionTypeMap に「0: "棚卸"」を追加 ▼▼▼
// ▼▼▼ transactionTypeMap に「30: "月末在庫"」を追加 ▼▼▼
export const transactionTypeMap = {
    0: "棚卸",
    1: "納品",
    2: "返品",
    3: "処方",
    4: "棚卸増",
    5: "棚卸減",
    11: "入庫",
    12: "出庫",
    30: "月末", // 追記
};
/**
 * アップロード結果表示用テーブルの HTML を生成
 * @param {string} tableId テーブル要素に付与する ID
 * @returns {string} テーブル HTML
 */
export function createUploadTableHTML(tableId) {
  const colWidths = [
    "5.83%", "2.91%", "9.15%", "13.77%", "13.77%",
    "2.91%", "5.34%", "7.69%", "5.34%", "7.21%",
    "5.91%", "7.21%", "7.21%", "5.75%"
  ];
  const colgroup = `<colgroup>${
    colWidths.map(w => `<col style="width:${w};">`).join("")
  }</colgroup>`;
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
    <tr><td colspan="14">ファイルを選択してください。</td></tr>
  </tbody></table>`;
}

/**
 * アップロード結果レコードをテーブルに描画
 * @param {string} tableId 対象テーブルの ID
 * @param {Array<object>} records サーバーから返却されたレコード配列
 */
export function renderUploadTableRows(tableId, records) {
  const tbody = document.querySelector(`#${tableId} tbody`);
  if (!records || records.length === 0) {
    tbody.innerHTML = `<tr><td colspan="14">対象データがありません。</td></tr>`;
    return;
  }
  
  let html = "";
  records.forEach(rec => {
    html += `
      <tr>
        <td rowspan="2">${rec.transactionDate || ""}</td>
        <td rowspan="2">${transactionTypeMap[rec.flag] ?? ""}</td>
        <td>${rec.yjCode || ""}</td>
        <td class="left" colspan="2">${rec.productName || ""}</td>
        <td class="right" rowspan="2">${rec.datQuantity?.toFixed(2) || ""}</td>
        <td class="right">${rec.janQuantity?.toFixed(2) || ""}</td>
        <td class="right">${rec.janPackUnitQty || ""}</td>
        
        <td>${rec.janUnitName || ""}</td>
        <td class="right">${rec.unitPrice?.toFixed(2) || ""}</td>
        <td class="right">${rec.taxAmount?.toFixed(2) || ""}</td>
        <td>${rec.expiryDate || ""}</td>
        <td class="left">${rec.lotNumber || ""}</td>
        <td class="left">${rec.processFlagMA || ""}</td>
      </tr>
      <tr>
        <td>${rec.janCode || ""}</td>
        <td class="left">${rec.packageSpec || ""}</td>
        <td class="left">${rec.makerName || ""}</td>
        <td class="right">${rec.yjQuantity?.toFixed(2) || ""}</td>
        <td class="right">${rec.yjPackUnitQty || ""}</td>
        <td>${rec.yjUnitName || ""}</td>
        <td class="right">${rec.subtotal?.toFixed(2) || ""}</td>
        <td class="right">${rec.taxRate != null ? (rec.taxRate * 100).toFixed(0) + "%" : ""}</td>
        <td class="left">${rec.clientCode || ""}</td>
        <td class="left">${rec.receiptNumber || ""}</td>
        <td class="right">${rec.lineNumber || ""}</td>
      </tr>
    `;
  });
  tbody.innerHTML = html;
}

export function setupDateDropdown(inputEl) {
  if (!inputEl) return;
  inputEl.value = new Date().toISOString().slice(0, 10);
}

/**
 * APIから得意先リストを取得してプルダウンに設定
 * @param {HTMLSelectElement} selectEl
 */
export async function setupClientDropdown(selectEl) {
  if (!selectEl) return;
  const preservedOptions = Array.from(selectEl.querySelectorAll('option[value=""]'));
  selectEl.innerHTML = '';
  preservedOptions.forEach(opt => selectEl.appendChild(opt));
  try {
    const res = await fetch('/api/clients');
    if (!res.ok) throw new Error('Failed to fetch clients');
    const clients = await res.json();

    if (clients) {
      clients.forEach(c => {
        const opt = document.createElement('option');
        opt.value = c.code;
        opt.textContent = `${c.code}:${c.name}`;
        selectEl.appendChild(opt);
      });
    }
  } catch (err) {
    console.error("得意先リストの取得に失敗:", err);
  }
}