// File: static/js/inout_modal.js

// --- プライベート変数 ---
let onProductSelectCallback = null;

// --- DOM要素の取得 ---
const modal = document.getElementById('search-modal'); 
const closeModalBtn = document.getElementById('closeModalBtn'); 
const searchInput = document.getElementById('product-search-input');
const searchBtn = document.getElementById('product-search-btn');
const searchResultsBody = document.querySelector('#search-results-table tbody'); 

/**
 * 検索結果テーブルでのクリックを処理（イベント委譲）
 * @param {Event} event
 */
function handleResultClick(event) {
  if (event.target && event.target.classList.contains('select-product-btn')) {
    const product = JSON.parse(event.target.dataset.product); 
    if (typeof onProductSelectCallback === 'function') { 
      onProductSelectCallback(product); 
    }
    modal.classList.add('hidden'); 
  }
}

/**
 * APIを叩いて製品を検索する
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
    if (!res.ok) throw new Error(`サーバーエラー: ${res.status}`); 
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
    // ★★★ Goから送られてきた整形済み包装表記をそのまま使う ★★★
    html += `
      <tr>
        <td class="left">${p.productName || ''}</td>
        <td class="left">${p.makerName || ''}</td>
        <td class="left">${p.formattedPackageSpec}</td>
        <td>${p.yjCode || ''}</td>
        <td>${p.productCode || ''}</td>
        <td><button class="select-product-btn" data-product='${productData}'>選択</button></td>
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

  closeModalBtn.addEventListener('click', () => modal.classList.add('hidden'));
  searchBtn.addEventListener('click', performSearch); 
  searchInput.addEventListener('keypress', (e) => { 
    if (e.key === 'Enter') performSearch();
  });
  searchResultsBody.addEventListener('click', handleResultClick); 
}

/**
 * モーダルを表示する
 */
export function showModal() {
  if (modal) {
    modal.classList.remove('hidden');
    searchInput.focus();
    searchResultsBody.innerHTML = '<tr><td colspan="6">製品名を入力して検索してください。</td></tr>'; 
  }
}