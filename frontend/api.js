const API_BASE = "/api/v1";
const DRONES_POLL_MS = 500;
const ORDERS_POLL_MS = 1500;

document.getElementById("dronesPollMs").textContent = String(DRONES_POLL_MS);
document.getElementById("ordersPollMs").textContent = String(ORDERS_POLL_MS);

const elApiStatus = document.getElementById("apiStatus");
const elOrdersList = document.getElementById("ordersList");
const elOrderCreateResult = document.getElementById("orderCreateResult");

const elItems = document.getElementById("items");
const elLat = document.getElementById("lat");
const elLon = document.getElementById("lon");
const btnPickOnMap = document.getElementById("btnPickOnMap");

const LS_KEY = "hive.orders.v1";

let stores = [];
let bases = [];

let dronesLayer;
let storesLayer;
let basesLayer;
let ordersLayer;

let droneMarkers = new Map();
let orderMarkers = new Map();

let pickMode = false;
let pickMarker = null;

let dronesUpdateInFlight = false;
let lastDronesOkAt = 0;

function setApiPill(state, text) {
    elApiStatus.classList.remove("pill--ok", "pill--bad", "pill--gray");
    elApiStatus.classList.add(state === "ok" ? "pill--ok" : state === "bad" ? "pill--bad" : "pill--gray");
    elApiStatus.querySelector(".text").textContent = text;
}

