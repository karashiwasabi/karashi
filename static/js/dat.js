document.addEventListener("DOMContentLoaded", () => {
  const btn     = document.getElementById("datBtn");
  const input   = document.getElementById("datInput");
  const table   = document.getElementById("outputTable");
  const thead   = table.querySelector("thead");
  const tbody   = table.querySelector("tbody");
  const indi    = document.getElementById("indicator");

  btn.addEventListener("click", () => {
    const filter = document.getElementById("aggregateFilter");
    if (filter) filter.style.display = "none";

    thead.innerHTML = "";
    tbody.innerHTML = "";
    indi.textContent = "";
    input.value = null;
    input.click();
  });

  input.addEventListener("change", async () => {
    if (!input.files.length) return;
    indi.textContent = "DATファイルアップロード中…";

    for (let file of input.files) {
      const form = new FormData();
      form.append("datFileInput[]", file);

      try {
        const res  = await fetch("/uploadDat", { method: "POST", body: form });
        const data = await res.json();

        indi.textContent = `${file.name}：読み込み ${data.count} 件`;
        thead.innerHTML = `
          <tr>
            <th>日付</th><th>JAN</th><th>YJ</th><th>商品名</th><th>包装</th>
            <th>数量</th><th>JAN数</th><th>JAN単位</th><th>JAN単位CD</th>
            <th>YJ数</th><th>YJ単位</th><th>単価</th><th>小計</th>
            <th>税額</th><th>税率</th><th>期限</th><th>ロット</th>
            <th>伝票</th><th>行</th><th>区分</th><th>得意先</th>
          </tr>`;

        tbody.innerHTML = "";
        data.records.forEach(rec => {
          const tr = document.createElement("tr");
          tr.innerHTML = `
            <td>${rec.slipdate}</td><td>${rec.jancode}</td><td>${rec.yjcode}</td><td>${rec.productname}</td><td>${rec.packaging}</td>
            <td>${rec.datqty}</td><td>${rec.janquantity}</td><td>${rec.janunitname}</td><td>${rec.janunitcode}</td>
            <td>${rec.yjquantity}</td><td>${rec.yjunitname}</td><td>${rec.unitprice}</td>
            <td>${rec.subtotalamount}</td><td>${rec.taxamount}</td><td>${rec.taxrate}</td>
            <td>${rec.expirydate}</td><td>${rec.lotnumber}</td><td>${rec.receiptnumber}</td>
            <td>${rec.linenumber}</td><td>${rec.flag}</td><td>${rec.partnercode}</td>
          `;
          tbody.appendChild(tr);
        });

      } catch (err) {
        console.error(err);
        indi.textContent = "DATアップロードエラー: " + err.message;
      }
    }
  });
});