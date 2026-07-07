/*
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
  selectedNodeIds: [], selectionAnchorId: null, visibleNodeIds: [],
  selectedTreePathKey: null,
  page: 'worlds', nodeDetail: null, components: [], memories: [],
  relations: [], policy: null, settings: null, logs: [], snapshots: [], snapshotMeta: null, snapshotListWorldId: null,
  autonomous: null, treeFilter: '', connected: false,
  dragNodeId: null, suppressTreeClickUntil: 0,
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
    '-- Select World --': '-- 选择世界 --',
    'API Key': 'API 密钥',
    'Add': '添加',
    'Add New Parent': '添加新父节点',
    'Link External Parent': '指向外父节点',
    'Add Component': '添加组件',
    'Add Memory': '添加记忆',
    'Add Relation': '添加关系',
    'Advance Tick': '推进 Tick',
    'Advancing scope...': '局部推进中...',
    'Advancing tick...': '推进 Tick 中...',
    'Assess': '评估',
    'Assessing event impact...': '评估事件影响中...',
    'Assessment Result': '评估结果',
    'Autonomous': '自主行为',
    'Autonomous Config': '自主行为配置',
    'Autonomous config saved': '自主行为配置已保存',
    'Autonomous triggered': '已触发自主行为',
    'Blocked Actions': '禁止动作',
    'Cancel': '取消',
    'Capabilities (JSON array)': '能力列表（JSON 数组）',
    'Close': '关闭',
    'Component Data': '组件数据',
    'Component Data (JSON/Markdown)': '组件数据（JSON/Markdown）',
    'Component Type': '组件类型',
    'Component added': '组件已添加',
    'Component deleted': '组件已删除',
    'Component updated': '组件已更新',
    'Components': '组件',
    'Config': '配置',
    'Config saved': '配置已保存',
    'Confirm': '确认',
    'Content': '内容',
    'Counts': '计数',
    'Create': '创建',
    'Create Child': '创建子节点',
    'Create Parent Node': '创建父节点',
    'Create Node': '创建节点',
    'Create Working Copy': '创建工作副本',
    'Create World': '创建世界',
    'Creating working copy...': '创建工作副本中...',
    'Current World Snapshot Metadata': '当前世界快照元数据',
    'Delete': '删除',
    'Delete this component?': '确定删除这个组件？',
    'Delete this memory?': '确定删除这条记忆？',
    'Delete this node?': '确定删除这个节点？',
    'Delete this relation?': '确定删除这条关系？',
    'Delete this snapshot world and its metadata?': '删除这个快照世界及其元数据？',
    'Deleting snapshot...': '删除快照中...',
    'Describe the event...': '描述事件内容...',
    'Drop here to move to root': '拖到这里可移动到根级',
    'Description': '描述',
    'Dry-run': '仅验证',
    'Duration': '耗时',
    'Copy': '复制',
    'Copy Node': '复制节点',
    'Copy subtree': '复制子树',
    'Edit': '编辑',
    'Edit World': '编辑世界',
    'Edit Component': '编辑组件',
    'Edit Config': '编辑配置',
    'Edit Memory': '编辑记忆',
    'Edit Node': '编辑节点',
    'Edit Relation': '编辑关系',
    'Edit Relation Endpoints': '编辑关系端点',
    'Enabled': '启用',
    'Engine': '引擎',
    'Enter a node name': '请输入节点名称',
    'Enter a world name': '请输入世界名称',
    'Enter component data': '请输入组件数据',
    'Enter component data...': '输入组件数据...',
    'Enter content': '请输入内容',
    'Enter event type and description': '请输入事件类型和描述',
    'Enter memory content': '请输入记忆内容',
    'Enter memory content...': '输入记忆内容...',
    'Enter name...': '输入名称...',
    'Enter world name...': '输入世界名称...',
    'Error': '错误',
    'Event Impact': '事件影响',
    'Event Impact Assessment': '事件影响评估',
    'Event Type': '事件类型',
    'Event assessed': '事件已评估',
    'Execution Control': '执行控制',
    'Failed to load logs': '加载日志失败',
    'Failed to load node details': '加载节点详情失败',
    'Failed to load nodes': '加载节点失败',
    'Failed to load traces: ': '加载轨迹失败：',
    'Failed to load worlds': '加载世界失败',
    'Failed: ': '失败：',
    'Filter nodes...': '筛选节点...',
    'Format': '格式',
    'Full features': '完整功能',
    'GameAgentCreator': 'GameAgentCreator',
    'ID': 'ID',
    'Import': '导入',
    'Import Config': '导入配置',
    'Import successful': '导入成功',
    'Inference Params': '推理参数',
    'Interval': '间隔',
    'Interval Seconds (scheduled)': '间隔秒数（scheduled）',
    'Invalid JSON': '无效 JSON',
    'Issues': '问题',
    'LLM Response': 'LLM 响应',
    'Level': '级别',
    'List World': '列表所属世界',
    'Load autonomous config first': '请先加载自主行为配置',
    'Loading snapshots...': '加载快照中...',
    'Loading...': '加载中...',
    'Lock source snapshot during restore': '恢复时锁定源快照世界',
    'Lock world during snapshot save? This prevents concurrent writes.': '保存快照时锁定世界？这会阻止并发写入。',
    'Lock world during working-copy creation? This prevents concurrent writes.': '创建工作副本时锁定世界？这会阻止并发写入。',
    'Logs': '日志',
    'Memories': '记忆',
    'Memory Limit': '记忆数量上限',
    'Memory added': '记忆已添加',
    'Memory deleted': '记忆已删除',
    'Memory updated': '记忆已更新',
    'Model': '模型',
    'Move this node?': '移动这个节点？',
    'Multi-round': '多轮轮询',
    'Name': '名称',
    'Notes': '备注',
    'New World Name': '新世界名称',
    'No': '否',
    'No logs yet.': '暂无日志。',
    'No nodes. Click + to create.': '暂无节点，点击 + 创建。',
    'No saved snapshots for this world yet.': '当前世界还没有存档快照。',
    'No traces yet. Run a task in Debug mode to see traces.': '暂无轨迹。请在 Debug 模式下运行任务后查看。',
    'No validation issues.': '没有校验问题。',
    'Node': '节点',
    'Node Name': '节点名称',
    'Node Type': '节点类型',
    'Node copied': '节点已复制',
    'External parent linked': '外父节点已关联',
    'Node created': '节点已创建',
    'Node deleted': '节点已删除',
    'Node moved': '节点已移动',
    'Node updated': '节点已更新',
    'Nodes': '节点',
    'Open Source World': '打开源世界',
    'Optional': '可选',
    'Overview': '概览',
    'Parent': '父节点',
    'Primary Parent': '主父节点',
    'External Parents': '外父节点',
    'Paste YAML/JSON content...': '粘贴 YAML/JSON 内容...',
    'Pipeline': '管线',
    'Pipeline & Propagation': '管线与传播',
    'Pipeline Mode': '管线模式',
    'Please select a node first': '请先选择节点',
    'Please select a world first': '请先选择世界',
    'Policy': '策略',
    'Policy saved': '策略已保存',
    'Processing...': '处理中...',
    'Propagation Max Depth': '传播最大深度',
    'Reason': '原因',
    'Relation Preview': '关系预览',
    'Relation Properties': '关系属性',
    'Relation Summary': '关系摘要',
    'Relation Meaning': '关系语义',
    'Refresh Logs': '刷新日志',
    'Refresh Snapshots': '刷新快照',
    'Refresh Traces': '刷新轨迹',
    'Refresh failed': '刷新失败',
    'Relation Type': '关系类型',
    'Relation added': '关系已添加',
    'Relation deleted': '关系已删除',
    'Relation updated': '关系已更新',
    'Relations': '关系',
    'Replan': '重新规划',
    'Replan Result': '重规划结果',
    'Replan done': '重规划完成',
    'Replanning timeline...': '重规划时间线中...',
    'Request': '请求',
    'Reset World': '重置世界',
    'Response': '响应',
    'Restorable': '可恢复',
    'Restore': '恢复',
    'Restore Snapshot': '恢复快照',
    'Restoring snapshot...': '恢复快照中...',
    'Round': '轮',
    'Rounds': '轮次',
    'Run Autonomous': '运行自主行为',
    'Running autonomous...': '运行自主行为中...',
    'Safe Actions': '安全动作',
    'Save': '保存',
    'Save Policy': '保存策略',
    'Save Settings': '保存设置',
    'Save Snapshot': '保存快照',
    'Saved Snapshots': '存档快照',
    'Saved Snapshots of Source World': '源世界的存档快照',
    'Saving snapshot...': '保存快照中...',
    'Schema': 'Schema',
    'Scope (node ID, optional)': '作用域（节点 ID，可选）',
    'Scope Advance': '局部推进',
    'Scope Advance requires a non-world node': '局部推进需要一个非 world 类型节点',
    'Scope advanced': '局部推进完成',
    'Select a node as scope': '请选择一个节点作为作用域',
    'Select a node to inspect, or right-click in the outline.': '选择一个节点查看详情，或在大纲树中右键操作。',
    'Select a target node': '请选择目标节点',
    'Select an external parent node': '请选择外父节点',
    'Select a world': '请选择世界',
    'Select a world first': '请先选择世界',
    'Select a world first.': '请先选择世界。',
    'Select a world to begin editing.': '请选择一个世界开始编辑。',
    'Search Nodes': '搜索节点',
    'Select source and target': '请选择源节点和目标节点',
    'Selected:': '已选择：',
    'nodes selected': '个节点已选中',
    'Server Config': '服务器配置',
    'Server URL': '服务器地址',
    'Settings': '设置',
    'Settings saved': '设置已保存',
    'Severity': '严重程度',
    'Showing all save snapshots created from the currently selected world.': '正在显示当前选中世界创建的全部存档快照。',
    'Showing all save snapshots created from the source world of the current snapshot.': '正在显示当前快照所属源世界的全部存档快照。',
    'Single round': '单轮直推',
    'Snapshot': '快照',
    'Snapshot Name': '快照名称',
    'Snapshot Validation': '快照校验',
    'Snapshot deleted': '快照已删除',
    'Snapshot restored': '快照已恢复',
    'Snapshot saved': '快照已保存',
    'Snapshots': '快照',
    'Snapshots refreshed': '快照已刷新',
    'Source Node': '源节点',
    'Source World': '源世界',
    'Status': '状态',
    'Sub-Task DAG': '子任务 DAG',
    'Sub-Task Max Retries': '子任务最大重试次数',
    'Sub-Task Timeout (sec)': '子任务超时（秒）',
    'Summary: ': '摘要：',
    'Switch Language': '切换语言',
    'System Prompt': '系统提示词',
    'Tags': '标签',
    'Tags (comma separated)': '标签（逗号分隔）',
    'Target Node': '目标节点',
    'Target node search...': '搜索目标节点...',
    'Tick Advance Result': '推进 Tick 结果',
    'Tick advanced': 'Tick 已推进',
    'Time': '时间',
    'Toggle Theme': '切换主题',
    'Tokens': 'Token',
    'Traces': '轨迹',
    'Trigger': '触发方式',
    'Type': '类型',
    'Source node search...': '搜索源节点...',
    'Unknown error': '未知错误',
    'Valid': '有效',
    'Validate': '校验',
    'Validating snapshot...': '校验快照中...',
    'Validation Issues': '校验问题',
    'Weight': '权重',
    'Working copy created': '工作副本已创建',
    'World': '世界',
    'World Name': '世界名称',
    'World Outline': '世界节点树',
    'World Policy': '世界策略',
    'World Settings': '世界设置',
    'World created': '世界已创建',
    'World updated': '世界已更新',
    'World:': '世界：',
    'Worlds': '世界',
    'Parent node created': '父节点已创建',
    'This node already points to the selected external parent': '该节点已经指向所选外父节点',
    'Target node is already the primary parent': '目标节点已经是主父节点',
    'Cannot link a node to itself': '不能将节点指向自己',
    'This relation already exists': '这条关系已经存在',
    'Properties (JSON or Markdown)': '属性（JSON 或 Markdown）',
    'Optional notes, tags, role metadata...': '可填写备注、标签、角色元数据等...',
    'Select a relation type': '请选择关系类型',
    'Select a source node': '请选择源节点',
    'Yes': '是',
    'critical': '严重',
    'diplomatic_shift, natural_disaster...': 'diplomatic_shift, natural_disaster...',
    'fork_world': '工作副本',
    'high': '高',
    'invalid': '无效快照',
    'low': '低',
    'medium': '中',
    'none': '无',
    'restore_snapshot': '恢复副本',
    'restored_copy': '已恢复副本',
    'save_snapshot': '存档快照',
    'tag1,tag2...': 'tag1,tag2...',
    'unknown': '未知',
    'valid': '有效快照',
    'working_copy': '工作副本',
  },
  en: {},
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

