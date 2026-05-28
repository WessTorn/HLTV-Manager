const rows = document.getElementById("hltvRows");
const demoRows = document.getElementById("demoRows");
const demosTitle = document.getElementById("demosTitle");
const statusEl = document.getElementById("status");
const healthStatus = document.getElementById("healthStatus");
const refreshBtn = document.getElementById("refreshBtn");
const startAllBtn = document.getElementById("startAllBtn");
const stopAllBtn = document.getElementById("stopAllBtn");

const apiBase = "/api/v1";
let selectedHLTV = null;

function setStatus(text, isError = false) {
  statusEl.textContent = text;
  statusEl.style.color = isError ? "#b22f2f" : "#61707f";
}

function setHealthState(up) {
  if (up) {
    healthStatus.textContent = "API: UP";
    healthStatus.className = "health-badge health-up";
    return;
  }

  healthStatus.textContent = "API: DOWN";
  healthStatus.className = "health-badge health-down";
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
        <button class="btn btn-ghost" data-action="demos">Demos</button>
      </div>
    </td>
  `;

  tr.querySelector('[data-action="start"]').addEventListener("click", () => runAction(item.id, "start"));
  tr.querySelector('[data-action="stop"]').addEventListener("click", () => runAction(item.id, "stop"));
  tr.querySelector('[data-action="restart"]').addEventListener("click", () => runAction(item.id, "restart"));
  tr.querySelector('[data-action="demos"]').addEventListener("click", () => loadDemos(item.id, item.name));

  return tr;
}

function renderDemoRow(hltvID, demo) {
  const tr = document.createElement("tr");
  const archived = demo.archived
    ? '<span class="pill pill-stop">YES</span>'
    : '<span class="pill pill-ok">NO</span>';
  const downloadLink = `${apiBase}/hltv/${hltvID}/demos/${demo.id}/download`;

  tr.innerHTML = `
    <td>${demo.id}</td>
    <td>${demo.map || "-"}</td>
    <td>${demo.date || "-"}</td>
    <td>${demo.time || "-"}</td>
    <td>${archived}</td>
    <td><a class="btn btn-small" href="${downloadLink}">Download</a></td>
  `;
  return tr;
}

async function checkHealth() {
  try {
    await request("/health");
    setHealthState(true);
  } catch (error) {
    setHealthState(false);
  }
}

async function loadDemos(id, name, silent = false) {
  selectedHLTV = { id, name };
  demosTitle.textContent = `Demos: ${name} (#${id})`;
  demoRows.innerHTML = "";

  if (!silent) {
    setStatus(`Loading demos for #${id}...`);
  }

  try {
    const data = await request(`/hltv/${id}/demos`);
    const items = data.items || [];

    if (!items.length) {
      const tr = document.createElement("tr");
      tr.innerHTML = '<td colspan="6" class="muted-cell">No demos found.</td>';
      demoRows.appendChild(tr);
    } else {
      for (const demo of items) {
        demoRows.appendChild(renderDemoRow(id, demo));
      }
    }

    if (!silent) {
      setStatus(`Loaded demos: ${items.length} for #${id}`);
    }
  } catch (error) {
    setStatus(`Demos error: ${error.message}`, true);
  }
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

    if (selectedHLTV) {
      await loadDemos(selectedHLTV.id, selectedHLTV.name, true);
    }
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

async function refreshAll() {
  await checkHealth();
  await loadHLTV();
}

refreshBtn.addEventListener("click", refreshAll);
startAllBtn.addEventListener("click", () => runGlobalAction("start"));
stopAllBtn.addEventListener("click", () => runGlobalAction("stop"));

refreshAll();
