/* ===================================================================
 * GameAgentEngine Demo App — 灰港边境
 * 纯用户态前端，通过 HTTP API 与引擎交互。
 * =================================================================== */

/* ============= Config ============= */
const DEMO_VERSION = "v0.4.5";
const CFG_KEY = "ga_demo_cfg";
const MSG_KEY = "ga_demo_msgs_"; // + advisorId
const ADV_KEY = "ga_demo_advisor";

function loadCfg() { try { return JSON.parse(localStorage.getItem(CFG_KEY)) || {}; } catch(e) { return {}; } }
function saveCfg(c) { localStorage.setItem(CFG_KEY, JSON.stringify(c)); }

let cfg = loadCfg();
if (!cfg.url) cfg.url = "http://127.0.0.1:8080";
if (!cfg.key) cfg.key = "dev-key";

/* ============= State ============= */
let state = {
  connected: false,
  world: null,         // { id, name }
  nodes: [],
  resources: null,     // resource_state
  districts: null,     // district_state nodes
  timeline: null,      // timeline component
  incidentIndex: 0,
  hasActed: false,
  actionsTaken: [],
  history: [],
  turn: 1,
  selectedChoice: null,
  advisors: [],
  selectedAdvisor: null,
  loading: false,
};

/* ============= Utilities ============= */
function $(id) { return document.getElementById(id); }

function show(id) { const e = $(id); if (e) e.classList.add("active"); }
function hide(id) { const e = $(id); if (e) e.classList.remove("active"); }
function setText(id, t) { const e = $(id); if (e) e.textContent = t; }
function setHTML(id, h) { const e = $(id); if (e) e.innerHTML = h; }
function disable(id, d) { const e = $(id); if (e) e.disabled = d; }

function setLoading(on, msg) {
  state.loading = on;
  if (on) { show("loadingOverlay"); setText("loadingText", msg || "加载中..."); }
  else { hide("loadingOverlay"); }
}

function toast(msg) {
  const t = document.createElement("div");
  t.className = "toast";
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => { t.classList.add("show"); }, 10);
  setTimeout(() => { t.remove(); }, 2300);
}

/* ============= API ============= */
async function api(method, path, body) {
  const url = cfg.url.replace(/\/+$/, "") + path;
  const opts = { method, headers: { "X-API-Key": cfg.key, "Content-Type": "application/json" } };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const res = await fetch(url, opts);
  const text = await res.text();
  if (!res.ok) throw new Error(text);
  try { return JSON.parse(text); } catch(e) { return text; }
}

/* ============= Connection & Init ============= */
async function checkConnection() {
  try {
    const info = await api("GET", "/health");
    state.connected = info.status === "ok";
    const v = info.version || "?";
    $("statusDot").className = "dot on";
    setText("statusLabel", "已连接 v" + v);
    // version check
    if (v && DEMO_VERSION) {
      const dv = DEMO_VERSION.replace(/^v/i, "");
      const ev = v.replace(/^v/i, "");
      const dp = dv.split(".");
      const ep = ev.split(".");
      if (dp[0] !== ep[0] || dp[1] !== ep[1]) {
        setText("versionBadge", "版本不匹配");
        $("versionBadge").style.borderColor = "var(--red)";
        $("versionBadge").style.color = "var(--red)";
      } else {
        setText("versionBadge", "引擎 v" + ev);
      }
    }
    return true;
  } catch(e) {
    state.connected = false;
    $("statusDot").className = "dot off";
    setText("statusLabel", "未连接");
    return false;
  }
}

async function ensureWorld() {
  const worlds = await api("GET", "/api/v1/worlds");
  if (worlds && worlds.length > 0) {
    state.world = worlds[0];
    return true;
  }
  return false;
}

