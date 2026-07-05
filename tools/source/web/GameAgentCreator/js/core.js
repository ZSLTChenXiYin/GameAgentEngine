/*
  ]);
*
 * GameAgentCreator - UE Style Creator
 * Pure SPA. URL/API Key stored in localStorage.
 */

/* ============= Config ============= */
const CFG_KEY = 'gameagent_creator_cfg';
function loadCfg() {
  try { return JSON.parse(localStorage.getItem(CFG_KEY)) || {}; } catch(e) { return {}; }
}
function saveCfg(cfg) { localStorage.setItem(CFG_KEY, JSON.stringify(cfg)); }

let cfg = loadCfg();
if (!cfg.url) cfg.url = 'http://127.0.0.1:8080';
if (!cfg.key) cfg.key = 'dev-key';
saveCfg(cfg);

/* ============= API ============= */
async function api(method, path, body) {
  const url = cfg.url.replace(/\/+$/, '') + path;
  const opts = { method, headers: { 'X-API-Key': cfg.key, 'Content-Type': 'application/json' } };
  if (body !== undefined) opts.body = JSON.stringify(body);
  const res = await fetch(url, opts);
  const text = await res.text();
  if (!res.ok) throw new Error(text);
  try { return JSON.parse(text); } catch(e) { return text; }
}

/* ============= State ============= */
let state = {
  worlds: [], nodes: [], selectedNodeId: null, selectedWorldId: null,
  page: 'worlds', nodeDetail: null, components: [], memories: [],
  relations: [], policy: null, settings: null, logs: [],
  autonomous: null, treeFilter: '', connected: false,
  leftWidth: 260, rightWidth: 300,
};

/* ============= DOM ============= */
function el(tag, attrs) {
  const e = document.createElement(tag);
  if (attrs) {
    for (const k in attrs) {
      if (k === 'className') e.className = attrs[k];
      else if (k.startsWith('on')) e.addEventListener(k.slice(2).toLowerCase(), attrs[k]);
      else if (k === 'style' && typeof attrs[k] === 'object') Object.assign(e.style, attrs[k]);
      else if (k === 'innerHTML') e.innerHTML = attrs[k];
      else if (k === 'textContent') e.textContent = attrs[k];
      else if (k === 'dataset') { for (const d in attrs[k]) e.dataset[d] = attrs[k][d]; }
      else e.setAttribute(k, attrs[k]);
    }
  }
  return e;
}

function txt(v) { return document.createTextNode(String(v)); }

function ce(tag, attr, children) {
  const e = el(tag, attr);
  if (children) { for (const c of children) if (c != null) e.appendChild(typeof c === 'string' ? txt(c) : c); }
  return e;
}



/* ============= i18n ============= */
const LANG_KEY = 'gameagent_creator_lang';
let lang = localStorage.getItem(LANG_KEY) || 'zh';
const THEME_KEY = 'gameagent_creator_theme';
let theme = localStorage.getItem(THEME_KEY) || 'dark';

