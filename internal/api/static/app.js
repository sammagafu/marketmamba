const API = '/api/v1';

function getKey() {
  return localStorage.getItem('mm_api_key') || '';
}

function headers() {
  const h = { 'Content-Type': 'application/json' };
  const key = getKey();
  if (key) h['X-API-Key'] = key;
  return h;
}

async function api(path, opts = {}) {
  const res = await fetch(API + path, { ...opts, headers: { ...headers(), ...opts.headers } });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

document.getElementById('saveKey').onclick = () => {
  localStorage.setItem('mm_api_key', document.getElementById('apiKey').value.trim());
  refresh();
};

document.getElementById('apiKey').value = getKey();

let brokerTypes = [];

function renderFields(providerId) {
  const box = document.getElementById('dynamicFields');
  box.innerHTML = '';
  const t = brokerTypes.find((b) => b.id === providerId);
  if (!t || !t.fields.length) return;
  t.fields.forEach((f) => {
    const div = document.createElement('div');
    div.className = 'row';
    const id = 'cred_' + f.key;
    div.innerHTML = `
      <label for="${id}">${f.label}${f.required ? ' *' : ''}</label>
      <input id="${id}" name="${f.key}" type="${f.type === 'password' ? 'password' : 'text'}"
        placeholder="${f.placeholder || ''}" ${f.required ? 'required' : ''} />
    `;
    box.appendChild(div);
  });
}

function collectCredentials() {
  const creds = {};
  document.querySelectorAll('#dynamicFields input').forEach((el) => {
    if (el.name && el.value) creds[el.name] = el.value;
  });
  return creds;
}

function showMsg(text, ok) {
  const el = document.getElementById('formMsg');
  el.textContent = text;
  el.className = 'msg ' + (ok ? 'ok' : 'err');
}

async function loadBrokerTypes() {
  const data = await api('/brokers/types');
  brokerTypes = data.brokers || [];
  const sel = document.getElementById('provider');
  sel.innerHTML = brokerTypes
    .map(
      (b) =>
        `<option value="${b.id}">${b.name} (${b.status === 'live' ? 'available' : 'coming soon'})</option>`
    )
    .join('');
  sel.onchange = () => renderFields(sel.value);
  renderFields(sel.value);
}

async function refresh() {
  try {
    const status = await api('/status');
    document.getElementById('status').innerHTML = `
      <p><strong>${status.app}</strong> · ${status.env}</p>
      <p>Broker: <code>${status.provider}</code></p>
      <p>Auto trading: ${status.auto_trading ? 'on' : 'off'} · Paused: ${status.is_paused ? 'yes' : 'no'}</p>
    `;

    const acct = await api('/account');
    document.getElementById('account').innerHTML = `
      <p>Balance: <strong>$${Number(acct.balance).toLocaleString(undefined, { minimumFractionDigits: 2 })}</strong></p>
      <p>Equity: $${Number(acct.equity).toLocaleString(undefined, { minimumFractionDigits: 2 })}</p>
    `;

    const pos = await api('/positions');
    const list = pos.positions || [];
    if (!list.length) {
      document.getElementById('positions').textContent = 'No open positions';
      document.getElementById('positions').className = 'muted';
    } else {
      document.getElementById('positions').className = '';
      document.getElementById('positions').innerHTML = `
        <table>
          <thead><tr><th>Symbol</th><th>Type</th><th>Qty</th><th>P/L</th></tr></thead>
          <tbody>
            ${list.map((p) => `<tr><td>${p.Symbol}</td><td>${p.Type}</td><td>${p.Quantity}</td><td>${(p.Profit ?? 0).toFixed(2)}</td></tr>`).join('')}
          </tbody>
        </table>`;
    }

    const conn = await api('/brokers/connection');
    if (conn.connection) {
      document.getElementById('provider').value = conn.connection.provider;
      renderFields(conn.connection.provider);
    }
  } catch (e) {
    document.getElementById('status').innerHTML = `<span class="msg err">${e.message}</span>`;
  }
}

document.getElementById('testBtn').onclick = async () => {
  try {
    const body = {
      provider: document.getElementById('provider').value,
      label: document.getElementById('label').value,
      credentials: collectCredentials(),
    };
    const r = await api('/brokers/test', { method: 'POST', body: JSON.stringify(body) });
    showMsg(`Connection OK · balance $${r.balance}`, true);
  } catch (e) {
    showMsg(e.message, false);
  }
};

document.getElementById('brokerForm').onsubmit = async (e) => {
  e.preventDefault();
  try {
    const body = {
      provider: document.getElementById('provider').value,
      label: document.getElementById('label').value,
      credentials: collectCredentials(),
    };
    await api('/brokers/connection', { method: 'POST', body: JSON.stringify(body) });
    showMsg('Broker saved and activated.', true);
    refresh();
  } catch (e) {
    showMsg(e.message, false);
  }
};

loadBrokerTypes().then(refresh).catch((e) => {
  document.getElementById('status').innerHTML = `<span class="msg err">${e.message}</span>`;
});