/* ============= Load World Data ============= */
async function loadWorldData() {
  if (!state.world) return;
  // Load nodes
  state.nodes = await api("GET", "/api/v1/nodes?world_id=" + encodeURIComponent(state.world.id));
  // Find world node details
  const worldDetail = await api("GET", "/api/v1/nodes/" + encodeURIComponent(state.world.id));
  // Parse components
  const comps = worldDetail.components || [];
  for (const c of comps) {
    if (c.component_type === "resource_state") {
      try { state.resources = JSON.parse(c.data); } catch(e) {}
    } else if (c.component_type === "timeline") {
      try { state.timeline = JSON.parse(c.data); } catch(e) {}
    } else if (c.component_type === "demo_state") {
      try {
        const ds = JSON.parse(c.data);
        state.incidentIndex = ds.incident_index || 0;
        state.hasActed = ds.has_acted_this_turn || false;
        state.actionsTaken = ds.actions_taken || [];
        state.history = ds.history || [];
        state.turn = (ds.actions_taken ? ds.actions_taken.length : 0) + 1;
      } catch(e) {}
    }
  }
  // Parse district states from all nodes
  state.districts = {};
  for (const n of state.nodes) {
    if (n.node_type === "location") {
      const nd = await api("GET", "/api/v1/nodes/" + encodeURIComponent(n.id));
      for (const c of (nd.components || [])) {
        if (c.component_type === "district_state") {
          try { state.districts[n.id] = { name: n.name, data: JSON.parse(c.data) }; } catch(e) {}
        }
      }
    }
  }
  // Find NPC advisors
  state.advisors = state.nodes.filter(n => n.node_type === "npc");
  // Load world settings
  try { state.settings = await api("GET", "/api/v1/worlds/" + encodeURIComponent(state.world.id) + "/settings"); } catch(e) {}
  // Load memories for context
  try { state.memories = await api("GET", "/api/v1/memories?world_id=" + encodeURIComponent(state.world.id) + "&limit=20"); } catch(e) {}
}

/* ============= Incidents ============= */
const INCIDENTS = [
  {
    id: "miners_unrest",
    title: "雾湾矿场工潮逼近失控",
    severity: "高危",
    focus: "雾湾矿场",
    body: "矿工连续抱怨工时、配给和安全问题。你必须判断是先止血、先保产，还是把风险往后压。",
    clues: ["矿区怨气抬头", "补给线在吃紧", "议会声望承压"],
    choices: [
      { id: "grant_relief", title: "先拨给救济，稳住矿工", description: "立刻调拨粮食和药品，优先换取矿区情绪回稳。", hints: ["安抚民心", "财政承压", "利于后续谈判"] },
      { id: "merchant_loan", title: "向商会借补给和周转", description: "用未来贸易让利换取眼前补给，让矿区迅速恢复运转。", hints: ["短期止损", "商会影响上升", "形成依赖"] },
      { id: "military_suppression", title: "派守军强压工潮", description: "直接压下混乱，表面最稳，但会把后果留到后面。", hints: ["立刻保产", "怨气累积", "军方势力更强"] },
    ],
  },
  {
    id: "north_gate_breach",
    title: "北门要塞外墙急需修补",
    severity: "高危",
    focus: "北门要塞",
    body: "北门外墙出现新的裂缝，守军坚持说再拖下去就是把未来押给运气。",
    clues: ["工事维护不足", "修缮材料短缺", "守军信任在摇晃"],
    choices: [
      { id: "fortify_now", title: "集中工匠立即修缮", description: "优先把钱和材料砸向北门，让后方项目让位。", hints: ["防线回稳", "财政失血", "军心回升"] },
      { id: "delay_and_watch", title: "继续观望，加强巡逻", description: "先省下大修成本，把问题往后推。", hints: ["暂缓开支", "前线不满", "风险后移"] },
      { id: "militia_mobilize", title: "动员民兵与商会车队", description: "用折中方案先把缺口补上，但代价会扩散到后方。", hints: ["多方分摊", "协调复杂", "短期有效"] },
    ],
  },
  {
    id: "market_tension",
    title: "河岸集市贸易争端激化",
    severity: "中危",
    focus: "河岸集市",
    body: "铁潮商会与本地粮商因定价权发生冲突，集市流通开始受阻。",
    clues: ["粮价波动", "商会施压", "民众不满"],
    choices: [
      { id: "price_control", title: "临时实行价格管制", description: "设定粮食和盐铁的最高限价，稳定民生。", hints: ["短期稳定", "商会抵制", "执行成本高"] },
      { id: "mediate_talk", title: "召集双方谈判调解", description: "议会出面协调商会与粮商，争取折中方案。", hints: ["缓和矛盾", "时间消耗", "考验议长威信"] },
      { id: "let_market", title: "不干预，由市场自行调节", description: "放手让市场自己找到平衡点，只加强治安。", hints: ["零成本", "短期阵痛", "可能激化"] },
    ],
  },
];