const i18n = {
  zh: {
    GameAgentCreator: 'GameAgentCreator',
    '-- Select World --': '-- 选择世界 --',
    'World Outline': '世界节点树',
    'No nodes. Click + to create.': '暂无节点，点击 + 创建',
    'Selected:': '选中:',
    'World:': '世界:',
    'Create World': '创建世界',
    'Import Config': '导入配置',
    'Advance Tick': '推进Tick',
    'Run Autonomous': '运行自主行为',
    'Event Impact': '事件影响',
    'Scope Advance': '局部推进',
    'Advancing Tick...': '推进Tick中...',
    'Running autonomous...': '运行自主行为中...',
    'Assessing event impact...': '事件影响评估中...',
    'Advancing scope...': '局部推进中...',
    'Replanning timeline...': '重新规划中...',
    'Processing...': '处理中...',
    'Replan': '重新规划',
    'Select a node to inspect, or right-click in the outline.': '选择节点查看详情，或右键节点树操作',
    'Select a world to begin editing.': '请先选择世界',
    'Overview': '概览',
    'ID': 'ID',
    'Name': '名称',
    'Type': '类型',
    'Parent': '父节点',
    'World Policy': '世界策略',
    'Select a world first.': '请先选择世界',
    'Blocked Actions': '阻止的动作',
    'Safe Actions': '安全的动作',
    'Save Policy': '保存策略',
    'World Settings': '世界设置',
    'Inference Params': '推理参数',
    'Execution Control': '执行控制',
    'Save Settings': '保存设置',
    'Refresh Logs': '刷新日志',
    'Time': '时间',
    'Model': '模型',
    'Tokens': 'Token',
    'Duration': '耗时',
    'No logs yet.': '暂无日志',
    'World Name': '世界名称',
    'Create': '创建',
    'Node Name': '节点名称',
    'Node Type': '节点类型',
    'Save': '保存',
    'Component Type': '组件类型',
    'Component Data (JSON/Markdown)': '组件数据(JSON/Markdown)',
    'Add': '添加',
    'Content': '内容',
    'Level': '层级',
    'Tags (comma separated)': '标签(逗号分隔)',
    'Target Node': '目标节点',
    'Relation Type': '关系类型',
    'Weight': '权重',
    'Format': '格式',
    'Dry-run': '仅校验',
    'Reset World': '重置世界',
    'Import': '导入',
    'Server URL': '服务地址',
    'API Key': 'API密钥',
    'Component Data': '组件数据',
    'Tags': '标签',
    'Source Node': '源节点',
    'Event Type': '事件类型',
    'Scope (node ID, optional)': '范围(节点ID, 可选)',
    'Description': '描述',
    'Severity': '严重程度',
    'Assess': '评估',
    'Enabled': '启用',
    'Trigger': '触发类型',
    'Interval Seconds (scheduled)': '间隔秒数(scheduled)',
    'Capabilities (JSON array)': '能力列表(JSON数组)',
    'Worlds': '世界',
    'Nodes': '节点',
    'Components': '组件',
    'Memories': '记忆',
    'Relations': '关系',
    'Autonomous': '自主行为',
    'Edit Config': '编辑配置',
    'Stats': '统计',
    'History': '历史',
    'Logs': '日志',
    'Create Child Node': '创建子节点',
    'Create Child': '创建子节点',
    'Add Component': '添加组件',
    'Add Relation': '添加关系',
    'Edit Node': '编辑节点',
    'Delete Node': '删除节点',
    'Create Root Node': '创建根节点',
    'Expand All': '展开全部',
    'Collapse All': '收起全部',
    'Connected': '已连接',
    'Disconnected': '断开',
    'Connecting...': '连接中...',
    'Node': '节点',
    'Settings': '设置',
    'World': '世界',
    'Search...': '搜索...',
    'Yes': '是',
    'No': '否',
    'Filter nodes...': '搜索节点...',
    'One action per line': '每行一个动作',
    'Impact: ': '影响: ',
    'Summary: ': '摘要: ',
    'Close': '关闭',
    'Connect': '连接',
    'Disconnect': '断开',
    '中/EN': '中/EN',

    'Cancel': '取消',
    'Optional': '可选',
    'Enter world name...': '输入世界名称...',
    'Enter name...': '输入名称...',
    'Enter component data...': '输入组件数据...',
    'Enter memory content...': '输入记忆内容...',
    'tag1,tag2...': '标签1,标签2...',
    'diplomatic_shift, natural_disaster...': 'diplomatic_shift, natural_disaster...',
    'Describe the event...': '描述事件...',
    'Assessment Result': '评估结果',
    'Replan Result': '重规划结果',
    'Policy': '策略',

    'Please select a world first': '请先选择世界',
    'Please select a node first': '请先选择节点',
    'Memory Limit': '记忆数量上限',
    'Analysis Rounds': '分析轮次上限',
    'Context Depth': '上下文追溯深度',
    'Auto Apply': '自动应用',
    'Review Threshold': '审核门槛',
    'none': '无',
    'low': '低',
    'medium': '中',
    'high': '高',
    'critical': '严重',
    'Paste YAML/JSON content...': '粘贴YAML/JSON内容...',

    'Enter a world name': '请输入世界名称',
    'World created': '世界已创建',
    'Enter content': '请输入内容',
    'Node created': '节点已创建',
    'Node deleted': '节点已删除',
    'Move this node?': '移动此节点？',
    'Component added': '组件已添加',
    'Component updated': '组件已更新',
    'Component deleted': '组件已删除',
    'Memory added': '记忆已添加',
    'Memory deleted': '记忆已删除',
    'Relation added': '关系已添加',
    'Relation deleted': '关系已删除',
    'Autonomous triggered': '自主行为已触发',
    'Event assessed': '事件已评估',
    'Replan done': '重新规划完成',
    'Policy saved': '策略已保存',
    'Settings saved': '设置已保存',
    'Config saved': '配置已保存',
    'Imported': '已导入',
    'Import Config': '导入配置',
    'Paste YAML/JSON content...': '粘贴YAML/JSON内容...',
    'Edit': '编辑',
    'Delete': '删除',
    'Select a node': '请选择节点',
    'Scope Advance requires a non-world node': '局部推进需要一个非世界类型的节点',
    'Pipeline & Propagation': '管线与传播',
    'Sub-Task DAG': '子任务 DAG',
    'Pipeline Mode': '管线模式',
    'Propagation Max Depth': '传播最大层数',
    'Enable Propagation Machine': '启用传播状态机',
    'Sub-Task Max Retries': '子任务最大重试次数',
    'Sub-Task Timeout (sec)': '子任务超时(秒)',
    'Single round': '单轮直通',
    'Multi-round': '多轮轮询',
    'Full features': '完整功能',
    'Clone World': '复制世界',
    'Cloning world...': '复制世界中...',
    'World cloned': '世界已复制',
    'Lock world during clone? This prevents concurrent writes.': '复制时锁定世界？将阻止并发写入',
    'Edit Node': '编辑节点',
    'Autonomous Config': '自主行为配置',
    'Create World': '创建世界',
    'Create Node': '创建节点',
    'Add Component': '添加组件',
    'Add Memory': '添加记忆',
    'Add Relation': '添加关系',
    'Server Config': '服务器配置',
    'Edit Component': '编辑组件',
    'Edit Memory': '编辑记忆',
    'Edit Relation': '编辑关系',
    'Event Impact Assessment': '事件影响评估',
    'Edit Node': '编辑节点',
    'Autonomous Config': '自主行为配置',  },

    'Delete this component?': '确定删除此组件？',
    'Delete this memory?': '确定删除此记忆？',
    'Delete this relation?': '确定删除此关系？',
    'Delete this node?': '确定删除此节点？',
    'Enter a node name': '请输入节点名称',
    'Enter component data': '请输入组件数据',
    'Enter memory content': '请输入记忆内容',
    'Select a target node': '请选择目标节点',
    'Enter event type and description': '请输入事件类型和描述',
    'Select source and target': '请选择源节点和目标节点',
    'Failed to load worlds': '加载世界列表失败',
    'Failed to load nodes': '加载节点失败',
    'Failed to load node details': '加载节点详情失败',
    'Failed to load policy': '加载策略失败',
    'Refresh failed': '刷新失败',
    'Node updated': '节点已更新',
    'Node moved': '节点已移动',
    'Memory updated': '记忆已更新',
    'Relation updated': '关系已更新',
    'Import successful': '导入成功',
    'Event Impact Result': '事件影响结果',
    'Replan Result': '重规划结果',
    'Scope Advance Result': '局部推进结果',
    'Tick Advance Result': '推进Tick结果',
    'Autonomous Run Result': '自主行为运行结果',
    'Connected': '已连接',
    'Disconnected': '断开',
    'Connecting...': '连接中...',
    'Yes': '是',
    'No': '否',
    'Edit': '编辑',
    'Delete': '删除',
    'Cancel': '取消',
    'Search...': '搜索...',
    'Filter nodes...': '搜索节点...',
    'One action per line': '每行一个动作',
    'Enter name...': '输入名称...',
    'Enter component data...': '输入组件数据...',
    'Enter memory content...': '输入记忆内容...',
    'tag1,tag2...': '标签1,标签2...',
    'Scope Advance requires a non-world node': '局部推进需要一个非世界类型的节点',

  en: {
  }
};

