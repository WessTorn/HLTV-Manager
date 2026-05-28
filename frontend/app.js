const rows = document.getElementById("hltvRows");
const statusEl = document.getElementById("status");
const refreshBtn = document.getElementById("refreshBtn");
const startAllBtn = document.getElementById("startAllBtn");
const stopAllBtn = document.getElementById("stopAllBtn");

const apiBase = "/api/v1";

function setStatus(text, isError = false) {
  statusEl.textContent = text;
  statusEl.style.color = isError ? "#b22f2f" : "#61707f";
}

async function request(path, options = {}) {
  const res = await fetch(`${apiBase}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });

  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return body;
}

function renderRow(item) {
  const tr = document.createElement("tr");
  const runningLabel = item.running
    ? '<span class="pill pill-ok">RUNNING</span>'
    : '<span class="pill pill-stop">STOPPED</span>';

  tr.innerHTML = `
    <td>${item.id}</td>
    <td>${item.name}</td>
    <td>${item.connect}</td>
    <td>${runningLabel}</td>
    <td>${item.demos_count}</td>
    <td>
      <div class="row-actions">
        <button class="btn" data-action="start">Start</button>
        <button class="btn btn-danger" data-action="stop">Stop</button>
        <button class="btn btn-ghost" data-action="restart">Restart</button>
      </div>
    </td>
  `;

  tr.querySelector('[data-action="start"]').addEventListener("click", () => runAction(item.id, "start"));
  tr.querySelector('[data-action="stop"]').addEventListener("click", () => runAction(item.id, "stop"));
  tr.querySelector('[data-action="restart"]').addEventListener("click", () => runAction(item.id, "restart"));

  return tr;
}

async function loadHLTV() {
  setStatus("Loading...");
  rows.innerHTML = "";
  try {
    const data = await request("/hltv");
    for (const item of data.items || []) {
      rows.appendChild(renderRow(item));
    }
    setStatus(`Loaded: ${(data.items || []).length} HLTV`);
  } catch (error) {
    setStatus(`Load error: ${error.message}`, true);
  }
}

async function runAction(id, action) {
  try {
    setStatus(`${action.toUpperCase()} #${id}...`);
    await request(`/hltv/${id}/${action}`, { method: "POST" });
    await loadHLTV();
  } catch (error) {
    setStatus(`${action} error: ${error.message}`, true);
  }
}

async function runGlobalAction(action) {
  try {
    setStatus(`${action.toUpperCase()} ALL...`);
    await request(`/hltv/${action}-all`, { method: "POST" });
    await loadHLTV();
  } catch (error) {
    setStatus(`${action}-all error: ${error.message}`, true);
  }
}

refreshBtn.addEventListener("click", loadHLTV);
startAllBtn.addEventListener("click", () => runGlobalAction("start"));
stopAllBtn.addEventListener("click", () => runGlobalAction("stop"));

loadHLTV();
