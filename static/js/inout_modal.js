// File: static/js/inout_modal.js (修正後・完全版)

// --- モジュールスコープ変数 ---
let onProductSelectCallback = null; // 製品選択時に呼び出されるコールバック関数
let activeRowElement = null; // モーダルを呼び出したアクティブな行要素

// --- DOM要素の取得 ---
const modal = document.getElementById('search-modal');
const closeModalBtn = document.getElementById('closeModalBtn');
const searchInput = document.getElementById('product-search-input');
const searchBtn = document.getElementById('product-search-btn');
const searchResultsBody = document.querySelector('#search-results-table tbody');

/**
 * 検索結果テーブルの「選択」ボタンクリックを処理（イベント委譲）
 * @param {Event} event
 */
function handleResultClick(event) {
  // クリックされたのが選択ボタンか確認
  if (event.target && event.target.classList.contains('select-product-btn')) {
    const product = JSON.parse(event.target.dataset.product);
    // コールバック関数が設定されていれば、選択された製品データと、
    // どの行から呼び出されたかの情報を渡して実行する
    if (typeof onProductSelectCallback === 'function') {
      onProductSelectCallback(product, activeRowElement);
    }
    modal.classList.add('hidden'); // モーダルを閉じる
  }
}

/**
 * APIを呼び出して製品を検索する
 */
async function performSearch() {
  const query = searchInput.value.trim();
  if (query.length < 2) {
    alert('2文字以上入力してください。');
    return;
  }
  searchResultsBody.innerHTML = '<tr><td colspan="6" class="center">検索中...</td></tr>';
  try {
    const res = await fetch(`/api/products/search?q=${encodeURIComponent(query)}`);
    if (!res.ok) {
        throw new Error(`サーバーエラー: ${res.status}`);
    }
    const products = await res.json();
    renderSearchResults(products);
  } catch (err) {
    searchResultsBody.innerHTML = `<tr><td colspan="6" class="center" style="color:red;">${err.message}</td></tr>`;
  }
}

/**
 * 検索結果をテーブルに描画する
 * @param {Array<object>} products
 */
function renderSearchResults(products) {
  if (!products || products.length === 0) {
    searchResultsBody.innerHTML = '<tr><td colspan="6" class="center">該当する製品が見つかりません。</td></tr>';
    return;
  }

  let html = '';
  products.forEach(p => {
    const productData = JSON.stringify(p);

    html += `
      <tr>
        <td class="left">${p.productName || ''}</td>
        <td class="left">${p.makerName || ''}</td>
        <td class="left">${p.formattedPackageSpec}</td>
        <td>${p.yjCode || ''}</td>
        <td>${p.productCode || ''}</td>
        <td><button class="select-product-btn" data-product='${productData.replace(/'/g, "&apos;")}'>選択</button></td>
      </tr>
    `;
  });
  searchResultsBody.innerHTML = html;
}

/**
 * モーダルを初期化し、イベントリスナーを設定する
 * @param {function} onSelect - 製品選択時に実行されるコールバック関数
 */
export function initModal(onSelect) {
  if (!modal || !closeModalBtn || !searchInput || !searchBtn || !searchResultsBody) {
    console.error("薬品検索モーダルの必須要素が見つかりません。");
    return;
  }
  onProductSelectCallback = onSelect;

  // イベントリスナーを一度だけ設定
  closeModalBtn.addEventListener('click', () => modal.classList.add('hidden'));
  searchBtn.addEventListener('click', performSearch);
  searchInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
      e.preventDefault(); // フォームの送信を防止
      performSearch();
    }
  });
  searchResultsBody.addEventListener('click', handleResultClick);
}

/**
 * モーダルを表示する
 * @param {HTMLElement} rowElement - モーダルを呼び出した明細の行要素
 */
export function showModal(rowElement) {
  if (modal) {
    activeRowElement = rowElement; // どの行から呼ばれたかを記憶
    modal.classList.remove('hidden');
    searchInput.value = ''; // 検索窓をクリア
    searchInput.focus(); // 検索窓にフォーカス
    searchResultsBody.innerHTML = '<tr><td colspan="6" class="center">製品名を入力して検索してください。</td></tr>';
  }
}