function getCurrentIncident() {
  return INCIDENTS[state.incidentIndex] || INCIDENTS[0];
}

/* ============= World Init ============= */
async function initWorld() {
  if (state.loading) return;
  setLoading(true, "正在连接引擎...");
  try {
    await checkConnection();
    if (!state.connected) { setLoading(false); toast("无法连接到引擎"); return; }
    setLoading(true, "正在加载世界数据...");
    if (!(await ensureWorld())) {
      // Try to import via creator import API
      try {
        await api("POST", "/api/v1/creator/import", {
          format: "yaml",
          content: "demo-world",
          reset: true,
          dry_run: false,
        });
        toast("世界创建成功");
      } catch(importErr) {
        setLoading(false);
        toast("请先通过 DevCli 导入世界: GameAgentDevCli import demo-world.yaml --reset");
        return;
      }
      await ensureWorld();
    }
    await loadWorldData();
    renderAll();
    setLoading(false);
    disable("advanceBtn", false);
    toast("世界已就绪");
  } catch(e) {
    setLoading(false);
    toast("初始化失败: " + e.message);
  }
}

/* ============= Render ============= */
function renderAll() {
  renderOverview();
  renderIncident();
  renderResources();
  disable("initBtn", true);
  $("versionBadge") && setText("versionBadge", "v" + (state.settings ? "就绪" : ""));
}

function renderOverview() {
  const ov = $("worldOverview");
  if (!ov) return;
  let html = '<div class="overview-section"><h4>当前回合</h4><p>第 ' + state.turn + ' 回合 · ' + (state.hasActed ? "已做出决策" : "待决策") + '</p></div>';
  html += '<div class="overview-section"><h4>边境派系</h4>';
  const factions = state.nodes.filter(n => n.node_type === "faction");
  for (const f of factions) {
    html += '<span class="faction-chip">' + escapeHtml(f.name) + '</span> ';
  }
  html += '</div>';
  html += '<div class="overview-section"><h4>顾问</h4>';
  for (const a of state.advisors) {
    html += '<span class="npc-chip" data-id="' + a.id + '">' + escapeHtml(a.name) + ' <span class="hint">' + advisorRole(a) + '</span></span> ';
  }
  html += '</div>';
  if (state.history && state.history.length > 0) {
    html += '<div class="overview-section"><h4>历史记录</h4>';
    for (const h of state.history.slice(-3)) {
      html += '<p style="font-size:12px;color:var(--text2);margin-bottom:4px;">\u2022 ' + escapeHtml(h) + '</p>';
    }
    html += '</div>';
  }
  ov.innerHTML = html;
}

