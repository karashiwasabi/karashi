// File: static/js/reprocess.js

export function initReprocessButton() {
    const reprocessBtn = document.getElementById('reprocessBtn');
    if (!reprocessBtn) return;

    reprocessBtn.addEventListener('click', async () => {
        if (!confirm('仮登録状態の取引データを、最新のマスター情報で更新します。よろしいですか？')) {
            return;
        }

        try {
            const res = await fetch('/api/transactions/reprocess', {
                method: 'POST',
            });
            const data = await res.json();
            if (!res.ok) {
                throw new Error(data.message || '処理に失敗しました。');
            }
            alert(data.message);
        } catch (err) {
            console.error(err);
            alert(`エラー: ${err.message}`);
        }
    });
}