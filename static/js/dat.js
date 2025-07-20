// File: static/js/dat.js

document.addEventListener("DOMContentLoaded", () => {
  const btn     = document.getElementById("datBtn");
  const input   = document.getElementById("datInput");
  const debug   = document.getElementById("debug");
  const table   = document.getElementById("outputTable");
  const thead   = table.querySelector("thead");
  const tbody   = table.querySelector("tbody");

  // 「納品・返品」クリックでクリア＆ファイル選択
  btn.addEventListener("click", () => {
    thead.innerHTML = "";
    tbody.innerHTML = "";
    debug.textContent = "";
    input.value = null;
    input.click();
  });

  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    debug.textContent = "DATファイルアップロード中…";

    // form にファイル詰めて POST
    const form = new FormData();
    for (const file of input.files) {
      form.append("file", file);
    }

    try {
      const res = await fetch("/uploadDat", { method: "POST", body: form });
      if (!res.ok) {
        debug.textContent = `アップロード失敗: ${res.status}`;
        return;
      }

      const data    = await res.json();
      const records = Array.isArray(data.records) ? data.records : [];

      // カウンタ表示
      debug.textContent =
        `Parsed: ${data.parsed}, Duplicates: ${data.duplicates}, ` +
        `MA: ${data.maCount}, DA: ${data.daCount}`;

      // ヘッダー生成
      thead.innerHTML = `
        <tr>
          <th>日付</th><th>種別</th><th>YJ</th><th>JAN</th><th>製品名</th>
          <th>包装</th><th>メーカー</th><th class="num">個数</th>
          <th class="num">JAN数量</th><th>JAN単位</th>
          <th class="num">YJ数量</th><th>YJ単位</th><th class="num">単価</th>
          <th class="num">金額</th><th class="num">税額</th><th class="num">税率</th>
          <th>期限</th><th>ロット</th><th>得意先</th>
          <th>伝票番号</th><th class="num">行</th><th>MA</th>
        </tr>`;
      tbody.innerHTML = "";

      // デバッグ: Ama の中身と型を一度ログに出す
      console.log(
        "Ama values:",
        records.map(r => ({ Ama: r.Ama, type: typeof r.Ama }))
      );

      // フィルタ：文字列化＋trim して「1〜6」を含むものだけ表示
      records
        .filter(r => {
          const a = String(r.Ama).trim();
          return ["1","2","3","4","5","6"].includes(a);
        })
        .forEach(rec => {
          const tr = document.createElement("tr");
          tr.classList.add("modified");
          tr.innerHTML = `
            <td>${rec.Adate            || ""}</td>
            <td>${rec.Aflag            || ""}</td>
            <td>${rec.Ayj              || ""}</td>
            <td>${rec.Ajc              || ""}</td>
            <td>${rec.Apname           || ""}</td>
            <td>${rec.Apkg             || ""}</td>
            <td>${rec.Amaker           || ""}</td>
            <td class="num">${rec.Adatqty    || ""}</td>
            <td class="num">${rec.Ajanqty    || ""}</td>
            <td>${rec.Ajanunitname     || ""}</td>
            <td class="num">${rec.Ayjqty     || ""}</td>
            <td>${rec.Ayjunitname      || ""}</td>
            <td class="num">${rec.Aunitprice || ""}</td>
            <td class="num">${rec.Asubtotal  || ""}</td>
            <td class="num">${rec.Ataxamount || ""}</td>
            <td class="num">${rec.Ataxrate   || ""}</td>
            <td>${rec.Aexpdate         || ""}</td>
            <td>${rec.Alot             || ""}</td>
            <td>${rec.Apcode           || ""}</td>
            <td>${rec.Arpnum           || ""}</td>
            <td class="num">${rec.Alnum      || ""}</td>
            <td>${String(rec.Ama).trim()  || ""}</td>
          `;
          tbody.appendChild(tr);
        });

    } catch (err) {
      console.error(err);
      debug.textContent = "DATアップロードエラー: " + err.message;
    }
  });
});