function escapeHtml(s) {
    return String(s ?? "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function fmtNum(n, digits = 6) {
    if (n === null || n === undefined || Number.isNaN(Number(n))) return "—";
    return Number(n).toFixed(digits);
}

function fmtBattery(n) {
    if (n === null || n === undefined || Number.isNaN(Number(n))) return "—";
    return `${Number(n).toFixed(1)}%`;
}

function fmtSpeed(n) {
    if (n === null || n === undefined || Number.isNaN(Number(n))) return "—";
    return `${Number(n).toFixed(2)} m/s`;
}

function ageFrom(ms) {
    if (!ms) return "—";
    const a = Date.now() - Number(ms);
    if (Number.isNaN(a)) return "—";
    if (a < 1000) return `${Math.round(a)}ms ago`;
    if (a < 60_000) return `${(a / 1000).toFixed(1)}s ago`;
    if (a < 3_600_000) return `${Math.round(a / 60_000)}m ago`;
    return `${Math.round(a / 3_600_000)}h ago`;
}

function normalizeEnum(v) {
    const s = String(v || "").trim().toUpperCase();
    if (!s) return "UNKNOWN";
    return (
        s
            .replace(/^.*STATUS_/, "")
            .replace(/^.*TARGET_/, "")
            .replace(/^.*ACTION_/, "")
            .replace(/^.*EVENT_/, "")
            .replace(/[^A-Z0-9_]/g, "_")
            .replace(/_+/g, "_")
            .replace(/^_+|_+$/g, "") || "UNKNOWN"
    );
}

function humanizeEnum(v) {
    const s = normalizeEnum(v);
    return s
        .split("_")
        .filter(Boolean)
        .map((w) => w[0] + w.slice(1).toLowerCase())
        .join(" ");
}

const DRONE_STATUS_LABEL = {
    FREE: "Available",
    BUSY: "On mission",
    CHARGING: "Charging",
    OFFLINE: "Offline",
    UNKNOWN: "Unknown",
};

const ORDER_STATUS_LABEL = {
    CREATED: "Created",
    PENDING: "Pending",
    ASSIGNED: "Assigned",
    DELIVERING: "Delivering",
    COMPLETED: "Completed",
    FAILED: "Failed",
    CANCELED: "Canceled",
    CANCELLED: "Canceled",
};

const ASSIGNMENT_STATUS_LABEL = {
    CREATED: "Created",
    ASSIGNED: "Assigned",
    FLYING_TO_STORE: "Heading to store",
    AT_STORE: "At store",
    PICKED_UP_CARGO: "Picked up cargo",
    FLYING_TO_CLIENT: "Heading to customer",
    AT_CLIENT: "At customer",
    DROPPED_CARGO: "Dropped cargo",
    RETURNING_BASE: "Returning to base",
    COMPLETED: "Completed",
    FAILED: "Failed",
};

const TARGET_KIND_LABEL = {
    STORE: "Store",
    CLIENT: "Customer",
    BASE: "Base",
    NONE: "None",
    UNKNOWN: "Unknown",
};

function droneStatusLabel(v) {
    const k = normalizeEnum(v);
    return DRONE_STATUS_LABEL[k] || humanizeEnum(k);
}

function orderStatusLabel(v) {
    const k = normalizeEnum(v);
    return ORDER_STATUS_LABEL[k] || humanizeEnum(k);
}

function assignmentStatusLabel(v) {
    const k = normalizeEnum(v);
    return ASSIGNMENT_STATUS_LABEL[k] || humanizeEnum(k);
}

function targetKindLabel(v) {
    const k = normalizeEnum(v);
    return TARGET_KIND_LABEL[k] || humanizeEnum(k);
}

function distMeters(aLat, aLon, bLat, bLon) {
    const R = 6371000;
    const toRad = (x) => (x * Math.PI) / 180;
    const dLat = toRad(bLat - aLat);
    const dLon = toRad(bLon - aLon);
    const lat1 = toRad(aLat);
    const lat2 = toRad(bLat);
    const sinDLat = Math.sin(dLat / 2);
    const sinDLon = Math.sin(dLon / 2);
    const h = sinDLat * sinDLat + Math.cos(lat1) * Math.cos(lat2) * sinDLon * sinDLon;
    return 2 * R * Math.asin(Math.min(1, Math.sqrt(h)));
}

function findNearestPoi(list, lat, lon, maxMeters = 120) {
    let best = null;
    let bestD = Infinity;
    for (const p of list) {
        const d = distMeters(lat, lon, p.location.lat, p.location.lon);
        if (d < bestD) {
            bestD = d;
            best = p;
        }
    }
    if (!best || bestD > maxMeters) return null;
    return {poi: best, distance_m: bestD};
}

function targetKindFromAssignmentStatus(st) {
    const s = normalizeEnum(st);
    if (["ASSIGNED", "FLYING_TO_STORE", "AT_STORE"].includes(s)) return "STORE";
    if (["PICKED_UP_CARGO", "FLYING_TO_CLIENT", "AT_CLIENT"].includes(s)) return "CLIENT";
    if (["DROPPED_CARGO", "RETURNING_BASE", "COMPLETED"].includes(s)) return "BASE";
    return "NONE";
}

function buildDronePopup(dr) {
    const asg = dr.assignment || null;
    const asgStatusTech = asg?.status || "UNKNOWN";
    const asgStatusHuman = asg ? assignmentStatusLabel(asg.status) : "—";
    const targetKindTech = asg ? targetKindFromAssignmentStatus(asg.status) : "NONE";
    const targetKindHuman = targetKindLabel(targetKindTech);

    let targetName = "—";
    let targetAddress = "—";
    let targetCoords = "—";

    const t = asg?.target_location || null;
    if (t && typeof t.lat === "number" && typeof t.lon === "number") {
        targetCoords = `${fmtNum(t.lat, 6)}, ${fmtNum(t.lon, 6)}`;

        if (targetKindTech === "STORE") {
            const n = findNearestPoi(stores, t.lat, t.lon);
            if (n) {
                targetName = n.poi.name;
                targetAddress = n.poi.address || "—";
            }
        } else if (targetKindTech === "BASE") {
            const n = findNearestPoi(bases, t.lat, t.lon);
            if (n) {
                targetName = n.poi.name;
                targetAddress = n.poi.address || "—";
            }
        } else if (targetKindTech === "CLIENT") {
            targetName = "Customer location";
            targetAddress = "—";
        }
    }

    const short = String(dr.drone_id || "").slice(0, 8);

    return `
    <div class="popup">
      <div class="title">Drone <code>${escapeHtml(short)}</code></div>
      <div class="sub"><code>${escapeHtml(dr.drone_id || "")}</code></div>

      <div class="kv"><div class="k">Status</div><div class="v">${escapeHtml(droneStatusLabel(dr.status))}</div></div>
      <div class="kv"><div class="k">Battery</div><div class="v">${escapeHtml(fmtBattery(dr.battery))}</div></div>
      <div class="kv"><div class="k">Speed</div><div class="v">${escapeHtml(fmtSpeed(dr.speed_mps))}</div></div>
      <div class="kv"><div class="k">Location</div><div class="v">${escapeHtml(fmtNum(dr.location?.lat, 6))}, ${escapeHtml(fmtNum(dr.location?.lon, 6))}</div></div>
      <div class="kv"><div class="k">Updated</div><div class="v">${escapeHtml(ageFrom(dr.updated_at_ms))}</div></div>

      <div style="height:8px"></div>

      ${asgStatusHuman !== "—" ? `<div class="kv"><div class="k">Assignment</div><div class="v">${escapeHtml(asgStatusHuman)}</div></div>` : ""}
      ${targetKindHuman !== "None" ? `<div class="kv"><div class="k">Target</div><div class="v">${escapeHtml(targetKindHuman)}</div></div>` : ""}
      ${targetName !== "—" ? `<div class="kv"><div class="k">Target name</div><div class="v">${escapeHtml(targetName)}</div></div>` : ""}
      ${targetAddress !== "—" ? `<div class="kv"><div class="k">Target address</div><div class="v">${escapeHtml(targetAddress)}</div></div>` : ""}
      ${targetCoords !== "—" ? `<div class="kv"><div class="k">Target coords</div><div class="v"><code>${escapeHtml(targetCoords)}</code></div></div>` : ""}

      <div style="height:8px"></div>
      <div class="muted">Raw: status=<code>${escapeHtml(normalizeEnum(dr.status))}</code>, assignment=<code>${escapeHtml(normalizeEnum(asgStatusTech))}</code></div>
    </div>
  `;
}
/* ========= SVG icons ========= */

function svgIcon(html, size = 28, anchor = size / 2) {
    return L.divIcon({
        className: "svgMarker",
        html,
        iconSize: [size, size],
        iconAnchor: [anchor, anchor],
        popupAnchor: [0, -anchor],
        tooltipAnchor: [0, -anchor],
    });
}

function iconBase() {
    return svgIcon(
        `
    <div class="svgMarker__wrap svgMarker__wrap--base">
      <svg viewBox="0 0 24 24" class="svgMarker__svg" aria-hidden="true">
        <path d="M12 3s-6.186 5.34-9.643 8.232c-.203.184-.357.452-.357.768 0 .553.447 1 1 1h2v7c0 .553.447 1 1 1h3c.553 0 1-.448 1-1v-4h4v4c0 .552.447 1 1 1h3c.553 0 1-.447 1-1v-7h2c.553 0 1-.447 1-1 0-.316-.154-.584-.383-.768-3.433-2.892-9.617-8.232-9.617-8.232z"/>
      </svg>
    </div>
  `,
        30
    );
}

function iconStore() {
    return svgIcon(
        `
    <div class="svgMarker__wrap svgMarker__wrap--store">
      <svg viewBox="0 0 24 24" class="svgMarker__svg" aria-hidden="true">
        <path d="M5 9V5C5 3.89543 5.89543 3 7 3H17C18.1046 3 19 3.89543 19 5V9M5 9H19M5 9V15M19 9V15M19 15V19C19 20.1046 18.1046 21 17 21H7C5.89543 21 5 20.1046 5 19V15M19 15H5M8 12H8.01M8 6H8.01M8 18H8.01" stroke="#000000" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </div>
  `,
        30
    );
}

function iconPick() {
    return svgIcon(
        `
    <div class="svgMarker__wrap svgMarker__wrap--pin">
      <svg viewBox="0 0 24 24" class="svgMarker__svg" aria-hidden="true">
        <path fill-rule="evenodd" clip-rule="evenodd" d="M12.5742 21.8187C12.2295 22.0604 11.7699 22.0601 11.4253 21.8184L11.4228 21.8166L11.4172 21.8127L11.3986 21.7994C11.3829 21.7882 11.3607 21.7722 11.3325 21.7517C11.2762 21.7106 11.1956 21.6511 11.0943 21.5741C10.8917 21.4203 10.6058 21.1962 10.2641 20.9101C9.58227 20.3389 8.67111 19.5139 7.75692 18.4988C5.96368 16.5076 4 13.6105 4 10.3636C4 8.16134 4.83118 6.0397 6.32548 4.46777C7.82141 2.89413 9.86146 2 12 2C14.1385 2 16.1786 2.89413 17.6745 4.46777C19.1688 6.0397 20 8.16134 20 10.3636C20 13.6105 18.0363 16.5076 16.2431 18.4988C15.3289 19.5139 14.4177 20.3389 13.7359 20.9101C13.3942 21.1962 13.1083 21.4203 12.9057 21.5741C12.8044 21.6511 12.7238 21.7106 12.6675 21.7517C12.6393 21.7722 12.6171 21.7882 12.6014 21.7994L12.5828 21.8127L12.5772 21.8166L12.5754 21.8179L12.5742 21.8187ZM9 10C9 8.34315 10.3431 7 12 7C13.6569 7 15 8.34315 15 10C15 11.6569 13.6569 13 12 13C10.3431 13 9 11.6569 9 10Z" fill="#000000"/>
      </svg>
    </div>
  `,
        30,
        15
    );
}

function iconDrone(status) {
    const st = normalizeEnum(status);
    return svgIcon(
        `
    <div class="svgMarker__wrap svgMarker__wrap--drone svgMarker__wrap--${escapeHtml(st)}">
      <svg viewBox="0 0 48 48" class="svgMarker__svg svgMarker__svg--fill" aria-hidden="true">
        <rect width="48" height="48" fill="white" fill-opacity="0.01"/>
        <path d="M11 11L19 19M37 37L29 29" stroke="#000000" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M37 11L29 19M11 37L19 29" stroke="#000000" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
        <rect x="19" y="19" width="10" height="10" fill="#2F88FF" stroke="#000000" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M37 18C38.3845 18 39.7379 17.5895 40.889 16.8203C42.0401 16.0511 42.9373 14.9579 43.4672 13.6788C43.997 12.3997 44.1356 10.9922 43.8655 9.63437C43.5954 8.2765 42.9287 7.02922 41.9498 6.05026C40.9708 5.07129 39.7235 4.4046 38.3656 4.13451C37.0078 3.86441 35.6003 4.00303 34.3212 4.53285C33.0421 5.06266 31.9489 5.95987 31.1797 7.11101C30.4105 8.26215 30 9.61553 30 11M37 30C38.3845 30 39.7379 30.4105 40.889 31.1797C42.0401 31.9489 42.9373 33.0421 43.4672 34.3212C43.997 35.6003 44.1356 37.0078 43.8655 38.3656C43.5954 39.7235 42.9287 40.9708 41.9498 41.9497C40.9708 42.9287 39.7235 43.5954 38.3656 43.8655C37.0078 44.1356 35.6003 43.997 34.3212 43.4672C33.0421 42.9373 31.9489 42.0401 31.1797 40.889C30.4105 39.7379 30 38.3845 30 37M11 18C9.61553 18 8.26216 17.5895 7.11101 16.8203C5.95987 16.0511 5.06266 14.9579 4.53285 13.6788C4.00303 12.3997 3.86441 10.9922 4.13451 9.63437C4.4046 8.2765 5.07129 7.02922 6.05026 6.05026C7.02922 5.07129 8.2765 4.4046 9.63437 4.13451C10.9922 3.86441 12.3997 4.00303 13.6788 4.53285C14.9579 5.06266 16.0511 5.95987 16.8203 7.11101C17.5895 8.26215 18 9.61553 18 11M11 30C9.61553 30 8.26216 30.4105 7.11101 31.1797C5.95987 31.9489 5.06266 33.0421 4.53285 34.3212C4.00303 35.6003 3.86441 37.0078 4.13451 38.3656C4.4046 39.7235 5.07129 40.9708 6.05026 41.9497C7.02922 42.9287 8.2765 43.5954 9.63437 43.8655C10.9922 44.1356 12.3997 43.997 13.6788 43.4672C14.9579 42.9373 16.0511 42.0401 16.8203 40.889C17.5895 39.7379 18 38.3845 18 37" stroke="#000000" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </div>
  `,
        34
    );
}

/* ========= API helpers ========= */

async function apiGet(path) {
    const url = `${API_BASE}${path}`;
    const res = await fetch(url, {headers: {Accept: "application/json"}});
    if (!res.ok) {
        let body = "";
        try {
            const ct = res.headers.get("content-type") || "";
            body = ct.includes("application/json") ? JSON.stringify(await res.json()) : await res.text();
        } catch {
        }
        throw new Error(`GET ${path} -> ${res.status} ${res.statusText}${body ? `: ${body}` : ""}`);
    }
    return res.json();
}

async function apiPost(path, payload) {
    const url = `${API_BASE}${path}`;
    const res = await fetch(url, {
        method: "POST",
        headers: {"Content-Type": "application/json", Accept: "application/json"},
        body: JSON.stringify(payload),
    });
    if (!res.ok) {
        let body = "";
        try {
            const ct = res.headers.get("content-type") || "";
            body = ct.includes("application/json") ? JSON.stringify(await res.json()) : await res.text();
        } catch {
        }
        throw new Error(`POST ${path} -> ${res.status} ${res.statusText}${body ? `: ${body}` : ""}`);
    }
    return res.json();
}

function setPickMode(on) {
    pickMode = on;
    btnPickOnMap.textContent = on ? "Picking… (click map)" : "Pick on map";
    btnPickOnMap.classList.toggle("btn--primary", on);
}

function parseItems(s) {
    return String(s || "")
        .split(",")
        .map((x) => x.trim())
        .filter(Boolean);
}

function loadOrdersFromLS() {
    try {
        const raw = localStorage.getItem(LS_KEY);
        if (!raw) return [];
        const arr = JSON.parse(raw);
        if (!Array.isArray(arr)) return [];
        return arr
            .filter((x) => x && typeof x.order_id === "string")
            .map((x) => ({order_id: x.order_id, created_at_ms: Number(x.created_at_ms || 0) || 0}));
    } catch {
        return [];
    }
}

function saveOrdersToLS(list) {
    localStorage.setItem(LS_KEY, JSON.stringify(list));
}

function addOrderToLS(orderId) {
    const list = loadOrdersFromLS();
    if (list.some((o) => o.order_id === orderId)) return;
    list.unshift({order_id: orderId, created_at_ms: Date.now()});
    saveOrdersToLS(list);
}

function removeOrderFromLS(orderId) {
    const list = loadOrdersFromLS().filter((o) => o.order_id !== orderId);
    saveOrdersToLS(list);
}

function clearOrdersLS() {
    localStorage.removeItem(LS_KEY);
}

function renderOrdersSkeleton() {
    const list = loadOrdersFromLS();
    if (list.length === 0) {
        elOrdersList.innerHTML = `<div class="muted">No orders saved yet.</div>`;
        return;
    }
    elOrdersList.innerHTML = list
        .map((o) => {
            const short = o.order_id.slice(0, 8);
            return `
        <div class="order" data-order-id="${escapeHtml(o.order_id)}">
          <div class="row1">
            <div class="id">Order <code>${escapeHtml(short)}</code></div>
            <div class="status status--PENDING">…</div>
          </div>
          <div class="meta">
            <div>Drone: <span class="drone">—</span></div>
            <div>Location: <span class="loc">—</span></div>
            <div>Updated: <span class="upd">—</span></div>
          </div>
          <div class="actions">
            <button class="btn btn--ghost act-focus" type="button">Focus</button>
            <button class="btn btn--danger act-remove" type="button">Remove</button>
          </div>
        </div>
      `;
        })
        .join("");
}

function updateOrderCard(order, fetchedAtMs) {
    const el = elOrdersList.querySelector(`.order[data-order-id="${CSS.escape(order.order_id)}"]`);
    if (!el) return;

    const tech = normalizeEnum(order.status);
    const human = orderStatusLabel(order.status);

    const statusEl = el.querySelector(".status");
    statusEl.textContent = human || "—";
    statusEl.className = `status status--${escapeHtml(tech)}`;

    el.querySelector(".drone").textContent = order.drone_id || "—";
    el.querySelector(".loc").textContent = order.delivery_location
        ? `${fmtNum(order.delivery_location.lat, 6)}, ${fmtNum(order.delivery_location.lon, 6)}`
        : "—";
    el.querySelector(".upd").textContent = ageFrom(fetchedAtMs);

    if (order.delivery_location && typeof order.delivery_location.lat === "number") {
        const id = order.order_id;
        const lat = order.delivery_location.lat;
        const lon = order.delivery_location.lon;

        let marker = orderMarkers.get(id);
        if (!marker) {
            marker = L.circleMarker([lat, lon], {radius: 6, weight: 2, opacity: 0.9, fillOpacity: 0.65}).addTo(
                ordersLayer
            );
            marker.bindTooltip(`Order ${id.slice(0, 8)} • ${human}`, {direction: "top"});
            marker.bindPopup(`
        <div class="popup">
          <div class="title">Order <code>${escapeHtml(id.slice(0, 8))}</code></div>
          <div class="sub"><code>${escapeHtml(id)}</code></div>
          <div class="kv"><div class="k">Status</div><div class="v">${escapeHtml(human)}</div></div>
          <div class="kv"><div class="k">Drone</div><div class="v">${escapeHtml(order.drone_id || "—")}</div></div>
          <div class="kv"><div class="k">Location</div><div class="v"><code>${fmtNum(lat, 6)}, ${fmtNum(lon, 6)}</code></div></div>
          <div class="muted">Raw status: <code>${escapeHtml(tech)}</code></div>
        </div>
      `);
            orderMarkers.set(id, marker);
        } else {
            marker.setLatLng([lat, lon]);
            marker.setTooltipContent(`Order ${id.slice(0, 8)} • ${human}`);
            if (marker.isPopupOpen()) {
                marker.setPopupContent(`
          <div class="popup">
            <div class="title">Order <code>${escapeHtml(id.slice(0, 8))}</code></div>
            <div class="sub"><code>${escapeHtml(id)}</code></div>
            <div class="kv"><div class="k">Status</div><div class="v">${escapeHtml(human)}</div></div>
            <div class="kv"><div class="k">Drone</div><div class="v">${escapeHtml(order.drone_id || "—")}</div></div>
            <div class="kv"><div class="k">Location</div><div class="v"><code>${fmtNum(lat, 6)}, ${fmtNum(lon, 6)}</code></div></div>
            <div class="muted">Raw status: <code>${escapeHtml(tech)}</code></div>
          </div>
        `);
            }
        }

        let color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-default").trim() || "#94a3b8";
        if (tech === "PENDING" || tech === "CREATED")
            color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-pending").trim() || "#38bdf8";
        if (tech === "ASSIGNED" || tech === "DELIVERING")
            color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-assigned").trim() || "#f59e0b";
        if (tech === "COMPLETED")
            color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-completed").trim() || "#22c55e";
        if (tech === "FAILED" || tech === "CANCELED")
            color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-failed").trim() || "#ef4444";
        marker.setStyle({color, fillColor: color});
    }
}

async function refreshOrders() {
    const list = loadOrdersFromLS();
    if (list.length === 0) return;

    await Promise.all(
        list.map(async (o) => {
            const t = Date.now();
            try {
                const resp = await apiGet(`/orders/${encodeURIComponent(o.order_id)}`);
                updateOrderCard(
                    {
                        order_id: resp.order_id,
                        user_id: resp.user_id,
                        drone_id: resp.drone_id,
                        items: resp.items,
                        status: resp.status,
                        delivery_location: resp.delivery_location,
                    },
                    t
                );
            } catch (e) {
                const el = elOrdersList.querySelector(`.order[data-order-id="${CSS.escape(o.order_id)}"]`);
                if (el) {
                    const st = el.querySelector(".status");
                    st.textContent = "Error";
                    st.className = "status status--FAILED";
                    el.querySelector(".upd").textContent = String(e.message || e);
                }
            }
        })
    );
}

function attachOrderListHandlers(map) {
    elOrdersList.addEventListener("click", (ev) => {
        const btn = ev.target.closest("button");
        if (!btn) return;
        const card = ev.target.closest(".order");
        if (!card) return;
        const id = card.getAttribute("data-order-id");
        if (!id) return;

        if (btn.classList.contains("act-remove")) {
            removeOrderFromLS(id);
            const m = orderMarkers.get(id);
            if (m) {
                ordersLayer.removeLayer(m);
                orderMarkers.delete(id);
            }
            renderOrdersSkeleton();
            refreshOrders();
            return;
        }

        if (btn.classList.contains("act-focus")) {
            const m = orderMarkers.get(id);
            if (m) {
                map.setView(m.getLatLng(), Math.max(map.getZoom(), 15), {animate: true});
                m.openPopup();
            }
        }
    });
}

async function bootstrapPois() {
    const [storesResp, basesResp] = await Promise.all([apiGet("/stores"), apiGet("/bases")]);
    stores = Array.isArray(storesResp.items) ? storesResp.items : [];
    bases = Array.isArray(basesResp.items) ? basesResp.items : [];
}

function renderPois() {
    storesLayer.clearLayers();
    basesLayer.clearLayers();

    for (const s of stores) {
        if (!s.location) continue;

        const m = L.marker([s.location.lat, s.location.lon], {icon: iconStore()});
        m.bindTooltip(`Store: ${escapeHtml(s.name)}`, {direction: "top", offset: [0, -8]});
        m.bindPopup(`
      <div class="popup">
        <div class="title">Store</div>
        <div class="sub">${escapeHtml(s.name)}</div>
        <div class="kv"><div class="k">ID</div><div class="v"><code>${escapeHtml(s.store_id)}</code></div></div>
        <div class="kv"><div class="k">Address</div><div class="v">${escapeHtml(s.address || "—")}</div></div>
        <div class="kv"><div class="k">Location</div><div class="v"><code>${fmtNum(s.location.lat, 6)}, ${fmtNum(
            s.location.lon,
            6
        )}</code></div></div>
      </div>
    `);
        m.addTo(storesLayer);
    }

    for (const b of bases) {
        if (!b.location) continue;

        const m = L.marker([b.location.lat, b.location.lon], {icon: iconBase()});
        m.bindTooltip(`Base: ${escapeHtml(b.name)}`, {direction: "top", offset: [0, -8]});
        m.bindPopup(`
      <div class="popup">
        <div class="title">Base</div>
        <div class="sub">${escapeHtml(b.name)}</div>
        <div class="kv"><div class="k">ID</div><div class="v"><code>${escapeHtml(b.base_id)}</code></div></div>
        <div class="kv"><div class="k">Address</div><div class="v">${escapeHtml(b.address || "—")}</div></div>
        <div class="kv"><div class="k">Location</div><div class="v"><code>${fmtNum(b.location.lat, 6)}, ${fmtNum(
            b.location.lon,
            6
        )}</code></div></div>
      </div>
    `);
        m.addTo(basesLayer);
    }
}

function fitDrones(map) {
    const latLngs = [];
    for (const {marker} of droneMarkers.values()) latLngs.push(marker.getLatLng());
    if (latLngs.length === 0) return;
    map.fitBounds(L.latLngBounds(latLngs).pad(0.25), {animate: true});
}

async function refreshDrones(map) {
    if (dronesUpdateInFlight) return;
    dronesUpdateInFlight = true;

    try {
        const resp = await apiGet("/drones");
        const items = Array.isArray(resp.items) ? resp.items : [];

        lastDronesOkAt = Date.now();
        setApiPill("ok", `API: OK • drones=${items.length}`);

        const alive = new Set();

        for (const dr of items) {
            if (!dr || !dr.drone_id || !dr.location) continue;
            alive.add(dr.drone_id);

            const lat = dr.location.lat;
            const lon = dr.location.lon;

            const old = droneMarkers.get(dr.drone_id);
            const icon = iconDrone(dr.status);
            const key = normalizeEnum(dr.status);

            const tooltip = `${dr.drone_id.slice(0, 8)} • ${droneStatusLabel(dr.status)} • ${fmtBattery(dr.battery)}`;

            if (!old) {
                const marker = L.marker([lat, lon], {icon});
                marker.__dr = dr;
                marker.bindTooltip(tooltip, {direction: "top", offset: [0, -14], opacity: 0.95});
                marker.bindPopup(buildDronePopup(dr));
                marker.addTo(dronesLayer);
                droneMarkers.set(dr.drone_id, {marker, lastKey: key});
            } else {
                const marker = old.marker;
                marker.__dr = dr;
                marker.setLatLng([lat, lon]);
                if (old.lastKey !== key) {
                    marker.setIcon(icon);
                    old.lastKey = key;
                }
                marker.setTooltipContent(tooltip);
                if (marker.isPopupOpen()) marker.setPopupContent(buildDronePopup(dr));
            }
        }

        for (const [id, obj] of droneMarkers.entries()) {
            if (!alive.has(id)) {
                dronesLayer.removeLayer(obj.marker);
                droneMarkers.delete(id);
            }
        }
    } catch (e) {
        const age = lastDronesOkAt ? ageFrom(lastDronesOkAt) : "—";
        setApiPill("bad", `API: ERROR • last OK ${age}`);
        console.warn("refreshDrones error:", e);
    } finally {
        dronesUpdateInFlight = false;
    }
}

function setCreateResult(html) {
    elOrderCreateResult.innerHTML = html || "";
}

function uuidv4() {
    const b = new Uint8Array(16);
    crypto.getRandomValues(b);
    b[6] = (b[6] & 0x0f) | 0x40; // version 4
    b[8] = (b[8] & 0x3f) | 0x80; // variant
    const h = [...b].map(x => x.toString(16).padStart(2, "0")).join("");
    return `${h.slice(0,8)}-${h.slice(8,12)}-${h.slice(12,16)}-${h.slice(16,20)}-${h.slice(20)}`;
}

async function handleCreateOrderSubmit(ev) {
    ev.preventDefault();

    const items = parseItems(elItems.value);
    const lat = Number(elLat.value);
    const lon = Number(elLon.value);

    if (items.length === 0) {
        setCreateResult(`<span class="text-danger">items required</span>`);
        return;
    }
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
        setCreateResult(`<span class="text-danger">delivery lat/lon required</span>`);
        return;
    }

    setCreateResult(`<span class="muted">Creating…</span>`);

    try {
        const user_id = crypto?.randomUUID ? crypto.randomUUID() : uuidv4();
        const resp = await apiPost("/orders", {user_id, items, delivery_location: {lat, lon}});

        addOrderToLS(resp.order_id);
        renderOrdersSkeleton();
        await refreshOrders();

        setCreateResult(`
      <div class="result-ok">
        <div><strong>OK</strong> — order_id: <code>${escapeHtml(resp.order_id)}</code></div>
        <div>Status: <code>${escapeHtml(orderStatusLabel(resp.status))}</code></div>
        <div>Drone: <code>${escapeHtml(resp.drone_id || "—")}</code></div>
        <div>ETA: <code>${escapeHtml(resp.eta_seconds ?? "—")}</code> sec</div>
      </div>
    `);
    } catch (e) {
        setCreateResult(`<span class="text-danger">${escapeHtml(e.message || String(e))}</span>`);
    }
}

