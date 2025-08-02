// File: static/js/update_master.js (新規作成)

/**
 * JCSHMSマスター更新ボタンの機能を初期化する
 */
export function initUpdateMaster() {
    const updateBtn = document.getElementById('updateJcshmsBtn');
    if (!updateBtn) return;

    updateBtn.addEventListener('click', async () => {
        if (!confirm('SOUフォルダのJCSHMS.CSVとJANCODE.CSVを新しいファイルに置き換えましたか？\n\nこの操作は時間がかかる場合があります。続行しますか？')) {
            return;
        }

        // ユーザーに応答を返すために、結果表示用のテーブルを準備
        const resultContainer = document.getElementById('upload-output-container');
        const inOutView = document.getElementById('in-out-view');
        const uploadView = document.getElementById('upload-view');
        
        resultContainer.innerHTML = `
            <div class="table-container">
                <h3>JCSHMSマスター更新処理中...</h3>
                <p>ブラウザを閉じないでください。</p>
            </div>`;
        inOutView.classList.add('hidden');
        uploadView.classList.remove('hidden');


        try {
            const res = await fetch('/api/master/update-jcshms', {
                method: 'POST',
            });
            const resData = await res.json();
            if (!res.ok) {
                throw new Error(resData.message || '更新処理に失敗しました。');
            }
            
            // 結果を表示
            let resultHTML = `
                <div class="table-container">
                    <h3>${resData.message}</h3>
                    <p>以下の製品マスターが更新されました。</p>
                    <table class="data-table">
                        <thead><tr><th>JANコード</th><th>製品名</th></tr></thead>
                        <tbody>`;
            
            resData.updatedProducts.forEach(p => {
                resultHTML += `<tr><td>${p.janCode}</td><td class="left">${p.productName}</td></tr>`;
            });

            resultHTML += `</tbody></table></div>`;
            resultContainer.innerHTML = resultHTML;

        } catch (err) {
            console.error(err);
            resultContainer.innerHTML = `<div class="table-container" style="color:red;">エラー: ${err.message}</div>`;
        }
    });
}