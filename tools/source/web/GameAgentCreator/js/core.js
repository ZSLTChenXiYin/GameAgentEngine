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
  relations: [], policy: null, settings: null, plans: [], logs: [], stateComponents: [], timelines: [], snapshots: [], snapshotMeta: null, snapshotListWorldId: null,
  continuityBundle: null, continuityRequestId: '', continuityMode: '',
  autonomous: null, treeFilter: '', connected: false,
  dragNodeId: null, suppressTreeClickUntil: 0,
  leftWidth: 260, rightWidth: 300,
};

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
    'Add Capability': '添加能力',
    'Add Entry': '添加条目',
    'Add Pair': '添加键值对',
    'Add Unit': '添加单位',
    'Add Unit Value': '添加单位值',
    'Add New Parent': '添加新父节点',
    'Link External Parent': '指向外父节点',
    'Add Outgoing Relation': '添加被指向关系',
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
    'Boolean values must be true or false': '布尔值只能填写 true 或 false',
    'Cancel': '取消',
    'Capabilities (JSON array)': '能力列表（JSON 数组）',
    'Capabilities': '能力列表',
    'Capability Description': '能力描述',
    'Capability ID': '能力 ID',
    'Capability Mode': '能力模式',
    'Capability Schema (JSON object)': '能力 Schema（JSON 对象）',
    'Close': '关闭',
    'Component Data': '组件数据',
    'Component Data (JSON/Markdown)': '组件数据（JSON/Markdown）',
    'Component data contains invalid JSON. Fix it in raw mode before saving.': '组件数据存在非法 JSON，请先在原始模式中修正再保存。',
    'Component data must be a valid JSON object': '组件数据必须是合法的 JSON 对象',
    'Autonomous component data must be valid JSON': '自主行为组件数据必须是合法 JSON',
    'Autonomous trigger must be one of manual, world_tick_sync, scheduled': '自主行为 trigger 必须是 manual、world_tick_sync、scheduled 之一',
    'Autonomous scheduled trigger requires interval_seconds > 0': 'scheduled 自主行为必须设置大于 0 的 interval_seconds',
    'JSON object required for this component type': '该组件类型需要 JSON 对象',
    'Free text allowed for this component type': '该组件类型允许自由文本',
    'Structured autonomous config JSON.': '结构化的 autonomous 配置 JSON',
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
    'ID': 'ID',
    'Import': '导入',
    'Import Config': '导入配置',
    'Import successful': '导入成功',
    'Incoming': '入向',
    'Incoming Relations': '入向关系',
    'Inference Params': '推理参数',
    'Initial value': '初始值',
    'Interval': '间隔',
    'Interval Seconds (scheduled)': '间隔秒数（scheduled）',
    'Invalid JSON': '无效 JSON',
    'Issues': '问题',
    'Key': '键',
    'Key cannot be empty': '键名不能为空',
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
    'Linked Logs': '关联日志',
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
    'Open Questions': '待解问题',
    'Optional': '可选',
    'Overview': '概览',
    'Parent': '父节点',
    'Primary Parent': '主父节点',
    'External Parents': '外父节点',
    'Paste YAML/JSON content...': '粘贴 YAML/JSON 内容...',
    'Pipeline': '管线',
    'Pipeline & Propagation': '管线与传播',
    'Pipeline Mode': '管线模式',
    'present': '已存在',
    'Please select a node first': '请先选择节点',
    'Please select a world first': '请先选择世界',
    'Policy': '策略',
    'Plans': '计划',
    'Policy saved': '策略已保存',
    'Processing...': '处理中...',
    'Propagation Max Depth': '传播最大深度',
    'Propagation Mode': '传播模式',
    'Propagation Targets (node IDs)': '传播目标（节点 ID）',
    'Propagation Tags': '传播标签',
    'Propagation request submitted': '传播请求已提交',
    'Propagate': '传播',
    'Propagate Memory': '传播记忆',
    'Propagate this memory': '传播这条记忆',
    'Publish Up': '继续向上发布',
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
    'tag_broadcast': '标签广播',
    'targeted': '定向',
    'manual': '手动',
    'Reason': '原因',
    'Relation Preview': '关系预览',
    'Relation Properties': '关系属性',
    'Relation Summary': '关系摘要',
    'Relation Meaning': '关系语义',
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
    'Relations': '关系',
    'Replan': '重新规划',
    'Replan Result': '重规划结果',
    'Replan done': '重规划完成',
    'Replanning timeline...': '重规划时间线中...',
    'Request': '请求',
    'Request ID': '请求 ID',
    'Recent Debug Traces': '最近调试轨迹',
    'Recent Changes': '近期变化',
    'Recent World Tick Bundle': '最近 World Tick 聚合',
    'Recent World Tick Logs': '最近 World Tick 日志',
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
    'Tick Scale Mode': 'Tick 尺度模式',
    'Tick Number': 'Tick 编号',
    'Tick Step': 'Tick 步长',
    'Tick Units': 'Tick 单位列表',
    'Tick Min Unit': 'Tick 最小单位',
    'Time Scale Carry': '时间尺度进位',
    'Enable Calendar Mode': '启用日历模式',
    'Calendar Name': '历名称',
    'Calendar Units': '日历单位',
    'Unit Value Sequences': '单位值序列',
    'Unit': '单位',
    'The engine advances this many minimum tick units per inferred tick.': '每次引擎推进一个推理 Tick 时，会按这里配置的最小 Tick 单位数量前进。',
    'Derived from the smallest configured tick unit to match Engine constraints.': '该值由最小 Tick 单位自动推导，以满足 Engine 的硬约束。',
    'When enabled, calendar units, carry rules, and sequences must stay consistent.': '启用后，日历单位、进位规则和单位值序列必须保持一致。',
    'Tick units must be ordered from largest to smallest. The smallest unit becomes Tick Min Unit automatically.': 'Tick 单位必须按从大到小排列，最小单位会自动成为 Tick 最小单位。',
    'Tick units contain blank values. Fill them before saving.': 'Tick 单位中存在空值，请补全后再保存。',
    'Tick units contain duplicates. Keep each unit unique before saving.': 'Tick 单位中存在重复值，请保持每个单位唯一。',
    'Add at least two tick units to configure carry rules.': '至少添加两个 Tick 单位后才能配置进位规则。',
    'Optional symbolic sequences become available after you add smaller tick units.': '添加更小的 Tick 单位后，才需要配置可选的符号序列。',
    'Example: 年': '示例：年',
    'Example: 时辰': '示例：时辰',
    'Example: 子 | 丑 | 寅 | 卯': '示例：子 | 丑 | 寅 | 卯',
    'Base': '进位基数',
    'Up': '上移',
    'Tracked Request': '跟踪请求',
    'Time': '时间',
    'Tone': '叙事语气',
    'Total Ticks': '累计 Tick 数',
    'Timeline Payload': '时间线载荷',
    'Timelines': '时间线',
    'Toggle Theme': '切换主题',
    'Tokens': 'Token',
    'Traces': '轨迹',
    'Trigger': '触发方式',
    'Type': '类型',
    'All Modes': '全部模式',
    'All Requests': '全部请求',
    'Clear Filter': '清空过滤',
    'Latest Tick': '最近 Tick',
    'Latest Tick Summary': '最近 Tick 摘要',
    'Previous Tick Summary': '上一 Tick 摘要',
    'Latest Future Outline': '最近未来大纲',
    'Previous Future Outline': '上一未来大纲',
    'Latest History Summary': '最近剧情历史摘要',
    'Previous History Summary': '上一剧情历史摘要',
    'Added Facts': '新增事实',
    'Removed Facts': '移除事实',
    'Stable Facts': '稳定事实',
    'Current Canonical Facts': '当前规范事实',
    'Previous Story Facts': '上一剧情事实',
    'No previous tick to compare.': '暂无可对比的上一 Tick。',
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
    'Value JSON': '值 JSON',
    'Value must be a valid number': '值必须是合法数字',
    'Weight': '权重',
    'Working copy created': '工作副本已创建',
    'World': '世界',
    'World Name': '世界名称',
    'World Outline': '世界节点树',
    'World Policy': '世界策略',
    'World Settings': '世界设置',
    'World Time Settings': '世界时间设置',
    'Future Outline': '未来大纲',
    'Structured world tick continuity state.': '结构化的 world tick 连续性状态。',
    'Structured current world state for tick continuity. Optional fields must keep their expected string / string-array / object shapes.': '用于 Tick 连续性的结构化世界状态，可选字段必须保持字符串、字符串数组或对象等既定结构。',
    'Structured current narrative state and unresolved threads. Optional fields must keep their expected string / string-array / object shapes.': '用于描述当前叙事状态和未解决剧情线的结构化数据，可选字段必须保持字符串、字符串数组或对象等既定结构。',
    'Structured rolling history of recent story beats. entries must be an array of structured history objects.': '用于记录近期剧情节拍的结构化历史数据，entries 必须是结构化历史对象数组。',
    'Structured tick policy and continuity constraints. Optional fields must keep their expected string-array / object shapes.': '用于定义 Tick 策略和连续性约束的结构化数据，可选字段必须保持字符串数组或对象等既定结构。',
    'Structured current world time state for engine-managed tick progression.': '用于引擎管理 Tick 推进的结构化世界时间状态。',
    'Structured snapshot payload for state rollups and checkpoints.': '用于状态汇总和检查点的结构化快照载荷。',
    'Last Error': '最近错误',
    'Last Run At': '最近运行时间',
    'Last Tick Type': '最近 Tick 类型',
    'Last Advanced Ticks': '最近推进 Tick 数',
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