function initMap() {
    const map = L.map("map", {zoomControl: true, attributionControl: false}).setView([55.75, 37.62], 11);

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        maxZoom: 19,
    }).addTo(map);

    storesLayer = L.layerGroup().addTo(map);
    basesLayer = L.layerGroup().addTo(map);
    dronesLayer = L.layerGroup().addTo(map);
    ordersLayer = L.layerGroup().addTo(map);

    map.on("click", (ev) => {
        if (!pickMode) return;

        const {lat, lng} = ev.latlng;
        elLat.value = String(lat);
        elLon.value = String(lng);

        if (!pickMarker) {
            pickMarker = L.marker([lat, lng], {icon: iconPick()}).addTo(map);
            pickMarker.bindTooltip("Delivery location", {direction: "top", offset: [0, -14]});
        } else {
            pickMarker.setLatLng([lat, lng]);
        }
    });

    return map;
}

async function main() {
    const map = initMap();

    btnPickOnMap.addEventListener("click", () => setPickMode(!pickMode));
    document.getElementById("orderForm").addEventListener("submit", handleCreateOrderSubmit);

    document.getElementById("btnRefreshOrders").addEventListener("click", async () => {
        renderOrdersSkeleton();
        await refreshOrders();
    });

    document.getElementById("btnClearOrders").addEventListener("click", () => {
        clearOrdersLS();
        for (const m of orderMarkers.values()) ordersLayer.removeLayer(m);
        orderMarkers.clear();
        renderOrdersSkeleton();
    });

    attachOrderListHandlers(map);
    renderOrdersSkeleton();

    setApiPill("gray", "API: bootstrapping…");
    try {
        await bootstrapPois();
        renderPois();
        setApiPill("ok", "API: OK • bootstrapped");
    } catch (e) {
        setApiPill("bad", "API: bootstrap failed");
        console.warn("bootstrapPois error:", e);
    }

    await refreshDrones(map);
    await refreshOrders();

    setInterval(() => refreshDrones(map), DRONES_POLL_MS);
    setInterval(() => refreshOrders(), ORDERS_POLL_MS);

    document.getElementById("btnFit").addEventListener("click", () => fitDrones(map));
}

window.addEventListener("DOMContentLoaded", main);
