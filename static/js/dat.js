// File: static/js/dat.js
document.addEventListener("DOMContentLoaded", () => {
  const btn = document.getElementById("datBtn");
  const input = document.getElementById("datInput");
  const debug = document.getElementById("debug");
  const table = document.getElementById("outputTable");
  const thead = table.querySelector("thead");
  const tbody = table.querySelector("tbody");

  btn.addEventListener("click", () => {
    input.value = null;
    input.click();
  });

  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    debug.textContent = `${input.files.length}件のDATファイルをアップロード中…`;

    let allRecords = [];
    for (const file of input.files) {
      const form = new FormData();
      form.append("file", file);

      try {
        const res = await fetch("/uploadDat", { method: "POST", body: form });
        if (!res.ok) {
          console.error(`Error uploading ${file.name}: ${res.status}`);
          continue;
        }

        const data = await res.json();
        if (Array.isArray(data.records)) {
          allRecords.push(...data.records);
        }
      } catch (err) {
        console.error(`Error processing ${file.name}:`, err);
        continue;
      }
    }

    debug.textContent = `合計処理件数: ${allRecords.length}`;
    
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
    tbody.innerHTML = "";

    allRecords.forEach(rec => {
      const tr = document.createElement("tr");
      tr.classList.add("modified");
      tr.innerHTML = `
        <td>${rec.Adate || ""}</td>
        <td>${rec.Aflag || ""}</td>
        <td>${rec.Ayj || ""}</td>
        <td>${rec.Ajc || ""}</td>
        <td>${rec.Apname || ""}</td>
        <td>${rec.Apkg || ""}</td>
        <td>${rec.Amaker || ""}</td>
        <td class="num">${rec.Adatqty || ""}</td>
        <td class="num">${rec.Ajanqty || ""}</td>
        <td class="num">${rec.Ajpu || ""}</td>
        <td>${rec.Ajanunitnm || ""}</td>
        <td class="num">${rec.Ayjqty || ""}</td>
        <td class="num">${rec.Ayjpu || ""}</td>
        <td>${rec.Ayjunitnm || ""}</td>
        <td class="num">${rec.Aunitprice || ""}</td>
        <td class="num">${rec.Asubtotal || ""}</td>
        <td class="num">${rec.Ataxamount || ""}</td>
        <td class="num">${rec.Ataxrate || ""}</td>
        <td>${rec.Aexpdate || ""}</td>
        <td>${rec.Alot || ""}</td>
        <td>${rec.Apcode || ""}</td>
        <td>${rec.Arpnum || ""}</td>
        <td class="num">${rec.Alnum || ""}</td>
        <td>${String(rec.Ama).trim() || ""}</td>
      `;
      tbody.appendChild(tr);
    });
  });
});