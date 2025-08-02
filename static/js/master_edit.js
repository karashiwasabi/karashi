// File: static/js/master_edit.js (Corrected)

// --- Global variables and DOM elements ---
const view = document.getElementById('master-edit-view');
const refreshBtn = document.getElementById('refreshMastersBtn');
const addMasterRowBtn = document.getElementById('addMasterRowBtn');
const tableHead = document.querySelector('#master-edit-table thead');
const tableBody = document.querySelector('#master-edit-table tbody');

// TANI.CSV data replacement
const taniMap = {
    "11": "包", "13": "本", "30": "錠", "33": "ｇ", "34": "ｍＬ", "43": "個",
};

// Header definitions
const upperHeaders = [
    { key: 'productCode', name: 'JC000(JANコード)', type: 'text' },
    { key: 'productName', name: 'JC018(製品名)', type: 'text' },
    { key: 'makerName', name: 'JC030(メーカー名)', type: 'text' },
    { key: 'janPackUnitQty', name: 'JA008(JAN包装数量)', type: 'number' },
    { key: 'janUnitCode', name: 'JA007(JAN単位)', type: 'text' },
    { key: 'janPackInnerQty', name: 'JA006(内包装数量)', type: 'number' },
    { key: 'flagPoison', name: 'JC061(毒薬)', type: 'select', options: [0, 1] },
    { key: 'flagDeleterious', name: 'JC062(劇薬)', type: 'select', options: [0, 1] },
    { key: 'flagNarcotic', name: 'JC063(麻薬)', type: 'select', options: [0, 1] },
];
const lowerHeaders = [
    { key: 'yjCode', name: 'JC009(YJコード)', type: 'text' },
    { key: 'kanaName', name: 'JC022(カナ名)', type: 'text' },
    { key: 'packageSpec', name: 'JC037(包装)', type: 'text' },
    { key: 'yjPackUnitQty', name: 'JC044(YJ包装数量)', type: 'number' },
    { key: 'yjUnitName', name: 'JC039(YJ単位)', type: 'text' },
    { key: 'nhiPrice', name: 'JC050(薬価)', type: 'number' },
    { key: 'flagPsychotropic', name: 'JC064(向精神薬)', type: 'select', options: [0, 1, 2, 3] },
    { key: 'flagStimulant', name: 'JC065(覚醒剤)', type: 'select', options: [0, 1] },
    { key: 'flagStimulantRaw', name: 'JC066(覚醒剤原料)', type: 'select', options: [0, 1] },
];

/**
 * Resolves a unit code to its Japanese name.
 */
function resolveTaniName(code) {
    return taniMap[code] || code;
}

/**
 * Assembles and displays the packaging string for a given row.
 * @param {HTMLTableRowElement} upperRow - The upper row of the record.
 */
function formatPackageSpecForRow(upperRow) {
    const lowerRow = upperRow.nextElementSibling;
    if (!lowerRow) return;

    const jc037 = lowerRow.querySelector('input[name="packageSpec"]').value;
    const jc044 = lowerRow.querySelector('input[name="yjPackUnitQty"]').value;
    const jc039_code = lowerRow.querySelector('input[name="yjUnitName"]').value;
    const ja006 = upperRow.querySelector('input[name="janPackInnerQty"]').value;
    const ja008 = upperRow.querySelector('input[name="janPackUnitQty"]').value;
    const ja007_code = upperRow.querySelector('input[name="janUnitCode"]').value;

    const yjUnitName = resolveTaniName(jc039_code);
    let pkg = `${jc037} ${jc044}${yjUnitName}`;
    
    if (ja006 && ja008) {
        let janUnitName = '';
        if (ja007_code && ja007_code !== '0') {
            janUnitName = resolveTaniName(ja007_code);
        }
        pkg += ` (${ja006}${yjUnitName}×${ja008}${janUnitName})`;
    }
    
    upperRow.querySelector('.package-spec-result').textContent = pkg;
}

/**
 * Creates the HTML for a two-row master record.
 * @param {object} master - The master data record.
 * @returns {string} - The HTML string for the two table rows.
 */
function createMasterRowHTML(master = {}) {
    let upperHtml = '<tr>';
    upperHeaders.forEach(h => {
        const value = master[h.key] ?? ''; // Use ?? to handle null/undefined
        if (h.type === 'select') {
            let options = h.options.map(o => `<option value="${o}" ${o == value ? 'selected' : ''}>${o}</option>`).join('');
            upperHtml += `<td><select name="${h.key}">${options}</select></td>`;
        } else {
            upperHtml += `<td><input type="${h.type}" name="${h.key}" value="${value}" step="any"></td>`;
        }
    });
    upperHtml += `<td rowspan="2" class="package-spec-result"></td>`;
    upperHtml += `<td rowspan="2"><button class="save-master-row-btn">保存</button></td>`;
    upperHtml += '</tr>';

    let lowerHtml = '<tr>';
    lowerHeaders.forEach(h => {
        const value = master[h.key] ?? '';
        if (h.type === 'select') {
             let options = h.options.map(o => `<option value="${o}" ${o == value ? 'selected' : ''}>${o}</option>`).join('');
            lowerHtml += `<td><select name="${h.key}">${options}</select></td>`;
        } else {
            lowerHtml += `<td><input type="${h.type}" name="${h.key}" value="${value}" step="any"></td>`;
        }
    });
    lowerHtml += '</tr>';

    return upperHtml + lowerHtml;
}