function renderIncident() {
  const area = $("incidentArea");
  if (!area) return;
  const inc = getCurrentIncident();
  if (state.hasActed) {
    area.innerHTML = '<div class="result-card"><h4>\u2713 本回合已决策</h4><p>等待推进到下一回合。点击"推进回合"按钮继续。</p></div>';
    return;
  }
  let html = '<div class="incident-card">';
  html += '<h3>' + escapeHtml(inc.title) + '<span class="sev">' + inc.severity + '</span></h3>';
  html += '<p class="desc">' + escapeHtml(inc.body) + '</p>';
  html += '<div class="clues">';
  for (const c of inc.clues) html += '<span class="clue">' + escapeHtml(c) + '</span> ';
  html += '</div>';
  html += '<div class="choices">';
  for (const ch of inc.choices) {
    html += '<div class="choice-btn" data-choice="' + ch.id + '">';
    html += '<h4>' + escapeHtml(ch.title) + '</h4>';
    html += '<p class="ch-desc">' + escapeHtml(ch.description) + '</p>';
    html += '<div class="ch-hints">';
    if (ch.hints) for (const h of ch.hints) html += '<span>' + escapeHtml(h) + '</span> ';
    html += '</div></div>';
  }
  html += '</div></div>';
  area.innerHTML = html;
  // Bind choice clicks
  area.querySelectorAll(".choice-btn").forEach(el => {
    el.addEventListener("click", () => selectChoice(el.dataset.choice));
  });
  updateTurnBadge();
}

function renderResources() {
  const rp = $("resourcePanel");
  if (!rp) return;
  if (!state.resources) { rp.innerHTML = '<div class="placeholder">无资源数据</div>'; return; }
  let html = '<div class="resource-grid">';
  const keys = [
    { k: "food", l: "粮食" },
    { k: "order", l: "秩序" },
    { k: "defense", l: "防线" },
    { k: "morale", l: "士气" },
    { k: "treasury", l: "财政" },
  ];
  for (const r of keys) {
    const v = state.resources[r.k] || 0;
    const cls = v >= 50 ? "good" : v >= 30 ? "warn" : "bad";
    html += '<div class="resource-row"><span class="label">' + r.l + '</span><div class="bar"><div class="bar-fill ' + cls + '" style="width:' + v + '%"></div></div><span class="val">' + v + '</span></div>';
  }
  html += '</div>';
  // District cards
  for (const did in state.districts) {
    const d = state.districts[did];
    html += '<div class="district-card"><h5>' + escapeHtml(d.name) + '</h5>';
    html += '<div class="dd">稳定' + (d.data.stability || "?") + ' | 压力' + (d.data.pressure || "?") + ' | 产出' + (d.data.output || "?") + '</div>';
    if (d.data.summary) html += '<div style="font-size:11px;color:var(--text2);margin-top:4px;">' + escapeHtml(d.data.summary) + '</div>';
    html += '</div>';
  }
  rp.innerHTML = html;
}

function updateTurnBadge() {
  setText("turnBadge", "第 " + state.turn + " 回合");
}

/* ============= Choice & Advance ============= */
function selectChoice(choiceId) {
  state.selectedChoice = choiceId;
  // Visual update
  document.querySelectorAll(".choice-btn").forEach(el => {
    el.classList.toggle("selected", el.dataset.choice === choiceId);
  });
}

async function advanceTurn() {
  if (state.loading || !state.world) return;
  if (!state.selectedChoice && !state.hasActed) {
    toast("请先选择一个决策方案");
    return;
  }
  setLoading(true, "正在推进回合...");
  try {
    // 1. Record the decision as a memory
    const inc = getCurrentIncident();
    const choice = inc.choices.find(c => c.id === state.selectedChoice);
    if (choice) {
      await api("POST", "/api/v1/memories", {
        node_id: state.world.id,
        content: "第" + state.turn + "回合决议: " + choice.title + " — " + choice.description,
        level: "world",
        tags: "demo,tick,decision",
      });
    }
    // 2. Call world tick via pipeline
    const invokeResp = await api("POST", "/api/v1/worlds/" + encodeURIComponent(state.world.id) + "/ticks/advance", {
      task_type: "world_tick",
      node_id: state.world.id,
    });
    // 3. Update state
    state.turn++;
    state.hasActed = false;
    state.selectedChoice = null;
    state.incidentIndex = (state.incidentIndex + 1) % INCIDENTS.length;
    if (!state.history) state.history = [];
    state.history.push("第" + (state.turn - 1) + "回合: " + (choice ? choice.title : "推进"));
    // 4. Reload world data
    await loadWorldData();
    renderAll();
    setLoading(false);
    toast("回合已推进到第 " + state.turn + " 回合");
  } catch(e) {
    setLoading(false);
    toast("推进失败: " + e.message);
  }
}

