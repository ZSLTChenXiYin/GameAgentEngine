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
/* ============= API Cache Layer (E13) ============= */
var _apiCache = {};
var _apiCacheTTL = 3000; // 3 seconds for GET responses

function invalidateAPICache(pattern) {
  if (!pattern) { _apiCache = {}; return; }
  for (var key in _apiCache) {
    if (key.indexOf(pattern) >= 0) delete _apiCache[key];
  }
}

async function api(method, path, body) {
  var url = cfg.url.replace(/\/+$/, '') + path;

  // GET caching: return cached response if fresh
  if (method === 'GET' && !body) {
    var cacheKey = method + ':' + url;
    var cached = _apiCache[cacheKey];
    if (cached && Date.now() - cached.ts < _apiCacheTTL) {
      return cached.data;
    }
  }

  var opts = { method, headers: { 'X-API-Key': cfg.key, 'Content-Type': 'application/json' } };
  if (body !== undefined) opts.body = JSON.stringify(body);
  var res = await fetch(url, opts);
  var text = await res.text();
  if (!res.ok) {
    // Invalidate cache on mutation errors (stale data risk)
    invalidateAPICache(path.split('?')[0]);
    throw new Error(text);
  }
  var parsed;
  try { parsed = JSON.parse(text); } catch(e) { parsed = text; }

  // Cache GET responses
  if (method === 'GET') {
    _apiCache[method + ':' + url] = { data: parsed, ts: Date.now() };
  } else {
    // Mutations (POST/PUT/DELETE) invalidate related GET cache
    invalidateAPICache(path.split('?')[0]);
  }

  return parsed;
}

/* ============= State ============= */
let state = {
  worlds: [], nodes: [], selectedNodeId: null, selectedWorldId: null,
  selectedNodeIds: [], selectionAnchorId: null, visibleNodeIds: [],
  selectedTreePathKey: null,
  page: 'worlds', nodeDetail: null, components: [], memories: [],
  relations: [], policy: null, settings: null, plans: [], logs: [], stateComponents: [], timelines: [], snapshots: [], snapshotMeta: null, snapshotListWorldId: null,
  tasks: [], tasksStats: null,
  taskFilters: { status: '', category: '', consumer: '', diagnosticView: '', mineOnly: true },
  continuityBundle: null, continuityRequestId: '', continuityMode: '',
  autonomous: null, treeFilter: '', connected: false,
  dragNodeId: null, suppressTreeClickUntil: 0,
  leftWidth: 260, rightWidth: 300,
};
/* ============= Tree Caching Layer ============= */
var _treeCache = {
  nodeMap: null,
  childMap: null,
  flatRows: null,
  dirty: true,
  filterToken: null,
};

function invalidateTreeCache() { _treeCache.dirty = true; _treeCache.flatRows = null; }

function buildNodeMap(nodes) {
  var map = {};
  for (var i = 0; i < nodes.length; i++) map[nodes[i].id] = nodes[i];
  return map;
}

function buildChildMap(nodes) {
  var map = { _root: [] };
  for (var i = 0; i < nodes.length; i++) {
    var n = nodes[i];
    var pid = n.parent_id || '_root';
    if (!map[pid]) map[pid] = [];
    map[pid].push(n.id);
  }
  return map;
}

function ensureTreeCache() {
  if (!_treeCache.dirty && _treeCache.nodeMap) return;
  _treeCache.nodeMap = buildNodeMap(state.nodes);
  _treeCache.childMap = buildChildMap(state.nodes);
  _treeCache.flatRows = null;
  _treeCache.dirty = false;
}

function buildFlatRows() {
  ensureTreeCache();
  if (_treeCache.flatRows) return _treeCache.flatRows;

  var rows = [];
  var filter = (state.treeFilter || '').toLowerCase();
  var collapsed = state.treeCollapsed || {};

  function walk(parentId, depth, currentPathKey, ancestorIds) {
    var children = _treeCache.childMap[parentId] || [];
    for (var ci = 0; ci < children.length; ci++) {
      var nodeId = children[ci];
      var node = _treeCache.nodeMap[nodeId];
      if (!node) continue;
      if (filter && node.name.toLowerCase().indexOf(filter) < 0 && node.node_type.indexOf(filter) < 0) continue;
      if (ancestorIds.indexOf(nodeId) >= 0) continue;

      var hasChildren = _treeCache.childMap[nodeId] && _treeCache.childMap[nodeId].length > 0;
      var isExpanded = !collapsed[nodeId];
      var pathKey = currentPathKey ? currentPathKey + '|' + nodeId : nodeId;
      var paddingLeft = 12 + depth * 16;

      rows.push({ nodeId: nodeId, depth: depth, hasChildren: hasChildren, isExpanded: isExpanded, paddingLeft: paddingLeft, pathKey: pathKey });

      if (hasChildren && isExpanded) walk(nodeId, depth + 1, pathKey, ancestorIds.concat([nodeId]));
    }
  }

  walk('_root', 0, '', []);
  _treeCache.flatRows = rows;
  return rows;
}

