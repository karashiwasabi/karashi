// File: static/js/master_edit.js (Corrected and Finalized)
import { initModal, showModal } from './inout_modal.js';

// --- Global variables and DOM elements ---
const view = document.getElementById('master-edit-view');
const refreshBtn = document.getElementById('refreshMastersBtn');
const addMasterRowBtn = document.getElementById('addMasterRowBtn');
const tableBody = document.querySelector('#master-edit-table tbody');
const tableHead = document.querySelector('#master-edit-table thead');

let unitMap = {}; // TANI.CSVから読み込んだ単位情報を格納する

// ヘッダー定義
const upperHeaders = [
    { key: 'productCode', name: 'JC000(JANコード)', type: 'text' },
    { key: 'productName', name: 'JC018(製品名)', type: 'text' },
    { key: 'makerName', name: 'JC030(メーカー名)', type: 'text' },
    { key: 'janPackUnitQty', name: 'JA008(JAN包装数量)', type: 'number' },
    { key: 'janUnitCode', name: 'JA007(JAN単位)', type: 'select' },
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

const rowHeaderHTML = `
    <tr class="repeated-header">
        ${upperHeaders.map(h => `<th>${h.name}</th>`).join('')}
        <th rowspan="2">組み立てた包装</th>
        <th rowspan="2">操作</th>
    </tr>
    <tr class="repeated-header">
        ${lowerHeaders.map(h => `<th>${h.name}</th>`).join('')}
    </tr>
`;

/**
 * バックエンドから単位マスタを取得する
 */
async function fetchUnitMap() {
    try {
        const res = await fetch('/api/units/map');
        if (!res.ok) throw new Error('単位マスタの取得に失敗');
        unitMap = await res.json();
    } catch (err) {
        console.error(err);
        alert(err.message);
    }
}

/**
 * 組み立てた包装文字列を表示する (バグ修正版)
 * @param {HTMLTableRowElement} upperRow - レコードの上段の行
 */
function formatPackageSpecForRow(upperRow) {
    const lowerRow = upperRow.nextElementSibling;
    if (!lowerRow || !upperRow.hasAttribute('data-product-code')) return;

    const jc037 = lowerRow.querySelector('input[name="packageSpec"]').value;
    const jc044 = lowerRow.querySelector('input[name="yjPackUnitQty"]').value;
    const jc039_name = lowerRow.querySelector('input[name="yjUnitName"]').value;
    const ja006 = upperRow.querySelector('input[name="janPackInnerQty"]').value;
    const ja008 = upperRow.querySelector('input[name="janPackUnitQty"]').value;
    const ja007_code = upperRow.querySelector('select[name="janUnitCode"]').value;

    let pkg = `${jc037} ${jc044}${jc039_name}`;
    
    if (ja006 && ja008) {
        let janUnitName = '';
        if (ja007_code === '0' || !ja007_code) {
            janUnitName = jc039_name;
        } else {
            janUnitName = unitMap[ja007_code] || ja007_code;
        }
        pkg += ` (${ja006}${jc039_name}×${ja008}${janUnitName})`;
    }
    
    upperRow.querySelector('.package-spec-result').textContent = pkg;
}

/**
 * 1品目分のHTML（ヘッダー2行＋データ2行）を生成する (バグ修正・機能追加版)
 * @param {object} master - マスターデータレコード
 * @returns {string} - 4つの<tr>要素からなるHTML文字列
 */
function createMasterRowHTML(master = {}) {
    let upperHtml = `<tr data-product-code="${master.productCode || 'new'}">`;
    upperHeaders.forEach(h => {
        const value = master[h.key] ?? '';
        if (h.key === 'janUnitCode') {
            const displayUnitMap = { ...unitMap };
            if (!displayUnitMap['0']) {
                displayUnitMap['0'] = '(YJ単位と同じ)';
            }
            let options = '';
            for (const [code, name] of Object.entries(displayUnitMap)) {
                options += `<option value="${code}" ${code == value ? 'selected' : ''}>${code}: ${name}</option>`;
            }
            upperHtml += `<td><select name="${h.key}">${options}</select></td>`;
        } else if (h.type === 'select') {
            let options = h.options.map(o => `<option value="${o}" ${o == value ? 'selected' : ''}>${o}</option>`).join('');
            upperHtml += `<td><select name="${h.key}">${options}</select></td>`;
        } else {
            upperHtml += `<td><input type="${h.type}" name="${h.key}" value="${value}" step="any"></td>`;
        }
    });
    upperHtml += `<td rowspan="2" class="package-spec-result"></td>`;
    upperHtml += `<td rowspan="2">
                    <button class="save-master-row-btn">保存</button>
                    <button class="quote-jcshms-btn">JCSHMSから引用</button>
                  </td>`;
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

    return rowHeaderHTML + upperHtml + lowerHtml;
}

/**
 * マスター編集画面の表示をリセット（クリア）する
 */
export function resetMasterEditView() {
    if (tableBody) {
        tableBody.innerHTML = '';
    }
}

/**
 * DBからマスターデータを取得してテーブルに描画する
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
        tableBody.querySelectorAll('tr[data-product-code]').forEach(formatPackageSpecForRow);
    } catch (err) {
        console.error(err);
        tableBody.innerHTML = `<tr><td colspan="${upperHeaders.length + 2}" style="color:red;">${err.message}</td></tr>`;
    }
}

/**
 * JCSHMSからの引用データをフォームに反映させるコールバック関数
 * @param {object} selectedProduct - モーダルで選択された製品データ
 * @param {HTMLTableRowElement} activeRow - 引用ボタンが押された行（上段）
 */
function populateFormWithJcshms(selectedProduct, activeRow) {
    const lowerRow = activeRow.nextElementSibling;
    if (!lowerRow) return;

    // 製品コード(JAN)は元のレコードのものを維持するため、更新しない
    // activeRow.querySelector('[name="productCode"]').value = selectedProduct.productCode;

    // 上段のフォームを更新
    activeRow.querySelector('[name="productName"]').value = selectedProduct.productName || '';
    activeRow.querySelector('[name="makerName"]').value = selectedProduct.makerName || '';
    activeRow.querySelector('[name="janPackUnitQty"]').value = selectedProduct.janPackUnitQty || 0;
    activeRow.querySelector('[name="janUnitCode"]').value = selectedProduct.janUnitCode || 0;
    activeRow.querySelector('[name="janPackInnerQty"]').value = selectedProduct.janPackInnerQty || 0;
    activeRow.querySelector('[name="flagPoison"]').value = selectedProduct.flagPoison || 0;
    activeRow.querySelector('[name="flagDeleterious"]').value = selectedProduct.flagDeleterious || 0;
    activeRow.querySelector('[name="flagNarcotic"]').value = selectedProduct.flagNarcotic || 0;
    
    // 下段のフォームを更新
    lowerRow.querySelector('[name="yjCode"]').value = selectedProduct.yjCode || '';
    lowerRow.querySelector('[name="kanaName"]').value = selectedProduct.kanaName || '';
    lowerRow.querySelector('[name="packageSpec"]').value = selectedProduct.packageSpec || '';
    lowerRow.querySelector('[name="yjPackUnitQty"]').value = selectedProduct.yjPackUnitQty || 0;
    lowerRow.querySelector('[name="yjUnitName"]').value = selectedProduct.yjUnitName || '';
    lowerRow.querySelector('[name="nhiPrice"]').value = selectedProduct.nhiPrice || 0;
    lowerRow.querySelector('[name="flagPsychotropic"]').value = selectedProduct.flagPsychotropic || 0;
    lowerRow.querySelector('[name="flagStimulant"]').value = selectedProduct.flagStimulant || 0;
    lowerRow.querySelector('[name="flagStimulantRaw"]').value = selectedProduct.flagStimulantRaw || 0;

    // 包装表示を更新
    formatPackageSpecForRow(activeRow);
}

/**
 * イベントの発生源から、対応するデータ行（上段）を見つけ出すためのヘルパー関数
 * @param {EventTarget} target
 * @returns {HTMLTableRowElement|null}
 */
function getUpperDataRowFromTarget(target) {
    const anyRow = target.closest('tr');
    if (!anyRow) return null;
    // 自身がデータ行（上段）の場合
    if (anyRow.hasAttribute('data-product-code')) {
        return anyRow;
    }
    // 自身がデータ行（下段）の場合、兄要素（上段）を返す
    const prevRow = anyRow.previousElementSibling;
    if (prevRow && prevRow.hasAttribute('data-product-code')) {
        return prevRow;
    }
    return null;
}

/**
 * マスター編集画面の初期化処理
 */
export async function initMasterEdit() {
    if (!view) return;

    await fetchUnitMap();
    tableHead.innerHTML = '';
    initModal(populateFormWithJcshms);

    refreshBtn.addEventListener('click', loadAndRenderMasters);
    addMasterRowBtn.addEventListener('click', () => {
        tableBody.insertAdjacentHTML('beforeend', createMasterRowHTML());
    });
    
    tableBody.addEventListener('input', (event) => {
        const upperRow = getUpperDataRowFromTarget(event.target);
        if (upperRow) {
            formatPackageSpecForRow(upperRow);
        }
    });

    tableBody.addEventListener('click', async (event) => {
        const target = event.target;
        const upperRow = getUpperDataRowFromTarget(target);
        if (!upperRow) return;

        // 保存ボタンの処理
        if (target.classList.contains('save-master-row-btn')) {
            const lowerRow = upperRow.nextElementSibling;
            
            const data = { origin: 'MANUAL_ENTRY' };
            upperHeaders.forEach(h => data[h.key] = upperRow.querySelector(`[name="${h.key}"]`).value);
            lowerHeaders.forEach(h => data[h.key] = lowerRow.querySelector(`[name="${h.key}"]`).value);
            
            if (!data.productCode || data.productCode.trim() === '') {
                alert('製品コード(JAN)は必須です。新しいレコードを保存する前に入力してください。');
                return;
            }

            const floatFields = ['yjPackUnitQty', 'nhiPrice', 'reorderPoint', 'janPackInnerQty', 'janPackUnitQty'];
            const intFields = ['janUnitCode', 'flagPoison', 'flagDeleterious', 'flagNarcotic', 'flagPsychotropic', 'flagStimulant', 'flagStimulantRaw'];
            floatFields.forEach(key => data[key] = parseFloat(data[key]) || 0);
            intFields.forEach(key => data[key] = parseInt(data[key], 10) || 0);

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
        
        // JCSHMSから引用ボタンの処理
        if (target.classList.contains('quote-jcshms-btn')) {
            showModal(upperRow);
        }
    });

    loadAndRenderMasters();
}