// Fill en with defaults
for (var k in i18n.zh) {
  if (!i18n.en[k]) i18n.en[k] = k;
}

function tr(key) { return i18n[lang][key] || key; }


function requireWorldGuard() {
  if (!state.selectedWorldId) {
    toast(tr('Please select a world first'), 'error');
    return false;
  }
  return true;
}
function requireNodeGuard() {
  if (!state.selectedNodeId) {
    toast(tr('Please select a node first'), 'error');
    return false;
  }
  return true;
}
function requireBothGuard() {
  if (!state.selectedWorldId) { toast(tr('Please select a world first'), 'error'); return false; }
  if (!state.selectedNodeId) { toast(tr('Please select a node first'), 'error'); return false; }
  return true;
}

function applyTheme(t) {
  theme = t;
  document.documentElement.classList.toggle('light', t === 'light');
  document.documentElement.classList.toggle('dark', t === 'dark');
  localStorage.setItem(THEME_KEY, t);
}

function toggleTheme() {
  applyTheme(theme === 'dark' ? 'light' : 'dark');
}

function toggleLang() {
  lang = (lang === 'zh') ? 'en' : 'zh';
  localStorage.setItem(LANG_KEY, lang);
  init();
}

function ttxt(text) { return document.createTextNode(tr(text)); }