/* ============= Advisor Dialogue ============= */
function advisorRole(a) {
  try {
    const p = JSON.parse(a.profile || "{}");
    return p.title || a.node_type;
  } catch(e) { return a.node_type; }
}

async function selectAdvisor(advisorId) {
  state.selectedAdvisor = state.advisors.find(a => a.id === advisorId) || null;
  renderAdvisorTabs();
  renderChatLog();
  $("chatInput").disabled = !state.selectedAdvisor;
  $("sendBtn").disabled = !state.selectedAdvisor;
  if (state.selectedAdvisor) {
    setText("chatTitle", state.selectedAdvisor.name);
  }
}

function renderAdvisorTabs() {
  const tabs = $("advisorTabs");
  if (!tabs) return;
  tabs.innerHTML = "";
  for (const a of state.advisors) {
    const chip = document.createElement("div");
    chip.className = "advisor-chip" + (state.selectedAdvisor && state.selectedAdvisor.id === a.id ? " active" : "");
    chip.innerHTML = "<strong>" + escapeHtml(a.name) + "</strong><span>" + escapeHtml(advisorRole(a)) + "</span>";
    chip.addEventListener("click", () => selectAdvisor(a.id));
    tabs.appendChild(chip);
  }
}

function loadMessages(advisorId) {
  try { return JSON.parse(localStorage.getItem(MSG_KEY + advisorId)) || []; } catch(e) { return []; }
}

function saveMessages(advisorId, msgs) {
  localStorage.setItem(MSG_KEY + advisorId, JSON.stringify(msgs));
}

function appendAdvisorMessage(advisorId, role, text) {
  const msgs = loadMessages(advisorId);
  msgs.push({ role, text, time: Date.now() });
  saveMessages(advisorId, msgs);
}

function renderChatLog() {
  const log = $("chatLog");
  if (!log) return;
  log.innerHTML = "";
  if (!state.selectedAdvisor) {
    log.innerHTML = '<div class="msg system">选择一位顾问开始对话。</div>';
    return;
  }
  const msgs = loadMessages(state.selectedAdvisor.id);
  if (msgs.length === 0) {
    log.innerHTML = '<div class="msg system">切换不同顾问时，每位顾问的对话历史独立保存。</div>';
    return;
  }
  for (const m of msgs) {
    const div = document.createElement("div");
    div.className = "msg " + (m.role === "player" ? "player" : m.role === "npc" ? "npc" : "system");
    div.textContent = m.text;
    log.appendChild(div);
  }
  log.scrollTop = log.scrollHeight;
}

async function sendChat(presetText) {
  if (state.loading || !state.selectedAdvisor || !state.world) return;
  const input = $("chatInput");
  const text = (presetText || input.value).trim();
  if (!text) return;
  appendAdvisorMessage(state.selectedAdvisor.id, "player", text);
  input.value = "";
  renderChatLog();
  setLoading(true, "顾问思考中...");
  try {
    const resp = await api("POST", "/api/v1/invoke", {
      world_id: state.world.id,
      task_type: "npc_dialogue",
      node_id: state.selectedAdvisor.id,
      messages: buildChatPayload(text),
    });
    appendAdvisorMessage(state.selectedAdvisor.id, "npc", resp.reply || "...");
    renderChatLog();
    setLoading(false);
  } catch(e) {
    appendAdvisorMessage(state.selectedAdvisor.id, "system", "对话失败: " + e.message);
    renderChatLog();
    setLoading(false);
  }
}

function buildChatPayload(text) {
  const msgs = loadMessages(state.selectedAdvisor.id);
  const recent = msgs.slice(-10);
  const payload = [];
  for (const m of recent) {
    payload.push({ role: m.role === "player" ? "user" : m.role === "npc" ? "assistant" : "system", content: m.text });
  }
  payload.push({ role: "user", content: text });
  return payload;
}

