// File: static/js/usage.js

document.addEventListener("DOMContentLoaded", () => {
  const btn = document.getElementById("usageBtn");
  const input = document.getElementById("usageInput");
  const debug = document.getElementById("debug");
  const table = document.getElementById("outputTable");
  const thead = table.querySelector("thead");
  const tbody = table.querySelector("tbody");

  // USAGE ボタン押下でテーブル初期化＆ファイル選択ダイアログ
  btn.addEventListener("click", () => {
    debug.textContent = "";
    // vvv ヘッダーに「JAN員数」「YJ員数」を追加 vvv
    thead.innerHTML = `
  <tr>
    <th>日付</th><th>種別</th><th>YJ</th><th>JAN</th><th>製品名</th>
    <th>包装</th><th>メーカー</th><th class="num">個数</th>
    <th class="num">JAN数量</th><th class="num">JAN包装数量</th><th>JAN単位</th>
    <th class="num">YJ数量</th><th class="num">YJ包装数量</th><th>YJ単位</th>
    <th class="num">単価</th><th class="num">金額</th><th class="num">税額</th>
    <th class="num">税率</th><th>期限</th><th>ロット</th><th>得意先</th>
    <th>伝票番号</th><th class="num">行</th><th>MA</th>
  </tr>`;
    // ^^^ ここまで ^^^
    tbody.innerHTML = "";
    input.value = null;
    input.click();
  });

  // ファイル選択 → サーバへアップロード → JSONで受信 → テーブル描画
  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    debug.textContent = "USAGEファイルアップロード中…";

    const form = new FormData();
    for (const file of input.files) {
      form.append("file", file);
    }

    try {
      const res = await fetch("/uploadUsage", {
        method: "POST",
        body: form
      });
      if (!res.ok) {
        debug.textContent = `アップロード失敗: ${res.status}`;
        return;
      }

      const data = await res.json();
      const records = Array.isArray(data.records) ? data.records : [];
      
      debug.textContent = `受信件数: ${records.length}`;

      tbody.innerHTML = "";
      records
        .filter(r => ["1","2","3","4","5","6"].includes(String(r.Ama).trim()))
        .forEach(rec => {
          const tr = document.createElement("tr");
          tr.classList.add("modified");
          // vvv データ行に rec.Ajpu と rec.Ayjpu を追加 vvv
          tr.innerHTML = `
            <td>${rec.Adate || ""}</td>
            <td>${rec.Aflag || ""}</td>
            <td>${rec.Ayj   || ""}</td>
            <td>${rec.Ajc   || ""}</td>
            <td>${rec.Apname|| ""}</td>
            <td>${rec.Apkg  || ""}</td>
            <td>${rec.Amaker|| ""}</td>
            <td class="num">${rec.Adatqty      || ""}</td>
            <td class="num">${rec.Ajanqty      || ""}</td>
            <td class="num">${rec.Ajpu         || ""}</td>
            <td>${rec.Ajanunitnm   || ""}</td>
            <td class="num">${rec.Ayjqty       || ""}</td>
            <td class="num">${rec.Ayjpu        || ""}</td>
            <td>${rec.Ayjunitnm    || ""}</td>
            <td class="num">${rec.Aunitprice   || ""}</td>
            <td class="num">${rec.Asubtotal    || ""}</td>
            <td class="num">${rec.Ataxamount   || ""}</td>
            <td class="num">${rec.Ataxrate     || ""}</td>
            <td>${rec.Aexpdate        || ""}</td>
            <td>${rec.Alot            || ""}</td>
            <td>${rec.Apcode          || ""}</td>
            <td>${rec.Arpnum          || ""}</td>
            <td class="num">${rec.Alnum      || ""}</td>
            <td>${String(rec.Ama).trim()       || ""}</td>
          `;
          // ^^^ ここまで ^^^
          tbody.appendChild(tr);
        });
    } catch (err) {
      console.error(err);
      debug.textContent = "アップロードエラー: " + err.message;
    }
  });
});