/* ============= Toast ============= */
function toast(msg, type) {
  const t = document.getElementById('toast');
  t.textContent = msg; t.className = type || ''; t.classList.add('show');
  clearTimeout(t._tm);
  t._tm = setTimeout(function() { t.classList.remove('show'); }, 2500);
}

/* ============= Modal ============= */
function openModal(title, bodyEl, footerEl) {
  document.getElementById('modalTitle').textContent = title;
  const mb = document.getElementById('modalBody'); mb.innerHTML = '';
  if (bodyEl) mb.appendChild(typeof bodyEl === 'string' ? el('div', {innerHTML: bodyEl}) : bodyEl);
  const mf = document.getElementById('modalFooter'); mf.innerHTML = '';
  if (footerEl) mf.appendChild(typeof footerEl === 'string' ? el('div', {innerHTML: footerEl}) : footerEl);
  document.getElementById('modalOverlay').classList.remove('hidden');
  document.getElementById('modalContainer').classList.remove('hidden');
}

function showLoading(msg) {
  document.getElementById('loadingOverlay').classList.remove('hidden');
  document.getElementById('loadingText').textContent = msg || tr('Processing...');
}

function hideLoading() {
  document.getElementById('loadingOverlay').classList.add('hidden');
}

function closeModal() {
  document.getElementById('modalOverlay').classList.add('hidden');
  document.getElementById('modalContainer').classList.add('hidden');
}

/* ============= Context Menu ============= */
function showContextMenu(items, x, y) {
  const cm = document.getElementById('contextMenu');
  cm.innerHTML = ''; cm.style.left = x + 'px'; cm.style.top = y + 'px';
  cm.classList.remove('hidden');
  for (const item of items) {
    const btn = ce('button', { className: 'm-item' + (item.danger ? ' danger' : '') }, [txt(item.label)]);
    if (item.tip) btn.appendChild(ce('span', { className: 'tip' }, [txt(item.tip)]));
    btn.addEventListener('click', function() { cm.classList.add('hidden'); if (item.onClick) item.onClick(); });
    cm.appendChild(btn);
  }
}

function hideContextMenu() { document.getElementById('contextMenu').classList.add('hidden'); }
document.addEventListener('click', hideContextMenu);
document.addEventListener('contextmenu', function(e) { e.preventDefault(); });