/* ============= Event Log ============= */
async function renderEventLog() {
  const log = $("eventLog");
  if (!log) return;
  try {
    let logs = [];
    if (state.world) {
      logs = await api("GET", "/api/v1/logs?world_id=" + encodeURIComponent(state.world.id) + "&limit=30");
    }
    if (!logs || logs.length === 0) {
      log.innerHTML = '<div class="event-item"><span>暂无日志记录。</span></div>';
      return;
    }
    log.innerHTML = "";
    for (const l of logs) {
      const item = document.createElement("div");
      item.className = "event-item";
      let t = l.created_at || "";
      try { t = new Date(t).toLocaleTimeString(); } catch(e) {}
      item.innerHTML = "<strong>" + escapeHtml(l.task_type || "?") + "</strong> <span>" + t + " · " + (l.duration_ms || 0) + "ms</span>";
      if (l.llm_model) item.innerHTML += "<br><span>模型: " + l.llm_model + " · Token: " + (l.tokens_used || 0) + "</span>";
      log.appendChild(item);
    }
  } catch(e) {
    log.innerHTML = '<div class="event-item"><span>加载日志失败</span></div>';
  }
}

/* ============= Config ============= */
function openConfig() {
  $("cfgUrl").value = cfg.url;
  $("cfgKey").value = cfg.key;
  show("configOverlay");
}

function saveConfig() {
  cfg.url = $("cfgUrl").value.trim();
  cfg.key = $("cfgKey").value.trim();
  saveCfg(cfg);
  hide("configOverlay");
  toast("已保存，重新连接中...");
  setTimeout(() => checkConnection(), 500);
}

/* ============= Helpers ============= */
function escapeHtml(s) {
  if (!s) return "";
  return String(s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
}

/* ============= Init ============= */
document.addEventListener("DOMContentLoaded", () => {
  // Connection
  checkConnection();

  // Topbar buttons
  $("initBtn").addEventListener("click", initWorld);
  $("advanceBtn").addEventListener("click", advanceTurn);
  $("dialogueBtn").addEventListener("click", () => {
    if (!state.world) { toast("请先初始化世界"); return; }
    renderAdvisorTabs();
    if (!state.selectedAdvisor && state.advisors.length > 0) {
      selectAdvisor(state.advisors[0].id);
    }
    renderChatLog();
    show("dialogueOverlay");
  });
  $("logBtn").addEventListener("click", () => {
    if (!state.world) { toast("请先初始化世界"); return; }
    renderEventLog();
    show("logOverlay");
  });
  $("configBtn").addEventListener("click", openConfig);

  // Modal closers
  $("closeDialogue").addEventListener("click", () => hide("dialogueOverlay"));
  $("closeLog").addEventListener("click", () => hide("logOverlay"));
  $("closeConfig").addEventListener("click", () => hide("configOverlay"));
  $("dialogueOverlay").addEventListener("click", (e) => { if (e.target === e.currentTarget) hide("dialogueOverlay"); });
  $("logOverlay").addEventListener("click", (e) => { if (e.target === e.currentTarget) hide("logOverlay"); });
  $("configOverlay").addEventListener("click", (e) => { if (e.target === e.currentTarget) hide("configOverlay"); });

  // Config
  $("saveConfig").addEventListener("click", saveConfig);

  // Chat
  $("sendBtn").addEventListener("click", () => sendChat());
  $("chatInput").addEventListener("keydown", (e) => { if (e.key === "Enter") sendChat(); });
  $("qaSuggest").addEventListener("click", () => {
    const inc = getCurrentIncident();
    sendChat("对于当前议题\u201c" + inc.title + "\u201d，你最建议我采取哪种方案，为什么？");
  });
  $("qaRisk").addEventListener("click", () => {
    sendChat("如果我这回合处理失误，最可能先在哪个地方出问题？请按你的身份判断。");
  });
  $("qaStatus").addEventListener("click", () => {
    sendChat("你现在看到的局势里，最关键的三条信息是什么？不要泛泛而谈。");
  });
});
