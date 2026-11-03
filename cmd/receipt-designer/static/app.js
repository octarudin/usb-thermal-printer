"use strict";

const $ = (selector) => document.querySelector(selector);

const blockList = $("#block-list");
const properties = $("#properties");
const receiptPreview = $("#receipt-preview");
const blockCount = $("#block-count");
const toast = $("#toast");

let selectedID = null;
let draggedID = null;
let toastTimer = null;

const state = {
  name: "Struk Taffware",
  blocks: sampleBlocks()
};

function uid() {
  if (globalThis.crypto?.randomUUID) return crypto.randomUUID();
  return `block-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function sampleBlocks() {
  return [
    { id: uid(), type: "text", text: "TOKO CONTOH", align: "center", bold: true, width: 2, height: 2 },
    { id: uid(), type: "text", text: "Jl. Contoh No. 123\nJakarta", align: "center", bold: false, width: 1, height: 1 },
    { id: uid(), type: "separator", character: "-" },
    {
      id: uid(), type: "items", items: [
        { name: "Kopi Susu", qty: "1 x Rp12.000", price: "Rp12.000" },
        { name: "Roti Bakar", qty: "2 x Rp8.000", price: "Rp16.000" }
      ]
    },
    { id: uid(), type: "separator", character: "-" },
    { id: uid(), type: "total", label: "TOTAL", value: "Rp28.000", align: "left" },
    { id: uid(), type: "text", text: "Terima kasih", align: "center", bold: false, width: 1, height: 1 },
    { id: uid(), type: "feed", lines: 3 }
  ];
}

function defaultBlock(type) {
  const common = { id: uid(), type };
  switch (type) {
    case "text": return { ...common, text: "Teks baru", align: "left", bold: false, width: 1, height: 1 };
    case "separator": return { ...common, character: "-" };
    case "items": return { ...common, items: [{ name: "Nama barang", qty: "1 x Rp0", price: "Rp0" }] };
    case "total": return { ...common, label: "TOTAL", value: "Rp0", align: "left" };
    case "image": return { ...common, image: "", align: "center" };
    case "qr": return { ...common, qrData: "https://example.com", qrSize: 5, qrLevel: 49, align: "center" };
    case "feed": return { ...common, lines: 3 };
    default: return common;
  }
}

const typeMeta = {
  text: ["Teks", "T"], separator: ["Pemisah", "—"], items: ["Daftar barang", "≡"],
  total: ["Total", "Σ"], image: ["Logo", "▧"], qr: ["QR Code", "▦"], feed: ["Feed kertas", "↧"]
};

function addBlock(type) {
  const block = defaultBlock(type);
  state.blocks.push(block);
  selectedID = block.id;
  renderAll();
  requestAnimationFrame(() => blockList.lastElementChild?.scrollIntoView({ behavior: "smooth", block: "nearest" }));
}

function selectedBlock() {
  return state.blocks.find((block) => block.id === selectedID) || null;
}

function blockSummary(block) {
  switch (block.type) {
    case "text": return block.text?.replaceAll("\n", " · ") || "Teks kosong";
    case "separator": return `Karakter “${block.character || "-"}”`;
    case "items": return `${block.items?.length || 0} barang`;
    case "total": return `${block.label || "TOTAL"} · ${block.value || "-"}`;
    case "image": return block.image ? "Gambar siap dicetak" : "Belum ada gambar";
    case "qr": return block.qrData || "Data QR kosong";
    case "feed": return `${block.lines || 3} baris`;
    default: return block.type;
  }
}

function renderStructure() {
  blockCount.textContent = `${state.blocks.length} blok`;
  if (!state.blocks.length) {
    blockList.innerHTML = '<div class="empty-list">Belum ada komponen.<br>Tambahkan blok dari panel sebelah kiri.</div>';
    return;
  }

  blockList.innerHTML = state.blocks.map((block) => {
    const meta = typeMeta[block.type] || [block.type, "?"];
    return `
      <article class="block-card ${block.id === selectedID ? "selected" : ""}" draggable="true" data-id="${escapeHTML(block.id)}">
        <span class="drag-handle" title="Seret untuk mengurutkan">⠿</span>
        <div><b>${escapeHTML(meta[0])}</b><small>${escapeHTML(blockSummary(block))}</small></div>
        <div class="card-actions">
          <button class="icon-button duplicate-block" title="Duplikasi" data-id="${escapeHTML(block.id)}">⧉</button>
          <button class="icon-button delete-block" title="Hapus" data-id="${escapeHTML(block.id)}">×</button>
        </div>
      </article>`;
  }).join("");
}

function renderProperties() {
  const block = selectedBlock();
  if (!block) {
    properties.className = "properties-empty";
    properties.innerHTML = '<div class="empty-symbol">↖</div><p>Pilih salah satu blok untuk mengubah isinya.</p>';
    return;
  }

  properties.className = "";
  switch (block.type) {
    case "text":
      properties.innerHTML = `
        ${textareaField("Isi teks", "text", block.text, "Baris baru akan dicetak sebagai LF.")}
        ${alignField(block.align)}
        <div class="field-row">
          ${selectField("Lebar", "width", block.width || 1, scaleOptions())}
          ${selectField("Tinggi", "height", block.height || 1, scaleOptions())}
        </div>
        <div class="field toggle-row"><span class="field-label">Tebal / bold</span><input class="toggle" type="checkbox" data-prop="bold" ${block.bold ? "checked" : ""}></div>`;
      break;
    case "separator":
      properties.innerHTML = `${inputField("Karakter pemisah", "character", block.character || "-", "Hanya karakter pertama yang digunakan.", 3)}`;
      break;
    case "items":
      properties.innerHTML = `
        <div class="field"><span class="field-label">Daftar barang</span><div class="item-editor">
          ${(block.items || []).map((item, index) => itemEditor(item, index)).join("")}
        </div><button class="small-action" id="add-item">+ Tambah barang</button>
        <p class="field-hint">Nama dan harga akan disusun otomatis dalam 32 kolom.</p></div>`;
      break;
    case "total":
      properties.innerHTML = `
        ${inputField("Label", "label", block.label || "TOTAL")}
        ${inputField("Nilai", "value", block.value || "Rp0")}
        ${alignField(block.align)}`;
      break;
    case "image":
      properties.innerHTML = `
        <div class="field"><label>File logo</label><input class="input" id="image-file" type="file" accept="image/png,image/jpeg,image/gif">
          <p class="field-hint">Maksimum 3 MB. Gambar diubah menjadi hitam-putih dan diperkecil maksimal 384 dot.</p>
          ${block.image ? `<img class="image-thumb" src="${escapeHTML(block.image)}" alt="Logo yang dipilih">` : ""}
        </div>${alignField(block.align)}`;
      break;
    case "qr":
      properties.innerHTML = `
        ${textareaField("Isi QR", "qrData", block.qrData || "", "Gunakan teks atau URL ASCII.")}
        ${alignField(block.align)}
        <div class="field-row">
          ${selectField("Ukuran modul", "qrSize", block.qrSize || 5, [2, 3, 4, 5, 6, 7, 8].map((n) => [n, `${n} dot`]))}
          ${selectField("Koreksi error", "qrLevel", block.qrLevel || 49, [[48, "L · 7%"], [49, "M · 15%"], [50, "Q · 25%"], [51, "H · 30%"]])}
        </div>`;
      break;
    case "feed":
      properties.innerHTML = `${inputField("Jumlah baris", "lines", block.lines || 3, "Berikan ruang untuk menyobek kertas.", 20, "number", 1)}`;
      break;
  }
}

function renderPreview() {
  if (!state.blocks.length) {
    receiptPreview.innerHTML = '<div class="empty-list">Preview akan muncul di sini.</div>';
    return;
  }

  receiptPreview.innerHTML = state.blocks.map((block) => previewBlock(block)).join("");
}

function previewBlock(block) {
  const align = ["left", "center", "right"].includes(block.align) ? block.align : "left";
  switch (block.type) {
    case "text": {
      const height = Number(block.height) || 1;
      const width = Number(block.width) || 1;
      const style = `text-align:${align};font-weight:${block.bold ? 900 : 400};font-size:${12 * height}px;letter-spacing:${Math.max(0, width - 1) * 1.8}px`;
      return `<div class="preview-block preview-text" style="${style}">${escapeHTML(block.text || "").replaceAll("\n", "<br>")}</div>`;
    }
    case "separator": {
      const char = (block.character || "-")[0];
      return `<div class="preview-block preview-separator">${escapeHTML(char.repeat(40))}</div>`;
    }
    case "items":
      return `<div class="preview-block">${(block.items || []).map((item) => `
        <div class="preview-item"><span>${escapeHTML(item.name || "")}</span><span>${escapeHTML(item.price || "")}</span></div>
        ${item.qty ? `<div class="preview-qty">${escapeHTML(item.qty)}</div>` : ""}`).join("")}</div>`;
    case "total":
      return `<div class="preview-block preview-total" style="text-align:${align}"><span>${escapeHTML(block.label || "TOTAL")}</span><span>${escapeHTML(block.value || "")}</span></div>`;
    case "image":
      return block.image ? `<div class="preview-block" style="text-align:${align}"><img class="preview-image" src="${escapeHTML(block.image)}" alt="Logo"></div>` : '<div class="preview-feed">Logo belum dipilih</div>';
    case "qr":
      return `<div class="preview-block" style="text-align:${align}"><div class="preview-qr"><span>QR</span></div></div>`;
    case "feed":
      return `<div class="preview-block preview-feed" style="height:${Math.max(24, (block.lines || 3) * 10)}px">feed ${block.lines || 3} baris</div>`;
    default: return "";
  }
}

function renderAll() {
  renderStructure();
  renderProperties();
  renderPreview();
}

function updateWithoutProperties() {
  renderStructure();
  renderPreview();
}

function alignField(value = "left") {
  return selectField("Perataan", "align", value, [["left", "Kiri"], ["center", "Tengah"], ["right", "Kanan"]]);
}

function scaleOptions() {
  return [1, 2, 3, 4].map((n) => [n, `${n}×`]);
}

function inputField(label, prop, value, hint = "", maxLength = 100, type = "text", min = "") {
  return `<div class="field"><label>${escapeHTML(label)}</label><input class="input" type="${type}" data-prop="${prop}" value="${escapeHTML(String(value ?? ""))}" maxlength="${maxLength}" ${min !== "" ? `min="${min}"` : ""}>${hint ? `<p class="field-hint">${escapeHTML(hint)}</p>` : ""}</div>`;
}

function textareaField(label, prop, value, hint = "") {
  return `<div class="field"><label>${escapeHTML(label)}</label><textarea class="textarea" data-prop="${prop}">${escapeHTML(value || "")}</textarea>${hint ? `<p class="field-hint">${escapeHTML(hint)}</p>` : ""}</div>`;
}

function selectField(label, prop, value, options) {
  return `<div class="field"><label>${escapeHTML(label)}</label><select class="select" data-prop="${prop}">${options.map(([optionValue, optionLabel]) => `<option value="${escapeHTML(String(optionValue))}" ${String(optionValue) === String(value) ? "selected" : ""}>${escapeHTML(String(optionLabel))}</option>`).join("")}</select></div>`;
}

function itemEditor(item, index) {
  return `<div class="item-row" data-item="${index}">
    <div class="item-row-top"><input class="input" data-item-prop="name" value="${escapeHTML(item.name || "")}" placeholder="Nama barang"><button class="icon-button remove-item" data-index="${index}" title="Hapus barang">×</button></div>
    <div class="item-row-bottom"><input class="input" data-item-prop="qty" value="${escapeHTML(item.qty || "")}" placeholder="1 x Rp0"><input class="input" data-item-prop="price" value="${escapeHTML(item.price || "")}" placeholder="Rp0"></div>
  </div>`;
}

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>'"]/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[char]);
}

function showToast(message, kind = "") {
  clearTimeout(toastTimer);
  toast.textContent = message;
  toast.className = `toast show ${kind}`;
  toastTimer = setTimeout(() => { toast.className = "toast"; }, 4200);
}

function cleanTemplate() {
  return JSON.parse(JSON.stringify(state));
}

async function printReceipt() {
  const button = $("#print-receipt");
  button.disabled = true;
  showToast("Mengirim data ke printer…");
  try {
    const response = await fetch("/api/print", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(cleanTemplate())
    });
    const result = await response.json();
    if (!response.ok) throw new Error(result.message || "Gagal mencetak struk.");
    showToast(`${result.message} ${result.bytes} byte melalui endpoint OUT ${result.endpoint}.`, "success");
  } catch (error) {
    showToast(error.message, "error");
  } finally {
    button.disabled = false;
  }
}

function saveTemplate() {
  const blob = new Blob([JSON.stringify(cleanTemplate(), null, 2)], { type: "application/json" });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = `${slug(state.name || "receipt-template")}.json`;
  link.click();
  URL.revokeObjectURL(link.href);
  showToast("Template JSON berhasil disimpan.", "success");
}

function slug(value) {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "") || "receipt-template";
}

async function loadTemplate(file) {
  try {
    const parsed = JSON.parse(await file.text());
    if (!parsed || !Array.isArray(parsed.blocks)) throw new Error("Berkas bukan template receipt yang valid.");
    state.name = String(parsed.name || "Template struk");
    state.blocks = parsed.blocks.map((block) => ({ ...block, id: block.id || uid() }));
    selectedID = state.blocks[0]?.id || null;
    renderAll();
    showToast("Template berhasil dibuka.", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

document.querySelectorAll("[data-add]").forEach((button) => {
  button.addEventListener("click", () => addBlock(button.dataset.add));
});

blockList.addEventListener("click", (event) => {
  const remove = event.target.closest(".delete-block");
  const duplicate = event.target.closest(".duplicate-block");
  const card = event.target.closest(".block-card");
  if (!card) return;

  if (remove) {
    event.stopPropagation();
    const index = state.blocks.findIndex((block) => block.id === remove.dataset.id);
    if (index >= 0) state.blocks.splice(index, 1);
    if (selectedID === remove.dataset.id) selectedID = state.blocks[Math.min(index, state.blocks.length - 1)]?.id || null;
    renderAll();
    return;
  }

  if (duplicate) {
    event.stopPropagation();
    const index = state.blocks.findIndex((block) => block.id === duplicate.dataset.id);
    if (index >= 0) {
      const copy = JSON.parse(JSON.stringify(state.blocks[index]));
      copy.id = uid();
      state.blocks.splice(index + 1, 0, copy);
      selectedID = copy.id;
      renderAll();
    }
    return;
  }

  selectedID = card.dataset.id;
  renderAll();
});

blockList.addEventListener("dragstart", (event) => {
  const card = event.target.closest(".block-card");
  if (!card) return;
  draggedID = card.dataset.id;
  card.classList.add("dragging");
  event.dataTransfer.effectAllowed = "move";
});

blockList.addEventListener("dragend", (event) => {
  event.target.closest(".block-card")?.classList.remove("dragging");
  draggedID = null;
});

blockList.addEventListener("dragover", (event) => {
  event.preventDefault();
  event.dataTransfer.dropEffect = "move";
});

blockList.addEventListener("drop", (event) => {
  event.preventDefault();
  const targetCard = event.target.closest(".block-card");
  if (!targetCard || !draggedID || targetCard.dataset.id === draggedID) return;
  const from = state.blocks.findIndex((block) => block.id === draggedID);
  const to = state.blocks.findIndex((block) => block.id === targetCard.dataset.id);
  if (from < 0 || to < 0) return;
  const [moved] = state.blocks.splice(from, 1);
  state.blocks.splice(to, 0, moved);
  renderAll();
});

properties.addEventListener("input", (event) => {
  const block = selectedBlock();
  if (!block) return;

  const prop = event.target.dataset.prop;
  if (prop) {
    if (event.target.type === "checkbox") block[prop] = event.target.checked;
    else if (["width", "height", "qrSize", "qrLevel", "lines"].includes(prop)) block[prop] = Number(event.target.value);
    else block[prop] = event.target.value;
    updateWithoutProperties();
    return;
  }

  const itemProp = event.target.dataset.itemProp;
  if (itemProp) {
    const itemRow = event.target.closest("[data-item]");
    const item = block.items?.[Number(itemRow.dataset.item)];
    if (item) item[itemProp] = event.target.value;
    updateWithoutProperties();
  }
});

properties.addEventListener("change", (event) => {
  if (event.target.id !== "image-file") return;
  const file = event.target.files?.[0];
  if (!file) return;
  if (file.size > 3 * 1024 * 1024) {
    showToast("Ukuran gambar maksimum 3 MB.", "error");
    return;
  }
  const reader = new FileReader();
  reader.onload = () => {
    const block = selectedBlock();
    if (block?.type === "image") {
      block.image = reader.result;
      renderAll();
      showToast("Logo berhasil dimuat.", "success");
    }
  };
  reader.onerror = () => showToast("Gambar tidak dapat dibaca.", "error");
  reader.readAsDataURL(file);
});

properties.addEventListener("click", (event) => {
  const block = selectedBlock();
  if (!block) return;
  if (event.target.id === "add-item") {
    block.items ||= [];
    block.items.push({ name: "Barang baru", qty: "1 x Rp0", price: "Rp0" });
    renderAll();
  }
  const remove = event.target.closest(".remove-item");
  if (remove && block.type === "items") {
    block.items.splice(Number(remove.dataset.index), 1);
    renderAll();
  }
});

$("#print-receipt").addEventListener("click", printReceipt);
$("#save-template").addEventListener("click", saveTemplate);
$("#load-template").addEventListener("click", () => $("#template-file").click());
$("#template-file").addEventListener("change", (event) => {
  const file = event.target.files?.[0];
  if (file) loadTemplate(file);
  event.target.value = "";
});
$("#new-template").addEventListener("click", () => {
  if (!confirm("Ganti desain saat ini dengan template contoh baru?")) return;
  state.name = "Struk Taffware";
  state.blocks = sampleBlocks();
  selectedID = state.blocks[0]?.id || null;
  renderAll();
});

selectedID = state.blocks[0]?.id || null;
renderAll();

fetch("/api/info")
  .then((response) => response.json())
  .then((info) => showToast(`Siap · printer USB ${info.vendorID}:${info.productID}`))
  .catch(() => showToast("Backend lokal tidak dapat dihubungi.", "error"));
