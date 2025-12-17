const API_BASE = "/api/v1";
const DRONES_POLL_MS = 500;
const ORDERS_POLL_MS = 1500;

document.getElementById("dronesPollMs").textContent = String(DRONES_POLL_MS);
document.getElementById("ordersPollMs").textContent = String(ORDERS_POLL_MS);

const elApiStatus = document.getElementById("apiStatus");
const elOrdersList = document.getElementById("ordersList");
const elOrderCreateResult = document.getElementById("orderCreateResult");

const elUserId = document.getElementById("userId");
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
    return s
        .replace(/^.*STATUS_/, "")
        .replace(/^.*TARGET_/, "")
        .replace(/^.*ACTION_/, "")
        .replace(/^.*EVENT_/, "")
        .replace(/[^A-Z0-9_]/g, "_")
        .replace(/_+/g, "_")
        .replace(/^_+|_+$/g, "") || "UNKNOWN";
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
    let targetCoords = "—";

    const t = asg?.target_location || null;
    if (t && typeof t.lat === "number" && typeof t.lon === "number") {
        targetCoords = `${fmtNum(t.lat, 6)}, ${fmtNum(t.lon, 6)}`;
        if (targetKindTech === "STORE") {
            const n = findNearestPoi(stores, t.lat, t.lon);
            if (n) targetName = `${n.poi.name} (${Math.round(n.distance_m)}m)`;
        } else if (targetKindTech === "BASE") {
            const n = findNearestPoi(bases, t.lat, t.lon);
            if (n) targetName = `${n.poi.name} (${Math.round(n.distance_m)}m)`;
        } else if (targetKindTech === "CLIENT") {
            targetName = "Customer location";
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

      <div class="kv"><div class="k">Assignment</div><div class="v">${escapeHtml(asgStatusHuman)}</div></div>
      <div class="kv"><div class="k">Target</div><div class="v">${escapeHtml(targetKindHuman)}</div></div>
      <div class="kv"><div class="k">Target name</div><div class="v">${escapeHtml(targetName)}</div></div>
      <div class="kv"><div class="k">Target coords</div><div class="v"><code>${escapeHtml(targetCoords)}</code></div></div>

      <div style="height:8px"></div>
      <div class="muted">Raw: status=<code>${escapeHtml(normalizeEnum(dr.status))}</code>, assignment=<code>${escapeHtml(normalizeEnum(asgStatusTech))}</code></div>
    </div>
  `;
}

function droneIcon(dr) {
    const status = normalizeEnum(dr.status);
    const short = String(dr.drone_id || "").slice(0, 4).toUpperCase();
    const key = `${status}:${short}`;
    const icon = L.divIcon({
        className: "",
        html: `<div class="droneIcon droneIcon--${escapeHtml(status)}"><div class="inner">${escapeHtml(short)}</div></div>`,
        iconSize: [42, 42],
        iconAnchor: [21, 21],
        popupAnchor: [0, -18],
        tooltipAnchor: [0, -18],
    });
    return {icon, key};
}

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
        if (tech === "PENDING" || tech === "CREATED") color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-pending").trim() || "#38bdf8";
        if (tech === "ASSIGNED" || tech === "DELIVERING") color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-assigned").trim() || "#f59e0b";
        if (tech === "COMPLETED") color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-completed").trim() || "#22c55e";
        if (tech === "FAILED" || tech === "CANCELED") color = getComputedStyle(document.documentElement).getPropertyValue("--c-order-failed").trim() || "#ef4444";
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
        const icon = L.divIcon({
            className: "",
            html: `<div class="poi poi--store"></div>`,
            iconSize: [16, 16],
            iconAnchor: [8, 8],
        });

        const m = L.marker([s.location.lat, s.location.lon], {icon});
        m.bindTooltip(`Store: ${escapeHtml(s.name)}`, {direction: "top", offset: [0, -8]});
        m.bindPopup(`
      <div class="popup">
        <div class="title">Store</div>
        <div class="sub">${escapeHtml(s.name)}</div>
        <div class="kv"><div class="k">ID</div><div class="v"><code>${escapeHtml(s.store_id)}</code></div></div>
        <div class="kv"><div class="k">Address</div><div class="v">${escapeHtml(s.address || "—")}</div></div>
        <div class="kv"><div class="k">Location</div><div class="v"><code>${fmtNum(s.location.lat, 6)}, ${fmtNum(s.location.lon, 6)}</code></div></div>
      </div>
    `);
        m.addTo(storesLayer);
    }

    for (const b of bases) {
        if (!b.location) continue;
        const icon = L.divIcon({
            className: "",
            html: `<div class="poi poi--base"></div>`,
            iconSize: [16, 16],
            iconAnchor: [8, 8],
        });

        const m = L.marker([b.location.lat, b.location.lon], {icon});
        m.bindTooltip(`Base: ${escapeHtml(b.name)}`, {direction: "top", offset: [0, -8]});
        m.bindPopup(`
      <div class="popup">
        <div class="title">Base</div>
        <div class="sub">${escapeHtml(b.name)}</div>
        <div class="kv"><div class="k">ID</div><div class="v"><code>${escapeHtml(b.base_id)}</code></div></div>
        <div class="kv"><div class="k">Address</div><div class="v">${escapeHtml(b.address || "—")}</div></div>
        <div class="kv"><div class="k">Location</div><div class="v"><code>${fmtNum(b.location.lat, 6)}, ${fmtNum(b.location.lon, 6)}</code></div></div>
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
            const {icon, key} = droneIcon(dr);

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

function ensureUserId() {
    if (!elUserId.value.trim()) {
        elUserId.value = crypto?.randomUUID ? crypto.randomUUID() : "00000000-0000-0000-0000-000000000000";
    }
}

function setCreateResult(html) {
    elOrderCreateResult.innerHTML = html || "";
}

async function handleCreateOrderSubmit(ev) {
    ev.preventDefault();

    ensureUserId();

    const user_id = elUserId.value.trim();
    const items = parseItems(elItems.value);
    const lat = Number(elLat.value);
    const lon = Number(elLon.value);

    if (!user_id) {
        setCreateResult(`<span class="text-danger">user_id required</span>`);
        return;
    }
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
    const map = L.map("map", {zoomControl: true}).setView([55.75, 37.62], 11);

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
        maxZoom: 19,
        attribution: "&copy; OpenStreetMap contributors",
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
            pickMarker = L.circleMarker([lat, lng], {radius: 7, weight: 2, opacity: 0.9, fillOpacity: 0.25}).addTo(map);
            pickMarker.bindTooltip("Delivery location", {direction: "top"});
        } else {
            pickMarker.setLatLng([lat, lng]);
        }
    });

    return map;
}

async function main() {
    const map = initMap();

    ensureUserId();
    document.getElementById("btnGenUser").addEventListener("click", () => {
        if (crypto?.randomUUID) elUserId.value = crypto.randomUUID();
    });

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