// Override slow global helpers with cached versions
var _origGetNodeById = window.getNodeById;
function getNodeById(id) {
  ensureTreeCache();
  return _treeCache.nodeMap[id] || null;
}

function getNodeNameById(id) {
  var n = getNodeById(id);
  return n ? n.name : id;
}


const componentMetaList = Array.isArray(window.GAMEAGENT_COMPONENT_META) ? window.GAMEAGENT_COMPONENT_META : [];
const componentMetaMap = {};
componentMetaList.forEach(function(item) {
  if (item && item.type) componentMetaMap[item.type] = item;
});

function componentTypeLabel(type) {
  var meta = componentMetaMap[type] || {};
  return meta.display_name || type;
}

function componentTypeDisplay(type) {
  var label = componentTypeLabel(type);
  if (!type || label === type) return type || '';
  return type + ' (' + label + ')';
}

function componentTypeOptionsHTML() {
  var items = componentMetaList.length > 0 ? componentMetaList.slice() : [];
  if (items.length === 0) {
    items = [
      { type: 'profile' },
      { type: 'rule' },
      { type: 'timeline' },
      { type: 'action_policy' },
      { type: 'prompt_profile' },
      { type: 'lore' },
      { type: 'autonomous' },
    ];
  }
  return items.map(function(item) {
    return '<option value="' + item.type + '">' + componentTypeDisplay(item.type) + '</option>';
  }).join('');
}