// ▼▼▼ 修正点: リセット関数をエクスポート ▼▼▼
export function resetMasterEditView() {
    if (tableBody) {
        tableBody.innerHTML = '';
    }
}

/**
 * Fetches and renders master data into the table.
 */
async function loadAndRenderMasters() {
    tableBody.innerHTML = `<tr><td colspan="${upperHeaders.length + 2}">読み込み中...</td></tr>`;
    try {
        const res = await fetch('/api/masters/editable');
        if (!res.ok) throw new Error('マスターデータの取得に失敗しました。');
        const masters = await res.json() || [];

        if (masters.length === 0) {
            tableBody.innerHTML = `<tr><td colspan="${upperHeaders.length + 2}">対象のマスターデータはありません。</td></tr>`;
            return;
        }

        tableBody.innerHTML = masters.map(createMasterRowHTML).join('');
        tableBody.querySelectorAll('tr:nth-child(odd)').forEach(formatPackageSpecForRow);

    } catch (err) {
        console.error(err);
        tableBody.innerHTML = `<tr><td colspan="${upperHeaders.length + 2}" style="color:red;">${err.message}</td></tr>`;
    }
}

/**
 * Initializes the master edit view.
 */
export function initMasterEdit() {
    if (!view) return;

    let headerHtml = '<tr>';
    upperHeaders.forEach(h => headerHtml += `<th>${h.name}</th>`);
    headerHtml += `<th rowspan="2">組み立てた包装</th><th rowspan="2">保存</th></tr>`;
    headerHtml += '<tr>';
    lowerHeaders.forEach(h => headerHtml += `<th>${h.name}</th>`);
    headerHtml += '</tr>';
    tableHead.innerHTML = headerHtml;

    refreshBtn.addEventListener('click', loadAndRenderMasters);
    
    addMasterRowBtn.addEventListener('click', () => {
        tableBody.insertAdjacentHTML('beforeend', createMasterRowHTML());
    });

    tableBody.addEventListener('input', (event) => {
        if (event.target.tagName === 'INPUT' || event.target.tagName === 'SELECT') {
            const upperRow = event.target.closest('tr:nth-child(odd), tr:nth-child(even)').previousElementSibling || event.target.closest('tr');
            formatPackageSpecForRow(upperRow);
        }
    });

    tableBody.addEventListener('click', async (event) => {
        if (event.target.classList.contains('save-master-row-btn')) {
            const upperRow = event.target.closest('tr');
            const lowerRow = upperRow.nextElementSibling;
            
            const data = { origin: 'MANUAL_ENTRY' };
            upperHeaders.forEach(h => data[h.key] = upperRow.querySelector(`[name="${h.key}"]`).value);
            lowerHeaders.forEach(h => data[h.key] = lowerRow.querySelector(`[name="${h.key}"]`).value);
            
            if (!data.productCode || data.productCode.trim() === '') {
                alert('製品コード(JAN)は必須です。新しいレコードを保存する前に入力してください。');
                return;
            }

            // *** CORRECTED: Explicitly convert all numeric fields ***
            const floatFields = ['yjPackUnitQty', 'nhiPrice', 'reorderPoint', 'janPackInnerQty', 'janPackUnitQty'];
            const intFields = ['janUnitCode', 'flagPoison', 'flagDeleterious', 'flagNarcotic', 'flagPsychotropic', 'flagStimulant', 'flagStimulantRaw'];

            floatFields.forEach(key => {
                data[key] = parseFloat(data[key]) || 0;
            });
            intFields.forEach(key => {
                data[key] = parseInt(data[key], 10) || 0;
            });
            // *** END CORRECTION ***

            try {
                const res = await fetch('/api/master/update', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data),
                });
                const resData = await res.json();
                if (!res.ok) throw new Error(resData.message || '保存に失敗しました。');
                
                alert(resData.message);
                loadAndRenderMasters();

            } catch (err) {
                console.error(err);
                alert(`エラー: ${err.message}`);
            }
        }
    });

    // ▼▼▼ 修正点: 起動時の読み込みをリセット関数で行う ▼▼▼
    resetMasterEditView();
    loadAndRenderMasters(); 

}