/* ============= DOM ============= */
function el(tag, attrs) {
  const e = document.createElement(tag);
  if (attrs) {
    for (const k in attrs) {
      const value = attrs[k];
      if (k === 'className') e.className = value;
      else if (k.startsWith('on')) e.addEventListener(k.slice(2).toLowerCase(), value);
      else if (k === 'style' && typeof value === 'object') Object.assign(e.style, value);
      else if (k === 'innerHTML') e.innerHTML = value;
      else if (k === 'textContent') e.textContent = value;
      else if (k === 'dataset') { for (const d in value) e.dataset[d] = value[d]; }
      else if (typeof value === 'boolean') {
        e[k] = value;
        if (!value) e.removeAttribute(k);
      }
      else e.setAttribute(k, value);
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
    'Connected': '已连接',
    'Disconnected': '未连接',
    'Add': '添加',
    'Add Capability': '添加能力',
    'Add Entry': '添加条目',
    'Add Pair': '添加键值对',
    'Add Unit': '添加单位',
    'Add Unit Value': '添加单位值',
    'Add tick units before configuring calendar values.': '请先添加 Tick Units（时间刻单位），再配置日历值。',
    'Add New Parent': '添加新父节点',
    'Link External Parent': '指向外父节点',
    'Add Outgoing Relation': '添加被指向关系',
    'Add Component': '添加组件',
    'Add Memory': '添加记忆',
    'Add Relation': '添加关系',
    'Advance Tick': 'Advance Tick（推进时间刻）',
    'Advanced Ticks': 'Advanced Ticks（实际推进时间刻数）',
    'Advancing scope...': '局部推进中...',
    'Advancing tick...': '正在推进 Tick（时间刻）...',
    'Assess': '评估',
    'Assessing event impact...': '评估事件影响中...',
    'Assessment Result': '评估结果',
    'Autonomous': '自主行为',
    'Autonomous Config': '自主行为配置',
    'Autonomous Limit': '自主运行上限',
    'Autonomous Runs': '自主运行次数',
    'Autonomous config saved': '自主行为配置已保存',
    'Autonomous triggered': '已触发自主行为',
    'Blocked Actions': '禁止动作',
    'Boolean values must be true or false': '布尔值只能填写 true / false',
    'Cancel': '取消',
    'Capabilities (JSON array)': '能力列表（JSON 数组）',
    'Capabilities': '能力列表',
    'Capability Description': '能力描述',
    'Capability ID': '能力 ID',
    'Capability Mode': '能力模式',
    'Capability Schema (JSON object)': 'Capability Schema（能力结构定义，JSON 对象）',
    'Close': '关闭',
    'Component Data': '组件数据',
    'Component Data (JSON/Markdown)': '组件数据（JSON/Markdown）',
    'Component data contains invalid JSON. Fix it in raw mode before saving.': '组件数据存在非法 JSON，请先在原始模式中修正再保存。',
    'Component data must be a valid JSON object': '组件数据必须是合法的 JSON 对象',
    'Autonomous component data must be valid JSON': '自主行为组件数据必须是合法 JSON',
    'Autonomous trigger must be one of manual, world_tick_sync, scheduled': '自主行为 trigger（触发方式）必须是 manual（手动）、world_tick_sync（世界 Tick 同步）、scheduled（定时）之一',
    'Autonomous scheduled trigger requires interval_seconds > 0': 'scheduled（定时）自主行为必须设置大于 0 的 interval_seconds（间隔秒数）',
    'JSON object required for this component type': '该组件类型需要 JSON 对象',
    'Free text allowed for this component type': '该组件类型允许自由文本',
    'Structured autonomous config JSON.': '结构化的 autonomous（自主行为）配置 JSON。',
    'JSON object required; fields are flexible.': '需要 JSON 对象，但字段保持灵活',
    'Free text allowed.': '允许自由文本',
    'Component Type': '组件类型',
    'Component added': '组件已添加',
    'Component deleted': '组件已删除',
    'Component updated': '组件已更新',
    'Components': '组件',
    'Config': '配置',
    'Config saved': '配置已保存',
    'Confirm': '确认',
    'Continuity': '连续性',
    'Continuity refreshed': '连续性已刷新',
    'Continuity Diff': '连续性差异',
    'Continuity State': '连续性状态',
    'Continuity Summary': '连续性摘要',
    'Continuity Rules': '连续性规则',
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
    'Current Situation': '当前情境',
    'Current Units': '当前单位值',
    'Current Time Label': '当前时间标签',
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
    'Down': '下移',
    'Duplicate keys are not allowed': '不允许重复键名',
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
    'Edit State Component': '编辑状态组件',
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
    'Execution Mode': '执行模式',
    'Expected count': '期望数量',
    'Facts (one per line)': '事实（每行一条）',
    'Focus Scopes': '聚焦范围',
    'Failed to load logs': '加载日志失败',
    'Failed to load continuity bundle': '加载连续性聚合失败',
    'Failed to load state components': '加载状态组件失败',
    'Failed to load timelines': '加载时间线失败',
    'Failed to load node details': '加载节点详情失败',
    'Failed to load nodes': '加载节点失败',
    'Failed to load traces: ': '加载轨迹失败：',
    'Failed to load worlds': '加载世界失败',
    'Failed: ': '失败：',
    'Filter nodes...': '筛选节点...',
    'Format': '格式',
    'Full features': '完整功能',
    'GameAgentCreator': 'GameAgentCreator',
    'Game Time': '游戏时间',
    'ID': 'ID（标识）',
    'Import': '导入',
    'Import Config': '导入配置',
    'Import successful': '导入成功',
    'Incoming': '入向',
    'Incoming Relations': '入向关系',
    'Inference Params': '推理参数',
    'Initial value': '初始值',
    'Interval': '间隔',
    'Interval Seconds (scheduled)': 'Interval Seconds（scheduled 定时触发间隔秒数）',
    'Invalid JSON': '无效 JSON',
    'Issues': '问题',
    'Key': '键',
    'Key cannot be empty': '键名不能为空',
    'LLM Response': 'LLM Response（大语言模型响应）',
    'Level': '级别',
    'List World': '列表所属世界',
    'Load autonomous config first': '请先加载自主行为配置',
    'Loading snapshots...': '加载快照中...',
    'Loading...': '加载中...',
    'Lock source snapshot during restore': '恢复时锁定源快照世界',
    'Lock world during snapshot save? This prevents concurrent writes.': '保存快照时锁定世界？这会阻止并发写入。',
    'Lock world during working-copy creation? This prevents concurrent writes.': '创建工作副本时锁定世界？这会阻止并发写入。',
    'Logs': '日志',
    'Linked Logs': '关联日志',
    'Linked': '关联',
    'Linked Traces': '关联轨迹',
    'Memories': '记忆',
    'Memory Limit': '记忆数量上限',
    'Memory added': '记忆已添加',
    'Memory deleted': '记忆已删除',
    'Memory updated': '记忆已更新',
    'Malformed JSON detected. Fix raw content before using the structured editor.': '检测到非法 JSON，请先修复原始内容后再使用结构化编辑器。',
    'Metadata': '元数据',
    'Metadata (JSON object)': '元数据（JSON 对象）',
    'Model': '模型',
    'Move this node?': '移动这个节点？',
    'Multi-round': '多轮轮询',
    'Name': '名称',
    'Notes': '备注',
    'New World Name': '新世界名称',
    'No': '否',
    'No logs yet.': '暂无日志。',
    'No continuity artifacts yet.': '暂无连续性工件。',
    'No request filter applied.': '当前未应用请求过滤。',
    'No state components yet.': '暂无状态组件。',
    'No timelines yet.': '暂无时间线。',
    'No nodes. Click + to create.': '暂无节点，点击 + 创建。',
    'No incoming relations': '没有指向当前节点的关系',
    'No outgoing relations': '当前节点没有指向其他节点的关系',
    'No saved snapshots for this world yet.': '当前世界还没有存档快照。',
    'No traces yet. Run a task in Debug mode to see traces.': '暂无轨迹。请在 Debug Mode（调试模式）下运行任务后查看。',
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
    'Optional external time label': '可选的外部时间标签',
    'Open Source World': '打开源世界',
    'Open Questions': '待解问题',
    'Optional': '可选',
    'Overview': '概览',
    'Parent': '父节点',
    'Primary Parent': 'Primary Parent（主父节点）',
    'Primary parent updated': 'Primary Parent（主父节点）已更新',
    'Primary Parent is the only hierarchy field shown in the outline. External Parents and other relations stay in the relations table.': 'Primary Parent（主父节点）是大纲树里唯一展示的层级字段；External Parents（外部父级）和其他 Relations（关系）会保留在关系表中。',
    'Primary Parent is the only hierarchy field shown in the outline. Use located_at for the current environment, belongs_to/subordinate for organization or control, and external_parent only for auxiliary DAG scope.': 'Primary Parent（主父节点）是大纲树里唯一展示的层级字段。当前环境请使用 located_at（当前位置关系），组织或控制链请使用 belongs_to（稳定归属）/subordinate（隶属汇报），external_parent（辅助父级作用域）只用于辅助 DAG（有向无环图）范围。',
    'External Parents': 'External Parents（外部父级）',
    'Current Location': '当前位置',
    'Organization Links': '组织关系',
    'Paste YAML/JSON content...': '粘贴 YAML/JSON 内容...',
    'Pipeline': '管线',
    'Pipeline & Propagation': '管线与传播',
    'Pipeline Mode': 'Pipeline Mode（管线模式）',
    'present': '已存在',
    'Please select a node first': '请先选择节点',
    'Please select a world first': '请先选择世界',
    'Policy': '策略',
    'Tasks': '任务',
	'Runtime Tasks': '运行时任务',
	'Click refresh': '点击刷新',
	'Tasks refreshed': '任务已刷新',
	'Category': '分类',
	'Attempts': '尝试次数',
	'Created': '创建时间',
	'Creates a new node and rewires only parent_id': '创建一个新节点，并只重连 parent_id（主父字段）',
	'Outline drag, drop, and Add New Parent only update the node primary parent. Use Relations to edit non-tree links.': '大纲中的拖拽、放置和“添加新父节点”只会更新节点的 Primary Parent（主父节点）。非树状链接请使用 Relations（关系）编辑。',
	'Relations are stored separately from the outline tree': 'Relations（关系）与大纲树分开存储',
	'Plans': '计划',
    'Policy saved': '策略已保存',
    'Processing...': '处理中...',
    'Propagation Max Depth': '传播最大深度',
    'Propagation Mode': '传播模式',
    'Propagation Targets (node IDs)': '传播目标（节点 ID）',
    'Propagation Tags': '传播标签',
    'Propagation request submitted': '传播请求已提交',
    'Unsupported propagation mode': '不支持的传播模式',
    'Targeted propagation requires at least one target node ID': '定向传播至少需要一个目标节点 ID',
    'Tag broadcast propagation requires at least one tag': '标签广播至少需要一个标签',
    'Propagate': '传播',
    'Propagate Memory': '传播记忆',
    'Propagate this memory': '传播这条记忆',
    'Publish Up': 'Publish Up（继续向上发布）',
    'Publish Up only extends environment_scope or organization_scope beyond their scoped graph. It has no effect on targeted or manual propagation.': 'Publish Up（继续向上发布）只会让 environment_scope（环境作用域）或 organization_scope（组织作用域）在其作用域之外继续向上扩展，对 targeted（定向传播）和 manual（手动传播）不生效。',
    'Outgoing': '出向',
    'Outgoing Relations': '出向关系',
    'Pending Plans': '待审批计划',
    'Pending Threads': '待处理剧情线',
    'pending': '待审批',
    'approved': '已批准',
    'rejected': '已拒绝',
    'No pending plans for this world.': '当前世界没有待审批计划。',
    'Refresh Plans': '刷新计划',
    'Plans refreshed': '计划已刷新',
    'Plan approved': '计划已批准',
    'Plan rejected': '计划已拒绝',
    'Plan ID': '计划 ID',
    'Task Type': '任务类型',
    'Created At': '创建时间',
    'Impact': '影响等级',
    'Approve': '批准',
    'Reject': '拒绝',
    'Summary': '摘要',
    'Actions': '动作',
    'Memory Updates': '记忆更新',
    'missing': '缺失',
    'Target node IDs, comma separated': '目标节点 ID，逗号分隔',
    'Tags, comma separated': '标签，逗号分隔',
    'upward': '向上',
    'environment_scope': '环境作用域',
    'organization_scope': '组织作用域',
    'tag_broadcast': '标签广播',
    'targeted': '定向',
    'manual': '手动',
    'Reason': '原因',
    'Relation Preview': '关系预览',
    'Relation Properties': '关系属性',
    'Relation Summary': '关系摘要',
    'Relation Meaning': '关系语义',
    'Relation Validation': '关系校验',
    'Graph Context Preview': '图谱上下文预览',
    'Modeling Warnings': '建模警告',
    'Relations are stored separately from the outline tree. To change hierarchy, edit Primary Parent or drag in the outline.': 'Relations（关系）独立存储在大纲树之外。若要调整层级，请编辑 Primary Parent（主父节点）或直接在大纲树中拖拽。',
    'Relations are stored separately from the outline tree. Primary Parent remains the only stable hierarchy field. Use located_at for current environment, belongs_to/subordinate for organization or control, and external_parent only for auxiliary DAG scope.': 'Relations（关系）独立存储在大纲树之外。Primary Parent（主父节点）仍然是唯一稳定层级字段。当前环境请使用 located_at（当前位置关系），组织或控制链请使用 belongs_to（稳定归属）/subordinate（隶属汇报），external_parent（辅助父级作用域）只用于辅助 DAG（有向无环图）范围。',
    'Relations do not change the outline hierarchy unless you separately edit Primary Parent.': '除非你另外修改 Primary Parent（主父节点），否则 Relations（关系）本身不会改变大纲层级。',
    'This target is already the Primary Parent. If the node is only temporarily here, prefer keeping Primary Parent stable and using located_at for movement.': '这个目标已经是 Primary Parent（主父节点）。如果节点只是临时在这里，优先保持 Primary Parent（主父节点）稳定，并用 located_at（当前位置关系）表达移动。',
    'This node already has another located_at relation. Keep only one active current environment unless you intentionally model multiple simultaneous positions.': '这个节点已经有另一条 located_at（当前位置关系）。除非你确实要表达同时处于多个位置，否则应只保留一个当前环境。',
    'This target is already the Primary Parent. If you want to change stable hierarchy, edit Primary Parent directly. Use this relation only to add organization/control semantics.': '这个目标已经是 Primary Parent（主父节点）。如果你要修改稳定层级，请直接编辑 Primary Parent（主父节点）。这个关系只应用来补充组织或控制语义。',
    'external_parent is auxiliary DAG scope only. It is excluded from default context assembly and default propagation, so do not rely on it for current location or primary organization modeling.': 'external_parent（辅助父级作用域）仅用于辅助 DAG（有向无环图）范围。它不会进入默认上下文组装，也不会参与默认传播，因此不要用它表达当前位置或主要组织归属。',
    'This node already has another external_parent relation. Keep this relation rare and use it only when a second parent-like scope is truly required.': '这个节点已经有另一条 external_parent（辅助父级作用域）。请尽量少用这类关系，只在确实需要第二条父向作用域时再使用。',
    'Social relations are background graph edges. They are not part of the default hierarchy walk or default environment context expansion.': '社会关系属于背景图谱边，不参与默认层级遍历，也不参与默认环境上下文扩展。',
    'Follow the stable Primary Parent chain upward. This is the default mode for promoting local memory into broader identity context.': '沿稳定的 Primary Parent（主父节点）链向上传播。这是把局部记忆提升到更大身份上下文中的默认模式。',
    'Publish through the current environment chain rooted by located_at, then optionally continue upward when Publish Up is enabled.': '沿由 located_at（当前位置关系）锚定的当前环境链传播；启用 Publish Up（继续向上发布）后，可以在该环境作用域之外继续向上扩展。',
    'Publish through organization/control links such as belongs_to and subordinate, then optionally continue upward when Publish Up is enabled.': '沿 belongs_to（稳定归属）、subordinate（隶属汇报）等组织或控制链传播；启用 Publish Up（继续向上发布）后，可以在该组织作用域之外继续向上扩展。',
    'Publish to nodes matched by propagation tags. Use this for cross-cutting subscriptions rather than structural graph flow.': '传播到匹配传播标签的节点。它适合跨结构的订阅分发，而不是结构化图遍历。',
    'Publish only to the explicit target node IDs. Use this when the recipients are known ahead of time.': '只传播到显式指定的目标节点 ID。适用于接收方在传播前就已经明确的情况。',
    'Record a manual propagation request without relying on graph traversal defaults.': '记录一条手动传播请求，不依赖默认图遍历规则。',
    'Refresh Logs': '刷新日志',
    'Refresh Continuity': '刷新连续性',
    'Refresh State': '刷新状态',
    'Refresh Snapshots': '刷新快照',
    'Refresh Timelines': '刷新时间线',
    'Refresh Traces': '刷新轨迹',
    'Refresh failed': '刷新失败',
    'Relation Type': '关系类型',
    'Relation added': '关系已添加',
    'Relation deleted': '关系已删除',
    'Relation updated': '关系已更新',
    'Relations': 'Relations（关系）',
    'Replan': '重新规划',
    'Replan Result': '重规划结果',
    'Replan done': '重规划完成',
    'Replanning timeline...': '重规划时间线中...',
    'Request': '请求',
    'Request ID': 'Request ID（请求 ID）',
    'Recent Debug Traces': '最近调试轨迹',
    'Recent Changes': '近期变化',
    'Recent World Tick Bundle': 'Recent World Tick Bundle（最近世界 Tick 聚合）',
    'Recent World Tick Logs': 'Recent World Tick Logs（最近世界 Tick 日志）',
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
    'Schema': 'Schema（结构定义）',
    'Scope (node ID, optional)': '作用域（节点 ID，可选）',
    'Scope Advance': '局部推进',
    'Scope Advance requires a non-world node': '局部推进需要一个非世界类型节点',
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
    'State': '状态',
    'State Component': '状态组件',
    'State component saved': '状态组件已保存',
    'Structured Fields': '结构化字段',
    'Structured editor unavailable for this component type.': '该组件类型暂未提供结构化编辑器。',
    'Severity': '严重程度',
    'Active Arcs': '活跃剧情弧',
    'Canonical Facts': '规范事实',
    'Key Facts': '关键事实',
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
    'Sub-Task DAG': 'Sub-Task DAG（子任务有向无环图）',
    'Sub-Task Max Retries': '子任务最大重试次数',
    'Sub-Task Timeout (sec)': '子任务超时（秒）',
    'Summary: ': '摘要：',
    'Switch Language': '切换语言',
    'System Prompt': '系统提示词',
    'Tags': '标签',
    'Tags (comma separated)': '标签（逗号分隔）',
    'Target Node': '目标节点',
    'Target node search...': '搜索目标节点...',
    'Tick': 'Tick（时间刻）',
    'Tick Advance Result': 'Tick Advance Result（时间刻推进结果）',
    'Tick advanced': 'Tick（时间刻）已推进',
    'Tick Type': 'Tick Type（时间刻类型）',
    'Tick Scale Mode': 'Tick Scale Mode（时间刻尺度模式）',
    'Tick Number': 'Tick Number（时间刻编号）',
    'Tick Step': 'Tick Step（时间刻步长）',
    'Tick Units': 'Tick Units（时间刻单位列表）',
    'Tick Min Unit': 'Tick Min Unit（最小时间刻单位）',
    'World Time Label': '世界时间标签',
    'Time Scale Carry': '时间尺度进位',
    'Enable Calendar Mode': '启用日历模式',
    'Calendar Name': '历名称',
    'Calendar Units': '日历单位',
    'Unit Value Sequences': '单位值序列',
    'Unit': '单位',
    'The engine advances this many minimum tick units per inferred tick.': '每次引擎推进一个推理 Tick（时间刻）时，会按这里配置的最小 Tick（时间刻）单位数量前进。',
    'Derived from the smallest configured tick unit to match Engine constraints.': '该值由最小 Tick（时间刻）单位自动推导，以满足 Engine（引擎）的硬约束。',
    'When enabled, calendar units, carry rules, and sequences must stay consistent.': '启用后，日历单位、进位规则和单位值序列必须保持一致。',
    'Tick units must be ordered from largest to smallest. The smallest unit becomes Tick Min Unit automatically.': 'Tick Units（时间刻单位）必须按从大到小排列，最小单位会自动成为 Tick Min Unit（最小时间刻单位）。',
    'Tick units contain blank values. Fill them before saving.': 'Tick Units（时间刻单位）中存在空值，请补全后再保存。',
    'Tick units contain duplicates. Keep each unit unique before saving.': 'Tick Units（时间刻单位）中存在重复值，请保持每个单位唯一。',
    'Add at least two tick units to configure carry rules.': '至少添加两个 Tick Units（时间刻单位）后才能配置进位规则。',
    'Optional symbolic sequences become available after you add smaller tick units.': '添加更小的 Tick Units（时间刻单位）后，才需要配置可选的符号序列。',
    'Example: 年': '示例：年',
    'Example: 时辰': '示例：时辰',
    'Example: 子 | 丑 | 寅 | 卯': '示例：子 | 丑 | 寅 | 卯',
    'Example: 太阴': '示例：太阴',
    'Base': '进位基数',
    'Up': '上移',
    'Tracked Request': '跟踪请求',
    'Time': '时间',
    'This action inserts a new primary parent node and rewires only parent_id. It does not create a relation row.': '这个操作会插入一个新的 Primary Parent（主父节点），并只重连 parent_id（主父字段），不会额外创建 relation（关系）记录。',
    'Tone': '叙事语气',
    'Total Ticks': 'Total Ticks（累计时间刻数）',
    'Timeline Payload': '时间线载荷',
    'Timelines': '时间线',
    'Toggle Theme': '切换主题',
    'Tokens': 'Tokens（令牌数）',
    'Traces': '轨迹',
    'Trigger': '触发方式',
    'Type': '类型',
    'All Modes': '全部模式',
    'All Requests': '全部请求',
    'Clear Filter': '清空过滤',
    'Latest Tick': 'Latest Tick（最近时间刻）',
    'Latest Tick Summary': 'Latest Tick Summary（最近时间刻摘要）',
    'Previous Tick Summary': 'Previous Tick Summary（上一时间刻摘要）',
    'Latest Future Outline': '最近未来大纲',
    'Previous Future Outline': '上一未来大纲',
    'Latest History Summary': '最近剧情历史摘要',
    'Previous History Summary': '上一剧情历史摘要',
    'Added Facts': '新增事实',
    'Removed Facts': '移除事实',
    'Stable Facts': '稳定事实',
    'Current Canonical Facts': '当前规范事实',
    'Previous Story Facts': '上一剧情事实',
    'No previous tick to compare.': '暂无可对比的上一 Tick（时间刻）。',
    'Select a request to focus linked logs and traces.': '选择一个请求以聚焦其关联日志和轨迹。',
    'Self': '自环',
    'Self Relations': '自环关系',
    'Source node search...': '搜索源节点...',
    'Unknown error': '未知错误',
    'Valid': '有效',
    'Validate': '校验',
    'Validating snapshot...': '校验快照中...',
    'Validation Issues': '校验问题',
    'Value': '值',
    'Value JSON': '值 JSON（结构化值）',
    'Value must be a valid number': '值必须是合法数字',
    'Weight': '权重',
    'Working copy created': '工作副本已创建',
    'World': '世界',
    'World Name': '世界名称',
    'World Outline': '世界节点树',
    'World Policy': '世界策略',
    'World Settings': '世界设置',
    'World Time Settings': 'World Time Settings（世界时间设置）',
    'Future Outline': '未来大纲',
    'requested_ticks is required for flexible mode and must stay 1 in fixed mode.': '在 flexible（弹性）模式下必须填写 requested_ticks（请求推进刻数），而 fixed（固定）模式下它必须保持为 1。',
    'Requested Ticks': 'Requested Ticks（请求推进时间刻数）',
    'Requested ticks must be greater than 0': 'Requested Ticks（请求推进时间刻数）必须大于 0',
    'Fixed tick scale mode only allows requested_ticks = 1': '固定 Tick Scale Mode（时间刻尺度模式）下只允许 requested_ticks = 1',
    'Structured world tick continuity state.': '结构化的 World Tick（世界时间刻）连续性状态。',
    'Structured current world state for tick continuity. Optional fields must keep their expected string / string-array / object shapes.': '用于 Tick（时间刻）连续性的结构化世界状态，可选字段必须保持字符串、字符串数组或对象等既定结构。',
    'Structured current narrative state and unresolved threads. Optional fields must keep their expected string / string-array / object shapes.': '用于描述当前叙事状态和未解决剧情线的结构化数据，可选字段必须保持字符串、字符串数组或对象等既定结构。',
    'Structured rolling history of recent story beats. entries must be an array of structured history objects.': '用于记录近期剧情节拍的结构化历史数据，entries（条目）必须是结构化历史对象数组。',
    'Structured tick policy and continuity constraints. Optional fields must keep their expected string-array / object shapes.': '用于定义 Tick（时间刻）策略和连续性约束的结构化数据，可选字段必须保持字符串数组或对象等既定结构。',
    'Structured current world time state for engine-managed tick progression.': '用于 Engine（引擎）管理 Tick（时间刻）推进的结构化世界时间状态。',
    'Structured snapshot payload for state rollups and checkpoints.': '用于状态汇总和检查点的结构化快照载荷。',
    'Last Error': '最近错误',
    'Last Run At': '最近运行时间',
    'Last Tick Type': 'Last Tick Type（最近时间刻类型）',
    'Last Advanced Ticks': 'Last Advanced Ticks（最近推进时间刻数）',
    'World created': '世界已创建',
    'World updated': '世界已更新',
    'World:': '世界：',
    'Worlds': '世界',
    'Parent node created': '父节点已创建',
    'This node already points to the selected external parent': '该节点已经指向所选外父节点',
    'Target node is already the primary parent': '目标节点已经是 Primary Parent（主父节点）',
    'Cannot link a node to itself': '不能将节点指向自己',
    'This relation already exists': '这条关系已经存在',
    'This node already has another located_at relation. Update the existing environment edge instead of stacking multiple active locations.': '这个节点已经有另一条 located_at（当前位置关系）。请更新现有环境边，而不是叠加多个同时生效的位置。',
    'Properties (JSON or Markdown)': '属性（JSON 或 Markdown 格式）',
    'Optional notes, tags, role metadata...': '可填写备注、标签、角色元数据等...',
    'Select a relation type': '请选择关系类型',
    'Select a source node': '请选择源节点',
    'Yes': '是',
    'critical': '严重',
    'diplomatic_shift, natural_disaster...': 'diplomatic_shift（外交变化）, natural_disaster（自然灾害）...',
    'fork_world': '工作副本',
    'high': '高',
    'invalid': '无效快照',
    'low': '低',
    'medium': '中',
    'none': '无',
    'restore_snapshot': '恢复副本',
    'restored_copy': '已恢复副本',
    'save_snapshot': '存档快照',
    'tag1,tag2...': 'tag1,tag2（标签示例）...',
    'world_state.summary must be a string': 'world_state.summary 必须是字符串。',
    'world_state.key_facts must be an array of strings': 'world_state.key_facts 必须是字符串数组。',
    'world_state.canonical_facts must be an array of strings': 'world_state.canonical_facts 必须是字符串数组。',
    'world_state.open_questions must be an array of strings': 'world_state.open_questions 必须是字符串数组。',
    'world_state.active_arcs must be an array of strings': 'world_state.active_arcs 必须是字符串数组。',
    'world_state.metadata must be an object': 'world_state.metadata 必须是对象。',
    'story_state.current_situation must be a string': 'story_state.current_situation 必须是字符串。',
    'story_state.recent_changes must be an array of strings': 'story_state.recent_changes 必须是字符串数组。',
    'story_state.pending_threads must be an array of strings': 'story_state.pending_threads 必须是字符串数组。',
    'story_state.tone must be a string': 'story_state.tone 必须是字符串。',
    'story_state.metadata must be an object': 'story_state.metadata 必须是对象。',
    'story_history.entries must be an array of objects': 'story_history.entries 必须是对象数组。',
    'story_history.entries[].tick_number must be a non-negative integer': 'story_history.entries[].tick_number 必须是非负整数。',
    'story_history.entries[].summary must be a string': 'story_history.entries[].summary 必须是字符串。',
    'story_history.entries[].facts must be an array of strings': 'story_history.entries[].facts 必须是字符串数组。',
    'story_history.entries[].game_time must be a string': 'story_history.entries[].game_time 必须是字符串。',
    'story_history.metadata must be an object': 'story_history.metadata 必须是对象。',
    'tick_policy.continuity_rules must be an array of strings': 'tick_policy.continuity_rules 必须是字符串数组。',
    'tick_policy.focus_scopes must be an array of strings': 'tick_policy.focus_scopes 必须是字符串数组。',
    'tick_policy.banned_resets must be an array of strings': 'tick_policy.banned_resets 必须是字符串数组。',
    'tick_policy.metadata must be an object': 'tick_policy.metadata 必须是对象。',
    'tick_scale_mode must be fixed or flexible': 'tick_scale_mode 必须是 fixed（固定）或 flexible（弹性）。',
    'tick_min_unit must not be empty': 'tick_min_unit（最小时间刻单位）不能为空。',
    'tick_step must be greater than 0': 'tick_step（时间刻步长）必须大于 0。',
    'tick_units must contain at least one unit': 'tick_units（时间刻单位）至少要包含一个单位。',
    'tick_units must not contain empty values': 'tick_units（时间刻单位）不能包含空值。',
    'tick_units must not contain duplicate values': 'tick_units（时间刻单位）不能包含重复值。',
    'tick_min_unit must match the smallest configured tick unit': 'tick_min_unit（最小时间刻单位）必须与当前配置中的最小 Tick Units（时间刻单位）一致。',
    'time_scale_carry must define exactly one adjacent rule per unit gap': 'time_scale_carry 必须为每一层相邻单位间隔定义且仅定义一条进位规则。',
    'time_scale_carry entries must define from, to, and base > 0': 'time_scale_carry 的每一项都必须定义 from、to，且 base 必须大于 0。',
    'time_calendar.calendar_name must not be empty when calendar mode is enabled': '启用日历模式后，time_calendar.calendar_name 不能为空。',
    'time_calendar.units must match tick_units exactly when calendar mode is enabled': '启用日历模式后，time_calendar.units 必须与 tick_units（时间刻单位）完全一致。',
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

Object.assign(i18n.zh, {
  'No tasks found.': '暂无任务。',
  'Analysis Rounds': '分析轮数',
  'Context Depth': '上下文深度',
  'Auto Apply': '自动应用',
  'Review Threshold': '审核阈值',
  'Enable Propagation Machine': '启用传播机制',
});

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

