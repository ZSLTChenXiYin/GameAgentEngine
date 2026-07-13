/* ============= Data Loaders ============= */
function apiErrorMessage(err) {
  if (!err) return tr('Unknown error');
  var message = err.message || String(err);
  try {
    var parsed = JSON.parse(message);
    if (parsed && parsed.error) {
      return parsed.error;
    }
  } catch (e) {}
  return message;
}

function normalizeMessagePrefix(prefixKey) {
  return tr(prefixKey).replace(/[：:\s]+$/, '');
}

function formatMessageWithDetail(prefixKey, detail) {
  return normalizeMessagePrefix(prefixKey) + '：' + detail;
}

function formatApiError(prefixKey, err) {
  return formatMessageWithDetail(prefixKey, apiErrorMessage(err));
}

function showApiError(prefixKey, err) {
  toast(formatApiError(prefixKey, err), 'error');
}

function quoteValue(value) {
  return '"' + String(value || '') + '"';
}

function formatWorldTimeValidationError(code, params) {
  switch (code) {
    case 'carry_rule_order':
      return 'time_scale_carry[' + params.index + '] 必须为 ' + params.expectedFrom + ' -> ' + params.expectedTo + '。';
    case 'sequence_unit_empty':
      return 'unit_value_sequences[' + params.index + '].unit 不能为空。';
    case 'sequence_unit_missing':
      return 'unit_value_sequences[' + params.index + '].unit ' + quoteValue(params.unit) + ' 必须存在于 tick_units（时间刻单位）中。';
    case 'sequence_unit_largest':
      return 'unit_value_sequences[' + params.index + '].unit ' + quoteValue(params.unit) + ' 不能是最大的 Tick Units（时间刻单位）。';
    case 'sequence_values_empty':
      return 'unit_value_sequences[' + params.index + '].values 不能为空。';
    case 'sequence_missing_carry_rule':
      return 'unit_value_sequences[' + params.index + '].unit ' + quoteValue(params.unit) + ' 需要一条对应的 time_scale_carry 规则。';
    case 'sequence_values_count':
      return 'unit_value_sequences[' + params.index + '].values 中针对 ' + quoteValue(params.unit) + ' 的值数量必须严格等于 ' + params.expectedCount + '。';
    case 'calendar_unit_empty':
      return 'time_calendar.units[' + params.index + '].unit 不能为空。';
    case 'calendar_unit_mismatch':
      return 'time_calendar.units[' + params.index + '].unit 必须是 ' + quoteValue(params.expectedUnit) + '。';
    case 'calendar_value_empty':
      return 'time_calendar.units[' + params.index + '].value 不能为空。';
    case 'calendar_value_range':
      return 'time_calendar.units[' + params.index + '].value 中针对 ' + quoteValue(params.unit) + ' 的数值必须介于 0 和 ' + params.maxValue + ' 之间。';
    case 'calendar_unit_requires_sequences':
      return 'time_calendar.units[' + params.index + '].unit ' + quoteValue(params.unit) + ' 需要对应的 unit_value_sequences（单位值序列）配置。';
    case 'calendar_value_missing_from_sequences':
      return 'time_calendar.units[' + params.index + '].value ' + quoteValue(params.value) + ' 必须存在于 ' + quoteValue(params.unit) + ' 的 unit_value_sequences（单位值序列）中。';
    case 'calendar_smallest_unit_mismatch':
      return 'time_calendar 中最小单位必须与 tick_min_unit（最小时间刻单位）一致。';
    default:
      return tr('Unknown error');
  }
}

async function checkHealth() {
  try {
    await api('GET', '/health');
    state.connected = true;
    document.getElementById('connStatus').innerHTML = '<span class="status-dot on"></span> ' + tr('Connected');
  } catch(e) {
    state.connected = false;
    document.getElementById('connStatus').innerHTML = '<span class="status-dot off"></span> ' + tr('Disconnected');
  }
}

async function loadWorlds() {
  try {
    state.worlds = await api('GET', '/api/v1/worlds');
    var selectedExists = !!(state.selectedWorldId && state.worlds.some(function(w) { return w.id === state.selectedWorldId; }));
    if (!selectedExists) {
      state.selectedWorldId = null;
      state.nodes = [];
state.tasks = [];
      state.relations = [];
      state.selectedNodeId = null;
      state.selectedNodeIds = [];
      state.selectionAnchorId = null;
      state.selectedTreePathKey = null;
      state.nodeDetail = null;
      state.snapshots = [];
      state.snapshotMeta = null;
      state.snapshotListWorldId = null;
      state.logs = [];
      state.stateComponents = [];
      state.timelines = [];
      state.continuityBundle = null;
      state.continuityRequestId = '';
      state.continuityMode = '';
      state.settings = null;
      state.policy = null;
      state.plans = [];
    }
    const sel = document.getElementById('worldSelector');
    if (sel) {
      const cur = selectedExists ? state.selectedWorldId : '';
      sel.innerHTML = '';
      sel.appendChild(el('option', { value: '', textContent: '-- Select World --' }));
      for (const w of state.worlds) sel.appendChild(el('option', { value: w.id, textContent: w.name }));
      if (cur) sel.value = cur;
    }
    if (!state.selectedWorldId && state.worlds.length > 0) {
      state.selectedWorldId = state.worlds[0].id;
      selectWorld(state.selectedWorldId);
      return;
    }
    renderTree(); renderCurrent();
  } catch(e) {
    state.worlds = [];
    showApiError('Failed to load worlds', e);
  }
}

async function selectWorld(worldId) {
  state.selectedWorldId = worldId;
  var worldSelector = document.getElementById('worldSelector');
  if (worldSelector) worldSelector.value = worldId || '';
  if (worldId) {
    try { state.nodes = await api('GET', '/api/v1/nodes?world_id=' + encodeURIComponent(worldId)); }
    catch(e) { state.nodes = []; showApiError('Failed to load nodes', e); }
    try { state.relations = await api('GET', '/api/v1/relations?world_id=' + encodeURIComponent(worldId)); }
    catch(e) { state.relations = []; }
    state.selectedNodeIds = [];
    state.selectionAnchorId = null;
    state.selectedNodeId = null;
    state.selectedTreePathKey = null;
    state.nodeDetail = null;
    state.logs = []; state.stateComponents = []; state.timelines = []; state.continuityBundle = null; state.continuityRequestId = ''; state.continuityMode = '';
    loadPolicy(); loadSettings(); loadPlans(true); loadSnapshots(); loadStateComponents(); loadTimelines(); loadContinuityOverview();
    if (state.page === 'continuity') loadContinuityOverview();
    if (state.page === 'logs') loadLogs();
    if (state.page === 'state') loadStateComponents();
    if (state.page === 'timelines') loadTimelines();
  } else {
    state.nodes = []; state.relations = []; state.selectedNodeId = null; state.selectedNodeIds = []; state.selectionAnchorId = null; state.selectedTreePathKey = null; state.nodeDetail = null; state.snapshots = []; state.snapshotMeta = null;
    state.snapshotListWorldId = null; state.logs = []; state.stateComponents = []; state.timelines = []; state.continuityBundle = null; state.continuityRequestId = ''; state.continuityMode = ''; state.settings = null; state.policy = null; state.plans = [];
  }
  renderTree(); renderCurrent();
}

function getVisibleTreeOrder() {
  if (state.visibleNodeIds && state.visibleNodeIds.length > 0) return state.visibleNodeIds.slice();
  return state.nodes.map(function(n) { return n.id; });
}

function uniqueNodeIds(ids) {
  var seen = {};
  var out = [];
  for (var i = 0; i < ids.length; i++) {
    var id = ids[i];
    if (!id || seen[id]) continue;
    seen[id] = true;
    out.push(id);
  }
  return out;
}

function relationTypeOptionsHTML() {
  return '<option value="belongs_to">belongs_to</option>' +
    '<option value="ally">ally</option>' +
    '<option value="enemy">enemy</option>' +
    '<option value="subordinate">subordinate</option>' +
    '<option value="kinship">kinship</option>' +
    '<option value="located_at">located_at</option>' +
    '<option value="external_parent">external_parent</option>';
}

function componentValidationMode(componentType) {
  var meta = componentMetaMap[componentType];
  if (!meta || !meta.validation_mode) return 'free';
  return meta.validation_mode;
}

function validateComponentEditorData(componentType, data) {
  var trimmed = (data || '').trim();
  if (!trimmed) return tr('Enter component data');
  var mode = componentValidationMode(componentType);
  if (mode === 'free') return '';
  var meta = componentMetaMap[componentType] || {};
  var parsed = null;
  try {
    parsed = JSON.parse(trimmed);
  } catch (e) {
    return componentType === 'autonomous' ? tr('Autonomous component data must be valid JSON') : tr('Component data must be a valid JSON object');
  }
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    return tr('Component data must be a valid JSON object');
  }
  if (mode === 'strong' && componentType === 'autonomous') {
    var trigger = parsed.trigger || 'manual';
    var triggerEnum = meta.enum_fields && meta.enum_fields.trigger ? meta.enum_fields.trigger : ['manual', 'world_tick_sync', 'scheduled'];
    if (triggerEnum.indexOf(trigger) < 0) {
      return tr('Autonomous trigger must be one of manual, world_tick_sync, scheduled');
    }
    if (trigger === 'scheduled') {
      var interval = Number(parsed.interval_seconds || 0);
      if (!Number.isFinite(interval) || interval <= 0) {
        return tr('Autonomous scheduled trigger requires interval_seconds > 0');
      }
    }
  }
  if (mode === 'strong' && componentType === 'world_state') {
    if (parsed.summary !== undefined && typeof parsed.summary !== 'string') return tr('world_state.summary must be a string');
    if (parsed.key_facts !== undefined && !Array.isArray(parsed.key_facts)) return tr('world_state.key_facts must be an array of strings');
    if (parsed.canonical_facts !== undefined && !Array.isArray(parsed.canonical_facts)) return tr('world_state.canonical_facts must be an array of strings');
    if (parsed.open_questions !== undefined && !Array.isArray(parsed.open_questions)) return tr('world_state.open_questions must be an array of strings');
    if (parsed.active_arcs !== undefined && !Array.isArray(parsed.active_arcs)) return tr('world_state.active_arcs must be an array of strings');
    if (parsed.metadata !== undefined && (!parsed.metadata || Array.isArray(parsed.metadata) || typeof parsed.metadata !== 'object')) return tr('world_state.metadata must be an object');
  }
  if (mode === 'strong' && componentType === 'story_state') {
    if (parsed.current_situation !== undefined && typeof parsed.current_situation !== 'string') return tr('story_state.current_situation must be a string');
    if (parsed.recent_changes !== undefined && !Array.isArray(parsed.recent_changes)) return tr('story_state.recent_changes must be an array of strings');
    if (parsed.pending_threads !== undefined && !Array.isArray(parsed.pending_threads)) return tr('story_state.pending_threads must be an array of strings');
    if (parsed.tone !== undefined && typeof parsed.tone !== 'string') return tr('story_state.tone must be a string');
    if (parsed.metadata !== undefined && (!parsed.metadata || Array.isArray(parsed.metadata) || typeof parsed.metadata !== 'object')) return tr('story_state.metadata must be an object');
  }
  if (mode === 'strong' && componentType === 'story_history') {
    if (parsed.entries !== undefined) {
      if (!Array.isArray(parsed.entries)) return tr('story_history.entries must be an array of objects');
      for (var i = 0; i < parsed.entries.length; i++) {
        var entry = parsed.entries[i];
        if (!entry || Array.isArray(entry) || typeof entry !== 'object') return tr('story_history.entries must be an array of objects');
        if (entry.tick_number !== undefined && (!Number.isInteger(entry.tick_number) || entry.tick_number < 0)) return tr('story_history.entries[].tick_number must be a non-negative integer');
        if (entry.summary !== undefined && typeof entry.summary !== 'string') return tr('story_history.entries[].summary must be a string');
        if (entry.facts !== undefined && !Array.isArray(entry.facts)) return tr('story_history.entries[].facts must be an array of strings');
        if (entry.game_time !== undefined && typeof entry.game_time !== 'string') return tr('story_history.entries[].game_time must be a string');
      }
    }
    if (parsed.metadata !== undefined && (!parsed.metadata || Array.isArray(parsed.metadata) || typeof parsed.metadata !== 'object')) return tr('story_history.metadata must be an object');
  }
  if (mode === 'strong' && componentType === 'tick_policy') {
    if (parsed.continuity_rules !== undefined && !Array.isArray(parsed.continuity_rules)) return tr('tick_policy.continuity_rules must be an array of strings');
    if (parsed.focus_scopes !== undefined && !Array.isArray(parsed.focus_scopes)) return tr('tick_policy.focus_scopes must be an array of strings');
    if (parsed.banned_resets !== undefined && !Array.isArray(parsed.banned_resets)) return tr('tick_policy.banned_resets must be an array of strings');
    if (parsed.metadata !== undefined && (!parsed.metadata || Array.isArray(parsed.metadata) || typeof parsed.metadata !== 'object')) return tr('tick_policy.metadata must be an object');
  }
  return '';
}

function updateComponentEditorHint(typeElementId, hintElementId) {
  var typeEl = document.getElementById(typeElementId);
  var hintEl = document.getElementById(hintElementId);
  if (!typeEl || !hintEl) return;
  var mode = componentValidationMode(typeEl.value);
  var meta = componentMetaMap[typeEl.value] || {};
  if (meta.help_text) {
    hintEl.textContent = tr(meta.help_text);
    return;
  }
  hintEl.textContent = mode === 'free' ? tr('Free text allowed for this component type') : tr('JSON object required for this component type');
}

function splitEditorLines(value) {
  return String(value || '').split('\n').map(function(item) { return item.trim(); }).filter(Boolean);
}

function safeParseJSONText(value) {
  try {
    return { value: JSON.parse(value), error: '' };
  } catch (e) {
    return { value: null, error: e && e.message ? e.message : String(e) };
  }
}

function defaultStructuredComponentPayload(componentType) {
  switch (componentType) {
    case 'autonomous':
      return { enabled: false, trigger: 'manual', interval_seconds: 0, capabilities: [], last_run_at: '', last_error: '' };
    case 'world_state':
      return { summary: '', key_facts: [], canonical_facts: [], open_questions: [], active_arcs: [], metadata: {} };
    case 'story_state':
      return { current_situation: '', recent_changes: [], pending_threads: [], tone: '', metadata: {} };
    case 'story_history':
      return { entries: [], metadata: {} };
    case 'tick_policy':
      return { continuity_rules: [], focus_scopes: [], banned_resets: [], metadata: {} };
    case 'world_time_state':
      return {
        tick_scale_mode: 'fixed',
        tick_min_unit: '',
        tick_step: 1,
        tick_units: [],
        calendar_name: '',
        current_units: [],
        current_time_label: '',
        total_ticks: 0,
        last_tick_number: 0,
        last_tick_type: '',
        last_advanced_ticks: 0,
        metadata: {},
      };
    default:
      return {};
  }
}

function getComponentEditorContext(typeElementId, rawElementId, hostElementId, rawLabelId, hintElementId) {
  return {
    typeEl: document.getElementById(typeElementId),
    rawEl: document.getElementById(rawElementId),
    hostEl: document.getElementById(hostElementId),
    rawLabelEl: document.getElementById(rawLabelId),
    hintEl: document.getElementById(hintElementId),
  };
}

function setComponentEditorHintText(ctx, componentType, fallbackText) {
  if (!ctx || !ctx.hintEl) return;
  var meta = componentMetaMap[componentType] || {};
  if (fallbackText) {
    ctx.hintEl.textContent = fallbackText;
    return;
  }
  if (meta.help_text) {
    ctx.hintEl.textContent = tr(meta.help_text);
    return;
  }
  var mode = componentValidationMode(componentType);
  ctx.hintEl.textContent = mode === 'free' ? tr('Free text allowed for this component type') : tr('JSON object required for this component type');
}

function normalizeJSONObject(value) {
  if (!value || Array.isArray(value) || typeof value !== 'object') return {};
  return value;
}

function metadataJSONText(value) {
  return JSON.stringify(normalizeJSONObject(value), null, 2);
}

function ensureComponentTextareaVisibility(ctx, visible) {
  if (!ctx || !ctx.rawEl || !ctx.rawLabelEl) return;
  ctx.rawEl.style.display = visible ? '' : 'none';
  ctx.rawLabelEl.style.display = visible ? '' : 'none';
}

function createComponentField(labelText, control) {
  return ce('div', { className: 'component-editor-field' }, [
    ce('label', {}, [ttxt(labelText)]),
    control,
  ]);
}

function createComponentSection(title, bodyChildren) {
  return ce('div', { className: 'component-editor-section' }, [
    ce('div', { className: 'component-editor-section-title' }, [ttxt(title)]),
    ce('div', { className: 'component-editor-section-body' }, bodyChildren),
  ]);
}

function appendKeyValueRow(listEl, entry) {
  var row = ce('div', { className: 'component-kv-row' }, [
    el('input', { type: 'text', className: 'component-kv-key', placeholder: tr('Key'), value: entry && entry.key ? entry.key : '' }),
    el('select', { className: 'component-kv-type', innerHTML: '<option value="string">string</option><option value="number">number</option><option value="boolean">boolean</option><option value="json">json</option>' }),
    el('textarea', { className: 'component-kv-value', rows: 2, placeholder: tr('Value'), textContent: entry && entry.value != null ? String(entry.value) : '' }),
    ce('button', { type: 'button', className: 'danger component-kv-delete' }, [ttxt('Delete')]),
  ]);
  listEl.appendChild(row);
  var typeEl = row.querySelector('.component-kv-type');
  var valueEl = row.querySelector('.component-kv-value');
  typeEl.value = entry && entry.type ? entry.type : 'string';
  function applyMode() {
    valueEl.rows = typeEl.value === 'json' ? 4 : 2;
    valueEl.placeholder = tr(typeEl.value === 'json' ? 'Value JSON' : 'Value');
  }
  typeEl.addEventListener('change', applyMode);
  row.querySelector('.component-kv-delete').addEventListener('click', function() { row.remove(); });
  applyMode();
}

function objectToKeyValueEntries(payload) {
  var obj = normalizeJSONObject(payload);
  return Object.keys(obj).map(function(key) {
    var value = obj[key];
    var type = 'string';
    var textValue = '';
    if (typeof value === 'number') {
      type = 'number';
      textValue = String(value);
    } else if (typeof value === 'boolean') {
      type = 'boolean';
      textValue = value ? 'true' : 'false';
    } else if (value && typeof value === 'object') {
      type = 'json';
      textValue = JSON.stringify(value, null, 2);
    } else {
      textValue = value == null ? '' : String(value);
    }
    return { key: key, type: type, value: textValue };
  });
}

function renderWeakComponentEditor(ctx, payload) {
  var listEl = ce('div', { className: 'component-kv-list' }, []);
  var entries = objectToKeyValueEntries(payload);
  if (entries.length === 0) entries.push({ key: '', type: 'string', value: '' });
  entries.forEach(function(entry) { appendKeyValueRow(listEl, entry); });
  ctx.hostEl.appendChild(createComponentSection('Structured Fields', [
    listEl,
    ce('div', { className: 'policy-actions component-editor-actions' }, [
      ce('button', { type: 'button', id: ctx.hostEl.id + '_addKv' }, [ttxt('Add Pair')]),
    ]),
  ]));
  document.getElementById(ctx.hostEl.id + '_addKv').addEventListener('click', function() {
    appendKeyValueRow(listEl, { key: '', type: 'string', value: '' });
  });
  ctx.hostEl.dataset.editorMode = 'weak';
}

function appendCapabilityRow(listEl, entry) {
  var row = ce('div', { className: 'component-collection-row' }, [
    createComponentField('Capability ID', el('input', { type: 'text', className: 'component-cap-id', value: entry && entry.id ? entry.id : '' })),
    createComponentField('Capability Mode', el('input', { type: 'text', className: 'component-cap-mode', value: entry && entry.mode ? entry.mode : '' })),
    createComponentField('Capability Description', el('textarea', { className: 'component-cap-desc', rows: 2, textContent: entry && entry.description ? entry.description : '' })),
    createComponentField('Capability Schema (JSON object)', el('textarea', { className: 'component-cap-schema', rows: 4, textContent: entry && entry.schema ? JSON.stringify(entry.schema, null, 2) : '{}' })),
    ce('div', { className: 'policy-actions component-editor-actions' }, [ce('button', { type: 'button', className: 'danger component-collection-delete' }, [ttxt('Delete')])]),
  ]);
  listEl.appendChild(row);
  row.querySelector('.component-collection-delete').addEventListener('click', function() { row.remove(); });
}

function appendStoryHistoryEntryRow(listEl, entry) {
  var row = ce('div', { className: 'component-collection-row' }, [
    createComponentField('Tick Number', el('input', { type: 'number', min: '0', className: 'component-entry-tick', value: entry && Number.isFinite(entry.tick_number) ? String(entry.tick_number) : '0' })),
    createComponentField('Summary', el('textarea', { className: 'component-entry-summary', rows: 2, textContent: entry && entry.summary ? entry.summary : '' })),
    createComponentField('Facts (one per line)', el('textarea', { className: 'component-entry-facts', rows: 4, textContent: entry && Array.isArray(entry.facts) ? entry.facts.join('\n') : '' })),
    createComponentField('Game Time', el('input', { type: 'text', className: 'component-entry-game-time', value: entry && entry.game_time ? entry.game_time : '' })),
    ce('div', { className: 'policy-actions component-editor-actions' }, [ce('button', { type: 'button', className: 'danger component-collection-delete' }, [ttxt('Delete')])]),
  ]);
  listEl.appendChild(row);
  row.querySelector('.component-collection-delete').addEventListener('click', function() { row.remove(); });
}

function appendCurrentUnitRow(listEl, entry) {
  var row = ce('div', { className: 'component-inline-row' }, [
    el('input', { type: 'text', className: 'component-unit-name', placeholder: tr('Unit'), value: entry && entry.unit ? entry.unit : '' }),
    el('input', { type: 'text', className: 'component-unit-value', placeholder: tr('Value'), value: entry && entry.value ? entry.value : '' }),
    ce('button', { type: 'button', className: 'danger component-collection-delete' }, [ttxt('Delete')]),
  ]);
  listEl.appendChild(row);
  row.querySelector('.component-collection-delete').addEventListener('click', function() { row.remove(); });
}

function renderStrongComponentEditor(ctx, componentType, payload) {
  var value = normalizeJSONObject(payload);
  var sections = [];
  switch (componentType) {
    case 'autonomous': {
      var capList = ce('div', { className: 'component-collection-list', id: ctx.hostEl.id + '_caps' }, []);
      (Array.isArray(value.capabilities) ? value.capabilities : []).forEach(function(item) { appendCapabilityRow(capList, item || {}); });
      sections.push(createComponentSection('Structured Fields', [
        createComponentField('Enabled', el('input', { type: 'checkbox', id: ctx.hostEl.id + '_enabled', checked: !!value.enabled })),
        createComponentField('Trigger', el('select', { id: ctx.hostEl.id + '_trigger', innerHTML: '<option value="manual">manual</option><option value="world_tick_sync">world_tick_sync</option><option value="scheduled">scheduled</option>' })),
        createComponentField('Interval Seconds (scheduled)', el('input', { type: 'number', min: '0', id: ctx.hostEl.id + '_interval', value: Number.isFinite(value.interval_seconds) ? String(value.interval_seconds) : '0' })),
        createComponentField('Last Run At', el('input', { type: 'text', id: ctx.hostEl.id + '_lastRunAt', value: value.last_run_at || '' })),
        createComponentField('Last Error', el('textarea', { id: ctx.hostEl.id + '_lastError', rows: 2, textContent: value.last_error || '' })),
      ]));
      sections.push(createComponentSection('Capabilities', [
        capList,
        ce('div', { className: 'policy-actions component-editor-actions' }, [ce('button', { type: 'button', id: ctx.hostEl.id + '_addCap' }, [ttxt('Add Capability')])]),
      ]));
      ctx.hostEl.appendChild(ce('div', { className: 'component-editor-grid' }, sections));
      document.getElementById(ctx.hostEl.id + '_trigger').value = value.trigger || 'manual';
      document.getElementById(ctx.hostEl.id + '_addCap').addEventListener('click', function() { appendCapabilityRow(capList, {}); });
      break;
    }
    case 'world_state':
      sections.push(createComponentSection('Structured Fields', [
        createComponentField('Summary', el('textarea', { id: ctx.hostEl.id + '_summary', rows: 3, textContent: value.summary || '' })),
        createComponentField('Key Facts', el('textarea', { id: ctx.hostEl.id + '_keyFacts', rows: 4, textContent: Array.isArray(value.key_facts) ? value.key_facts.join('\n') : '' })),
        createComponentField('Canonical Facts', el('textarea', { id: ctx.hostEl.id + '_canonicalFacts', rows: 4, textContent: Array.isArray(value.canonical_facts) ? value.canonical_facts.join('\n') : '' })),
        createComponentField('Open Questions', el('textarea', { id: ctx.hostEl.id + '_openQuestions', rows: 4, textContent: Array.isArray(value.open_questions) ? value.open_questions.join('\n') : '' })),
        createComponentField('Active Arcs', el('textarea', { id: ctx.hostEl.id + '_activeArcs', rows: 4, textContent: Array.isArray(value.active_arcs) ? value.active_arcs.join('\n') : '' })),
        createComponentField('Metadata (JSON object)', el('textarea', { id: ctx.hostEl.id + '_metadata', rows: 5, textContent: metadataJSONText(value.metadata) })),
      ]));
      ctx.hostEl.appendChild(ce('div', { className: 'component-editor-grid' }, sections));
      break;
    case 'story_state':
      sections.push(createComponentSection('Structured Fields', [
        createComponentField('Current Situation', el('textarea', { id: ctx.hostEl.id + '_currentSituation', rows: 3, textContent: value.current_situation || '' })),
        createComponentField('Recent Changes', el('textarea', { id: ctx.hostEl.id + '_recentChanges', rows: 4, textContent: Array.isArray(value.recent_changes) ? value.recent_changes.join('\n') : '' })),
        createComponentField('Pending Threads', el('textarea', { id: ctx.hostEl.id + '_pendingThreads', rows: 4, textContent: Array.isArray(value.pending_threads) ? value.pending_threads.join('\n') : '' })),
        createComponentField('Tone', el('input', { type: 'text', id: ctx.hostEl.id + '_tone', value: value.tone || '' })),
        createComponentField('Metadata (JSON object)', el('textarea', { id: ctx.hostEl.id + '_metadata', rows: 5, textContent: metadataJSONText(value.metadata) })),
      ]));
      ctx.hostEl.appendChild(ce('div', { className: 'component-editor-grid' }, sections));
      break;
    case 'story_history': {
      var entryList = ce('div', { className: 'component-collection-list', id: ctx.hostEl.id + '_entries' }, []);
      (Array.isArray(value.entries) ? value.entries : []).forEach(function(item) { appendStoryHistoryEntryRow(entryList, item || {}); });
      sections.push(createComponentSection('Entries', [
        entryList,
        ce('div', { className: 'policy-actions component-editor-actions' }, [ce('button', { type: 'button', id: ctx.hostEl.id + '_addEntry' }, [ttxt('Add Entry')])]),
      ]));
      sections.push(createComponentSection('Metadata', [
        createComponentField('Metadata (JSON object)', el('textarea', { id: ctx.hostEl.id + '_metadata', rows: 5, textContent: metadataJSONText(value.metadata) })),
      ]));
      ctx.hostEl.appendChild(ce('div', { className: 'component-editor-grid' }, sections));
      document.getElementById(ctx.hostEl.id + '_addEntry').addEventListener('click', function() { appendStoryHistoryEntryRow(entryList, {}); });
      break;
    }
    case 'tick_policy':
      sections.push(createComponentSection('Structured Fields', [
        createComponentField('Continuity Rules', el('textarea', { id: ctx.hostEl.id + '_continuityRules', rows: 4, textContent: Array.isArray(value.continuity_rules) ? value.continuity_rules.join('\n') : '' })),
        createComponentField('Focus Scopes', el('textarea', { id: ctx.hostEl.id + '_focusScopes', rows: 4, textContent: Array.isArray(value.focus_scopes) ? value.focus_scopes.join('\n') : '' })),
        createComponentField('Banned Resets', el('textarea', { id: ctx.hostEl.id + '_bannedResets', rows: 4, textContent: Array.isArray(value.banned_resets) ? value.banned_resets.join('\n') : '' })),
        createComponentField('Metadata (JSON object)', el('textarea', { id: ctx.hostEl.id + '_metadata', rows: 5, textContent: metadataJSONText(value.metadata) })),
      ]));
      ctx.hostEl.appendChild(ce('div', { className: 'component-editor-grid' }, sections));
      break;
    case 'world_time_state': {
      var unitList = ce('div', { className: 'component-collection-list', id: ctx.hostEl.id + '_units' }, []);
      (Array.isArray(value.current_units) ? value.current_units : []).forEach(function(item) { appendCurrentUnitRow(unitList, item || {}); });
      sections.push(createComponentSection('Structured Fields', [
        createComponentField('Tick Scale Mode', el('select', { id: ctx.hostEl.id + '_tickScaleMode', innerHTML: '<option value="fixed">fixed</option><option value="flexible">flexible</option>' })),
        createComponentField('Tick Min Unit', el('input', { type: 'text', id: ctx.hostEl.id + '_tickMinUnit', value: value.tick_min_unit || '' })),
        createComponentField('Tick Step', el('input', { type: 'number', min: '0', id: ctx.hostEl.id + '_tickStep', value: Number.isFinite(value.tick_step) ? String(value.tick_step) : '1' })),
        createComponentField('Tick Units', el('textarea', { id: ctx.hostEl.id + '_tickUnits', rows: 4, textContent: Array.isArray(value.tick_units) ? value.tick_units.join('\n') : '' })),
        createComponentField('Calendar Name', el('input', { type: 'text', id: ctx.hostEl.id + '_calendarName', value: value.calendar_name || '' })),
        createComponentField('Current Time Label', el('input', { type: 'text', id: ctx.hostEl.id + '_currentTimeLabel', value: value.current_time_label || '' })),
        createComponentField('Total Ticks', el('input', { type: 'number', min: '0', id: ctx.hostEl.id + '_totalTicks', value: Number.isFinite(value.total_ticks) ? String(value.total_ticks) : '0' })),
        createComponentField('Last Tick Number', el('input', { type: 'number', min: '0', id: ctx.hostEl.id + '_lastTickNumber', value: Number.isFinite(value.last_tick_number) ? String(value.last_tick_number) : '0' })),
        createComponentField('Last Tick Type', el('input', { type: 'text', id: ctx.hostEl.id + '_lastTickType', value: value.last_tick_type || '' })),
        createComponentField('Last Advanced Ticks', el('input', { type: 'number', min: '0', id: ctx.hostEl.id + '_lastAdvancedTicks', value: Number.isFinite(value.last_advanced_ticks) ? String(value.last_advanced_ticks) : '0' })),
      ]));
      sections.push(createComponentSection('Current Units', [
        unitList,
        ce('div', { className: 'policy-actions component-editor-actions' }, [ce('button', { type: 'button', id: ctx.hostEl.id + '_addUnitValue' }, [ttxt('Add Unit Value')])]),
      ]));
      sections.push(createComponentSection('Metadata', [
        createComponentField('Metadata (JSON object)', el('textarea', { id: ctx.hostEl.id + '_metadata', rows: 5, textContent: metadataJSONText(value.metadata) })),
      ]));
      ctx.hostEl.appendChild(ce('div', { className: 'component-editor-grid' }, sections));
      document.getElementById(ctx.hostEl.id + '_tickScaleMode').value = value.tick_scale_mode || 'fixed';
      document.getElementById(ctx.hostEl.id + '_addUnitValue').addEventListener('click', function() { appendCurrentUnitRow(unitList, {}); });
      break;
    }
    default:
      ctx.hostEl.appendChild(ce('div', { className: 'hint', style: { textAlign: 'left', padding: '0' } }, [ttxt('Structured editor unavailable for this component type.')]));
      ensureComponentTextareaVisibility(ctx, true);
      ctx.hostEl.dataset.editorMode = 'free';
      return;
  }
  ctx.hostEl.dataset.editorMode = 'strong:' + componentType;
}

function renderComponentEditor(typeElementId, rawElementId, hostElementId, rawLabelId, hintElementId) {
  var ctx = getComponentEditorContext(typeElementId, rawElementId, hostElementId, rawLabelId, hintElementId);
  if (!ctx.typeEl || !ctx.rawEl || !ctx.hostEl) return;
  var componentType = ctx.typeEl.value || '';
  var mode = componentValidationMode(componentType);
  ctx.hostEl.innerHTML = '';
  if (mode === 'free') {
    ensureComponentTextareaVisibility(ctx, true);
    ctx.hostEl.dataset.editorMode = 'free';
    setComponentEditorHintText(ctx, componentType, '');
    return;
  }

  var raw = ctx.rawEl.value.trim();
  var parsed = raw ? safeParseJSONText(raw) : { value: defaultStructuredComponentPayload(componentType), error: '' };
  if (parsed.error || !parsed.value || Array.isArray(parsed.value) || typeof parsed.value !== 'object') {
    ensureComponentTextareaVisibility(ctx, true);
    ctx.hostEl.appendChild(ce('div', { className: 'hint', style: { textAlign: 'left', padding: '0 0 8px 0' } }, [ttxt('Malformed JSON detected. Fix raw content before using the structured editor.')]));
    ctx.hostEl.dataset.editorMode = 'fallback';
    setComponentEditorHintText(ctx, componentType, tr('Component data contains invalid JSON. Fix it in raw mode before saving.'));
    return;
  }

  ensureComponentTextareaVisibility(ctx, false);
  setComponentEditorHintText(ctx, componentType, '');
  if (mode === 'weak') {
    renderWeakComponentEditor(ctx, parsed.value);
    return;
  }
  renderStrongComponentEditor(ctx, componentType, parsed.value);
}

function parseJSONObjectEditorField(value, label) {
  var raw = String(value || '').trim();
  if (!raw) return {};
  var parsed = safeParseJSONText(raw);
  if (parsed.error || !parsed.value || Array.isArray(parsed.value) || typeof parsed.value !== 'object') {
    throw new Error(formatMessageWithDetail(label, tr('JSON object required for this component type')));
  }
  return parsed.value;
}

function parseJSONEditorField(value, label) {
  var raw = String(value || '').trim();
  if (!raw) return {};
  var parsed = safeParseJSONText(raw);
  if (parsed.error) {
    throw new Error(formatMessageWithDetail(label, parsed.error));
  }
  return parsed.value;
}

function collectWeakComponentEditorData(ctx) {
  var result = {};
  var rows = Array.prototype.slice.call(ctx.hostEl.querySelectorAll('.component-kv-row'));
  for (var i = 0; i < rows.length; i++) {
    var key = String(rows[i].querySelector('.component-kv-key').value || '').trim();
    var type = rows[i].querySelector('.component-kv-type').value || 'string';
    var rawValue = String(rows[i].querySelector('.component-kv-value').value || '').trim();
    if (!key && !rawValue) continue;
    if (!key) throw new Error(tr('Key cannot be empty'));
    if (Object.prototype.hasOwnProperty.call(result, key)) throw new Error(tr('Duplicate keys are not allowed'));
    if (type === 'number') {
      var num = rawValue === '' ? 0 : Number(rawValue);
      if (!Number.isFinite(num)) throw new Error(tr('Value must be a valid number'));
      result[key] = num;
    } else if (type === 'boolean') {
      if (rawValue !== 'true' && rawValue !== 'false') throw new Error(tr('Boolean values must be true or false'));
      result[key] = rawValue === 'true';
    } else if (type === 'json') {
      result[key] = parseJSONEditorField(rawValue || '{}', 'Value JSON');
    } else {
      result[key] = rawValue;
    }
  }
  return result;
}

function collectStrongComponentEditorData(ctx, componentType) {
  var root = ctx.hostEl;
  switch (componentType) {
    case 'autonomous':
      return {
        enabled: !!document.getElementById(root.id + '_enabled').checked,
        trigger: document.getElementById(root.id + '_trigger').value || 'manual',
        interval_seconds: parseInt(document.getElementById(root.id + '_interval').value, 10) || 0,
        capabilities: Array.prototype.slice.call(root.querySelectorAll('.component-collection-row')).map(function(row) {
          return {
            id: String(row.querySelector('.component-cap-id').value || '').trim(),
            mode: String(row.querySelector('.component-cap-mode').value || '').trim(),
            description: String(row.querySelector('.component-cap-desc').value || '').trim(),
            schema: parseJSONObjectEditorField(row.querySelector('.component-cap-schema').value, 'Capability Schema (JSON object)'),
          };
        }).filter(function(item) { return item.id || item.mode || item.description || Object.keys(item.schema || {}).length > 0; }),
        last_run_at: String(document.getElementById(root.id + '_lastRunAt').value || '').trim(),
        last_error: String(document.getElementById(root.id + '_lastError').value || '').trim(),
      };
    case 'world_state':
      return {
        summary: String(document.getElementById(root.id + '_summary').value || '').trim(),
        key_facts: splitEditorLines(document.getElementById(root.id + '_keyFacts').value),
        canonical_facts: splitEditorLines(document.getElementById(root.id + '_canonicalFacts').value),
        open_questions: splitEditorLines(document.getElementById(root.id + '_openQuestions').value),
        active_arcs: splitEditorLines(document.getElementById(root.id + '_activeArcs').value),
        metadata: parseJSONObjectEditorField(document.getElementById(root.id + '_metadata').value, 'Metadata (JSON object)'),
      };
    case 'story_state':
      return {
        current_situation: String(document.getElementById(root.id + '_currentSituation').value || '').trim(),
        recent_changes: splitEditorLines(document.getElementById(root.id + '_recentChanges').value),
        pending_threads: splitEditorLines(document.getElementById(root.id + '_pendingThreads').value),
        tone: String(document.getElementById(root.id + '_tone').value || '').trim(),
        metadata: parseJSONObjectEditorField(document.getElementById(root.id + '_metadata').value, 'Metadata (JSON object)'),
      };
    case 'story_history':
      return {
        entries: Array.prototype.slice.call(root.querySelectorAll('.component-collection-row')).map(function(row) {
          return {
            tick_number: parseInt(row.querySelector('.component-entry-tick').value, 10) || 0,
            summary: String(row.querySelector('.component-entry-summary').value || '').trim(),
            facts: splitEditorLines(row.querySelector('.component-entry-facts').value),
            game_time: String(row.querySelector('.component-entry-game-time').value || '').trim(),
          };
        }).filter(function(entry) { return entry.tick_number || entry.summary || entry.facts.length > 0 || entry.game_time; }),
        metadata: parseJSONObjectEditorField(document.getElementById(root.id + '_metadata').value, 'Metadata (JSON object)'),
      };
    case 'tick_policy':
      return {
        continuity_rules: splitEditorLines(document.getElementById(root.id + '_continuityRules').value),
        focus_scopes: splitEditorLines(document.getElementById(root.id + '_focusScopes').value),
        banned_resets: splitEditorLines(document.getElementById(root.id + '_bannedResets').value),
        metadata: parseJSONObjectEditorField(document.getElementById(root.id + '_metadata').value, 'Metadata (JSON object)'),
      };
    case 'world_time_state':
      return {
        tick_scale_mode: document.getElementById(root.id + '_tickScaleMode').value || 'fixed',
        tick_min_unit: String(document.getElementById(root.id + '_tickMinUnit').value || '').trim(),
        tick_step: parseInt(document.getElementById(root.id + '_tickStep').value, 10) || 0,
        tick_units: splitEditorLines(document.getElementById(root.id + '_tickUnits').value),
        calendar_name: String(document.getElementById(root.id + '_calendarName').value || '').trim(),
        current_units: Array.prototype.slice.call(root.querySelectorAll('.component-inline-row')).map(function(row) {
          return {
            unit: String(row.querySelector('.component-unit-name').value || '').trim(),
            value: String(row.querySelector('.component-unit-value').value || '').trim(),
          };
        }).filter(function(item) { return item.unit || item.value; }),
        current_time_label: String(document.getElementById(root.id + '_currentTimeLabel').value || '').trim(),
        total_ticks: parseInt(document.getElementById(root.id + '_totalTicks').value, 10) || 0,
        last_tick_number: parseInt(document.getElementById(root.id + '_lastTickNumber').value, 10) || 0,
        last_tick_type: String(document.getElementById(root.id + '_lastTickType').value || '').trim(),
        last_advanced_ticks: parseInt(document.getElementById(root.id + '_lastAdvancedTicks').value, 10) || 0,
        metadata: parseJSONObjectEditorField(document.getElementById(root.id + '_metadata').value, 'Metadata (JSON object)'),
      };
    default:
      return parseJSONObjectEditorField(ctx.rawEl.value, 'Component Data');
  }
}

function collectComponentEditorData(typeElementId, rawElementId, hostElementId, rawLabelId, hintElementId) {
  var ctx = getComponentEditorContext(typeElementId, rawElementId, hostElementId, rawLabelId, hintElementId);
  if (!ctx.typeEl || !ctx.rawEl || !ctx.hostEl) return { data: '' };
  var componentType = ctx.typeEl.value || '';
  var mode = componentValidationMode(componentType);
  if (mode === 'free' || ctx.hostEl.dataset.editorMode === 'fallback') {
    return { data: ctx.rawEl.value.trim() };
  }
  try {
    var payload = mode === 'weak' ? collectWeakComponentEditorData(ctx) : collectStrongComponentEditorData(ctx, componentType);
    var text = JSON.stringify(payload, null, 2);
    ctx.rawEl.value = text;
    return { data: text };
  } catch (e) {
    return { error: e && e.message ? e.message : String(e) };
  }
}

function getProjectedParentIds(nodeId) {
  var parentIds = [];
  var node = state.nodes.find(function(x) { return x.id === nodeId; });
  if (node && node.parent_id) parentIds.push(node.parent_id);
  return parentIds;
}

function buildProjectedChildMap(nodes) {
  var childMap = {};
  var parentMap = {};
  var visibleNodeSet = {};
  nodes.forEach(function(node) { visibleNodeSet[node.id] = true; });
  nodes.forEach(function(node) {
    var parentIds = getProjectedParentIds(node.id).filter(function(parentId) { return !!visibleNodeSet[parentId]; });
    parentMap[node.id] = parentIds.slice();
    if (parentIds.length === 0) {
      if (!childMap._root) childMap._root = [];
      childMap._root.push(node.id);
      return;
    }
    parentIds.forEach(function(parentId) {
      if (!childMap[parentId]) childMap[parentId] = [];
      if (childMap[parentId].indexOf(node.id) < 0) childMap[parentId].push(node.id);
    });
  });
  return { childMap: childMap, parentMap: parentMap };
}

function getNodeNameById(nodeId) {
  var node = state.nodes.find(function(x) { return x.id === nodeId; });
  return node ? node.name : nodeId;
}

function relationTypeDescription(relType) {
  switch (relType) {
    case 'belongs_to': return 'belongs_to（稳定归属）用于表达某节点属于某组织、阵营、资产体系或拥有者，不表示当前所处位置，也不会自动改变树中的 Primary Parent（主父节点）。';
    case 'located_at': return 'located_at（当前位置关系）用于表达节点此刻位于哪个地点、房间或场景中，不表示稳定归属，也不会自动改变树中的 Primary Parent（主父节点）。';
    case 'subordinate': return 'subordinate（隶属汇报）用于表达谁向谁负责、受谁管理，属于组织/控制链，不表示位置，也不会自动改变树中的 Primary Parent（主父节点）。';
    case 'external_parent': return 'external_parent（辅助父级作用域）用于补充第二条父向语义链，属于辅助 DAG（有向无环图）范围；默认不会进入树形大纲、不会进入默认上下文，也不会参与默认传播。';
    case 'ally': return '社会协作关系。用于表达盟友、合作或友方网络，默认不参与节点身份树、环境树和默认上下文扩展。';
    case 'enemy': return '社会对抗关系。用于表达敌对或竞争网络，默认不参与节点身份树、环境树和默认上下文扩展。';
    case 'kinship': return '社会背景关系。用于表达亲属、血缘或家族关系，默认不参与节点身份树、环境树和默认上下文扩展。';
    default: return '';
  }
}

function relationSemanticsHint() {
  return tr('Relations are stored separately from the outline tree. Primary Parent remains the only stable hierarchy field. Use located_at for current environment, belongs_to/subordinate for organization or control, and external_parent only for auxiliary DAG scope.');
}

function validPropagationModes() {
  return ['upward', 'environment_scope', 'organization_scope', 'tag_broadcast', 'targeted', 'manual'];
}

function getNodeById(nodeId) {
  return (state.nodes || []).find(function(node) { return node.id === nodeId; }) || null;
}

function relationFormWarnings(sourceId, targetId, relType) {
  var warnings = [];
  var source = getNodeById(sourceId);
  var target = getNodeById(targetId);
  if (!source || !target || !relType) return warnings;
  if (relType === 'located_at') {
    if (source.parent_id === targetId) {
      warnings.push(tr('This target is already the Primary Parent. If the node is only temporarily here, prefer keeping Primary Parent stable and using located_at for movement.'));
    }
    var otherLocatedAt = (state.relations || []).some(function(rel) {
      return rel.source_id === sourceId && rel.target_id !== targetId && rel.relation_type === 'located_at';
    });
    if (otherLocatedAt) {
      warnings.push(tr('This node already has another located_at relation. Keep only one active current environment unless you intentionally model multiple simultaneous positions.'));
    }
  }
  if (relType === 'belongs_to' || relType === 'subordinate') {
    if (source.parent_id === targetId) {
      warnings.push(tr('This target is already the Primary Parent. If you want to change stable hierarchy, edit Primary Parent directly. Use this relation only to add organization/control semantics.'));
    }
  }
  if (relType === 'external_parent') {
    warnings.push(tr('external_parent is auxiliary DAG scope only. It is excluded from default context assembly and default propagation, so do not rely on it for current location or primary organization modeling.'));
    if ((state.relations || []).some(function(rel) {
      return rel.source_id === sourceId && rel.target_id !== targetId && rel.relation_type === 'external_parent';
    })) {
      warnings.push(tr('This node already has another external_parent relation. Keep this relation rare and use it only when a second parent-like scope is truly required.'));
    }
  }
  if (relType === 'ally' || relType === 'enemy' || relType === 'kinship') {
    warnings.push(tr('Social relations are background graph edges. They are not part of the default hierarchy walk or default environment context expansion.'));
  }
  return warnings;
}

function propagationModeDescription(mode) {
  switch (mode) {
    case 'upward':
      return tr('Follow the stable Primary Parent chain upward. This is the default mode for promoting local memory into broader identity context.');
    case 'environment_scope':
      return tr('Publish through the current environment chain rooted by located_at, then optionally continue upward when Publish Up is enabled.');
    case 'organization_scope':
      return tr('Publish through organization/control links such as belongs_to and subordinate, then optionally continue upward when Publish Up is enabled.');
    case 'tag_broadcast':
      return tr('Publish to nodes matched by propagation tags. Use this for cross-cutting subscriptions rather than structural graph flow.');
    case 'targeted':
      return tr('Publish only to the explicit target node IDs. Use this when the recipients are known ahead of time.');
    case 'manual':
      return tr('Record a manual propagation request without relying on graph traversal defaults.');
    default:
      return '';
  }
}

function updatePropagationModePreview() {
  var modeEl = document.getElementById('propMemMode');
  var descEl = document.getElementById('propMemModeMeaning');
  if (!modeEl || !descEl) return;
  descEl.textContent = propagationModeDescription(modeEl.value);
}

function parseMultilineList(value) {
  return String(value || '').split('\n').map(function(item) { return item.trim(); }).filter(Boolean);
}

function parseWorldTimeCarryLines(value) {
  return parseMultilineList(value).map(function(line) {
    var parts = line.split('=');
    if (parts.length !== 2) return null;
    var edge = parts[0].split('->');
    if (edge.length !== 2) return null;
    var base = parseInt(parts[1].trim(), 10);
    if (!Number.isFinite(base)) return null;
    return { from: edge[0].trim(), to: edge[1].trim(), base: base };
  }).filter(Boolean);
}

function parseWorldTimeCalendarLines(value) {
  return parseMultilineList(value).map(function(line) {
    var parts = line.split('=');
    if (parts.length < 2) return null;
    return { unit: parts[0].trim(), value: parts.slice(1).join('=').trim() };
  }).filter(Boolean);
}

function parseWorldTimeSequenceLines(value) {
  return parseMultilineList(value).map(function(line) {
    var parts = line.split('=');
    if (parts.length < 2) return null;
    return {
      unit: parts[0].trim(),
      values: parts.slice(1).join('=').split('|').map(function(item) { return item.trim(); }).filter(Boolean),
    };
  }).filter(Boolean);
}

function collectWorldTimeSettingsFromForm() {
  var tickScaleModeEl = document.getElementById('setTickScaleMode');
  if (!tickScaleModeEl) return null;

  var tickUnits = Array.prototype.slice.call(document.querySelectorAll('.world-time-unit-input')).map(function(input) {
    return String(input.value || '').trim();
  }).filter(Boolean);
  var minUnitEl = document.getElementById('setTickMinUnit');
  var timeScaleCarry = Array.prototype.slice.call(document.querySelectorAll('.world-time-carry-base')).map(function(input) {
    var base = parseInt(String(input.value || '').trim(), 10);
    return {
      from: String(input.dataset.from || '').trim(),
      to: String(input.dataset.to || '').trim(),
      base: Number.isFinite(base) ? base : 0,
    };
  }).filter(function(rule) { return rule.from && rule.to; });
  var unitValueSequences = Array.prototype.slice.call(document.querySelectorAll('.world-time-sequence-values')).map(function(input) {
    return {
      unit: String(input.dataset.unit || '').trim(),
      values: String(input.value || '').split('|').map(function(item) { return item.trim(); }).filter(Boolean),
    };
  }).filter(function(seq) { return seq.unit && seq.values.length > 0; });

  var calendarEnabled = !!document.getElementById('setCalendarEnabled').checked;
  var timeCalendar = {
    enabled: calendarEnabled,
    calendar_name: document.getElementById('setCalendarName').value.trim(),
    units: Array.prototype.slice.call(document.querySelectorAll('.world-time-calendar-value')).map(function(input) {
      return {
        unit: String(input.dataset.unit || '').trim(),
        value: String(input.value || '').trim(),
      };
    }).filter(function(unit) { return unit.unit || unit.value; }),
  };
  if (!calendarEnabled) {
    timeCalendar = { enabled: false, calendar_name: '', units: [] };
  }

  return {
    tick_scale_mode: tickScaleModeEl.value || 'fixed',
    tick_min_unit: minUnitEl ? String(minUnitEl.value || '').trim() : '',
    tick_step: parseInt(document.getElementById('setTickStep').value.trim(), 10) || 0,
    tick_units: tickUnits,
    time_scale_carry: timeScaleCarry,
    time_calendar: timeCalendar,
    unit_value_sequences: unitValueSequences,
  };
}

function validateWorldTimeSettingsInput(settings) {
  if (!settings) return '';
  if (settings.tick_scale_mode !== 'fixed' && settings.tick_scale_mode !== 'flexible') return tr('tick_scale_mode must be fixed or flexible');
  if (!settings.tick_min_unit) return tr('tick_min_unit must not be empty');
  if (!Number.isInteger(settings.tick_step) || settings.tick_step <= 0) return tr('tick_step must be greater than 0');
  if (!Array.isArray(settings.tick_units) || settings.tick_units.length === 0) return tr('tick_units must contain at least one unit');
  var seen = {};
  var normalizedUnits = [];
  for (var i = 0; i < settings.tick_units.length; i++) {
    var unit = String(settings.tick_units[i] || '').trim();
    if (!unit) return tr('tick_units must not contain empty values');
    if (seen[unit]) return tr('tick_units must not contain duplicate values');
    seen[unit] = true;
    normalizedUnits.push(unit);
  }
  var tickMinUnit = String(settings.tick_min_unit || '').trim();
  if (normalizedUnits[normalizedUnits.length - 1] !== tickMinUnit) return tr('tick_min_unit must match the smallest configured tick unit');

  var carryRules = Array.isArray(settings.time_scale_carry) ? settings.time_scale_carry : [];
  if (normalizedUnits.length > 1 && carryRules.length !== normalizedUnits.length - 1) {
    return tr('time_scale_carry must define exactly one adjacent rule per unit gap');
  }
  var carryByFrom = {};
  for (var j = 0; j < carryRules.length; j++) {
    var rule = carryRules[j] || {};
    var from = String(rule.from || '').trim();
    var to = String(rule.to || '').trim();
    var expectedFrom = normalizedUnits[normalizedUnits.length - 1 - j];
    var expectedTo = normalizedUnits[normalizedUnits.length - 2 - j];
    if (!rule.from || !rule.to || !Number.isInteger(rule.base) || rule.base <= 0) {
      return tr('time_scale_carry entries must define from, to, and base > 0');
    }
    if (from !== expectedFrom || to !== expectedTo) {
      return formatWorldTimeValidationError('carry_rule_order', { index: j, expectedFrom: expectedFrom, expectedTo: expectedTo });
    }
    carryByFrom[from] = rule;
  }

  var sequences = Array.isArray(settings.unit_value_sequences) ? settings.unit_value_sequences : [];
  var sequenceByUnit = {};
  for (var k = 0; k < sequences.length; k++) {
    var seq = sequences[k] || {};
    var seqUnit = String(seq.unit || '').trim();
    if (!seqUnit) return formatWorldTimeValidationError('sequence_unit_empty', { index: k });
    if (!seen[seqUnit]) return formatWorldTimeValidationError('sequence_unit_missing', { index: k, unit: seqUnit });
    if (seqUnit === normalizedUnits[0]) return formatWorldTimeValidationError('sequence_unit_largest', { index: k, unit: seqUnit });
    if (!Array.isArray(seq.values) || seq.values.length === 0) return formatWorldTimeValidationError('sequence_values_empty', { index: k });
    if (!carryByFrom[seqUnit]) return formatWorldTimeValidationError('sequence_missing_carry_rule', { index: k, unit: seqUnit });
    if (seq.values.length !== carryByFrom[seqUnit].base) {
      return formatWorldTimeValidationError('sequence_values_count', { index: k, unit: seqUnit, expectedCount: carryByFrom[seqUnit].base });
    }
    sequenceByUnit[seqUnit] = seq.values.map(function(item) { return String(item || '').trim(); });
  }

  if (settings.time_calendar && settings.time_calendar.enabled) {
    if (!settings.time_calendar.calendar_name) return tr('time_calendar.calendar_name must not be empty when calendar mode is enabled');
    if (!Array.isArray(settings.time_calendar.units) || settings.time_calendar.units.length !== settings.tick_units.length) {
      return tr('time_calendar.units must match tick_units exactly when calendar mode is enabled');
    }
    for (var m = 0; m < settings.time_calendar.units.length; m++) {
      var calendarUnit = settings.time_calendar.units[m] || {};
      var calendarUnitName = String(calendarUnit.unit || '').trim();
      var calendarValue = String(calendarUnit.value || '').trim();
      if (!calendarUnitName) return formatWorldTimeValidationError('calendar_unit_empty', { index: m });
      if (calendarUnitName !== normalizedUnits[m]) return formatWorldTimeValidationError('calendar_unit_mismatch', { index: m, expectedUnit: normalizedUnits[m] });
      if (!calendarValue) return formatWorldTimeValidationError('calendar_value_empty', { index: m });
      if (/^-?\d+$/.test(calendarValue)) {
        var numericValue = parseInt(calendarValue, 10);
        if (carryByFrom[calendarUnitName]) {
          if (numericValue < 0 || numericValue >= carryByFrom[calendarUnitName].base) {
            return formatWorldTimeValidationError('calendar_value_range', { index: m, unit: calendarUnitName, maxValue: carryByFrom[calendarUnitName].base - 1 });
          }
        }
      } else {
        if (!sequenceByUnit[calendarUnitName] || sequenceByUnit[calendarUnitName].length === 0) {
          return formatWorldTimeValidationError('calendar_unit_requires_sequences', { index: m, unit: calendarUnitName });
        }
        if (sequenceByUnit[calendarUnitName].indexOf(calendarValue) < 0) {
          return formatWorldTimeValidationError('calendar_value_missing_from_sequences', { index: m, value: calendarValue, unit: calendarUnitName });
        }
      }
    }
    if (String(settings.time_calendar.units[settings.time_calendar.units.length - 1].unit || '').trim() !== tickMinUnit) {
      return formatWorldTimeValidationError('calendar_smallest_unit_mismatch');
    }
  }
  return '';
}

function filterNodeOptions(searchText, excludeNodeIds) {
  var keyword = (searchText || '').trim().toLowerCase();
  var excluded = excludeNodeIds || [];
  return state.nodes.filter(function(node) {
    if (excluded.indexOf(node.id) >= 0) return false;
    if (!keyword) return true;
    return node.name.toLowerCase().includes(keyword) || node.node_type.toLowerCase().includes(keyword) || node.id.toLowerCase().includes(keyword);
  });
}

function renderRelationPreview(sourceId, relType, targetId) {
  var sourceName = sourceId ? getNodeNameById(sourceId) : '-';
  var targetName = targetId ? getNodeNameById(targetId) : '-';
  return sourceName + ' --[' + (relType || '-') + ']--> ' + targetName;
}

function bindRelationNodeFilter(inputId, selectId, excludeNodeIds, selectedValue) {
  var input = document.getElementById(inputId);
  var select = document.getElementById(selectId);
  if (!input || !select) return;
  function refill() {
    var nodes = filterNodeOptions(input.value, excludeNodeIds);
    select.innerHTML = nodes.map(function(node) {
      return '<option value="' + node.id + '">' + node.name + ' (' + node.node_type + ')</option>';
    }).join('');
    if (selectedValue && nodes.some(function(node) { return node.id === selectedValue; })) {
      select.value = selectedValue;
    }
  }
  input.addEventListener('input', refill);
  refill();
}

function updateRelationFormPreview(prefix) {
  var sourceEl = document.getElementById(prefix + 'RelSource');
  var targetEl = document.getElementById(prefix + 'RelTarget');
  var typeEl = document.getElementById(prefix + 'RelType');
  var previewEl = document.getElementById(prefix + 'RelPreview');
  var descEl = document.getElementById(prefix + 'RelMeaning');
  var warnEl = document.getElementById(prefix + 'RelWarnings');
  if (!sourceEl || !targetEl || !typeEl || !previewEl || !descEl) return;
  previewEl.textContent = renderRelationPreview(sourceEl.value, typeEl.value, targetEl.value);
  descEl.textContent = relationTypeDescription(typeEl.value);
  if (warnEl) {
    var warnings = relationFormWarnings(sourceEl.value, targetEl.value, typeEl.value);
    warnEl.textContent = warnings.join(' ');
    warnEl.style.display = warnings.length > 0 ? 'block' : 'none';
  }
}

async function selectNode(nodeId, nodeType, options) {
  options = options || {};
  var mode = options.mode || 'single';
  var preserveAnchor = !!options.preserveAnchor;
  var treePathKey = options.treePathKey || null;
  var selectedIds = state.selectedNodeIds ? state.selectedNodeIds.slice() : [];
  var nextSelectedNodeId = nodeId || null;
  var nextAnchorId = preserveAnchor ? state.selectionAnchorId : (nodeId || null);

  if (mode === 'single') {
    selectedIds = nodeId ? [nodeId] : [];
  } else if (mode === 'toggle') {
    var idx = selectedIds.indexOf(nodeId);
    if (idx >= 0) {
      selectedIds.splice(idx, 1);
      if (state.selectedNodeId === nodeId) {
        nextSelectedNodeId = selectedIds.length > 0 ? selectedIds[selectedIds.length - 1] : null;
      } else {
        nextSelectedNodeId = state.selectedNodeId;
      }
    } else {
      selectedIds.push(nodeId);
      nextSelectedNodeId = nodeId;
    }
  } else if (mode === 'range') {
    var order = getVisibleTreeOrder();
    var anchor = state.selectionAnchorId || state.selectedNodeId || nodeId;
    var ai = order.indexOf(anchor);
    var bi = order.indexOf(nodeId);
    if (ai >= 0 && bi >= 0) {
      var start = Math.min(ai, bi);
      var end = Math.max(ai, bi);
      selectedIds = order.slice(start, end + 1);
      preserveAnchor = true;
      nextAnchorId = anchor;
    } else {
      selectedIds = nodeId ? [nodeId] : [];
    }
  } else if (mode === 'preserve') {
    if (nodeId && selectedIds.indexOf(nodeId) < 0) selectedIds.push(nodeId);
  }

  selectedIds = uniqueNodeIds(selectedIds);
  if (nextSelectedNodeId && selectedIds.indexOf(nextSelectedNodeId) < 0) {
    nextSelectedNodeId = selectedIds.length > 0 ? selectedIds[selectedIds.length - 1] : null;
  }
  state.selectedNodeId = nextSelectedNodeId;
  state.selectedNodeIds = selectedIds;
  state.selectedNodeType = nodeType || null;
  if (treePathKey || mode === 'single') state.selectedTreePathKey = treePathKey;
  if (preserveAnchor) {
    if (nextAnchorId && selectedIds.indexOf(nextAnchorId) < 0) {
      nextAnchorId = selectedIds.length > 0 ? selectedIds[0] : null;
    }
    state.selectionAnchorId = nextAnchorId || null;
  } else {
    state.selectionAnchorId = nodeId || null;
  }
  renderTree();
  if (!state.selectedNodeId) { state.nodeDetail = null; renderCurrent(); return; }
  if (state.selectedNodeId !== nodeId) {
    var currentNode = state.nodes.find(function(x) { return x.id === state.selectedNodeId; });
    nodeId = state.selectedNodeId;
    nodeType = currentNode ? currentNode.node_type : null;
    state.selectedNodeType = nodeType || null;
  }
  try { state.nodeDetail = await api('GET', '/api/v1/nodes/' + encodeURIComponent(nodeId)); }
  catch(e) { state.nodeDetail = null; showApiError('Failed to load node details', e); }
  loadAutonomous(); renderCurrent(); if (typeof updateActionButtons === 'function') updateActionButtons();
}

async function loadPolicy() {
  if (!state.selectedWorldId) return;
  try { state.policy = await api('GET', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/policy'); }
  catch(e) { state.policy = null; }
  if (state.page === 'policy') renderCurrent();
}

async function loadSettings() {
  if (!state.selectedWorldId) return;
  try { state.settings = await api('GET', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/settings'); }
  catch(e) { state.settings = null; }
  if (state.page === 'settings') renderCurrent();
}

async function loadPlans(silent) {
  try {
    var q = '/api/v1/plans/pending';
    if (state.selectedWorldId) q += '?world_id=' + encodeURIComponent(state.selectedWorldId);
    state.plans = await api('GET', q);
    if (state.page === 'plans') renderCurrent();
    if (!silent) toast(tr('Plans refreshed'), 'success');
  } catch(e) {
    state.plans = [];
    if (state.page === 'plans') renderCurrent();
    if (!silent) showApiError('Failed', e);
  }
}

async function approvePlan(worldId, planId) {
  showLoading(tr('Processing...'));
  try {
    await api('POST', '/api/v1/worlds/' + encodeURIComponent(worldId) + '/plan/approve', { plan_id: planId });
    hideLoading();
    toast(tr('Plan approved'), 'success');
    loadPlans();
  } catch(e) {
    hideLoading();
    showApiError('Failed', e);
  }
}

async function rejectPlan(worldId, planId) {
  showLoading(tr('Processing...'));
  try {
    await api('POST', '/api/v1/worlds/' + encodeURIComponent(worldId) + '/plan/reject', { plan_id: planId });
    hideLoading();
    toast(tr('Plan rejected'), 'success');
    loadPlans();
  } catch(e) {
    hideLoading();
    showApiError('Failed', e);
  }
}

async function loadLogs() {
  try {
    var q = '/api/v1/logs?limit=100';
    if (state.selectedWorldId) q += '&world_id=' + encodeURIComponent(state.selectedWorldId);
    state.logs = await api('GET', q);
    if (state.page === 'logs') renderCurrent();
  } catch(e) {
    state.logs = [];
    if (state.page === 'logs') renderCurrent();
    showApiError('Failed to load logs', e);
  }
}

function getContinuityRequestIds(bundle) {
  var seen = {};
  var result = [];
  var logs = bundle && bundle.logs ? bundle.logs : [];
  var traces = bundle && bundle.traces ? bundle.traces : [];
  logs.forEach(function(log) {
    if (!log || !log.request_id || seen[log.request_id]) return;
    seen[log.request_id] = true;
    result.push(log.request_id);
  });
  traces.forEach(function(trace) {
    if (!trace || !trace.request_id || seen[trace.request_id]) return;
    seen[trace.request_id] = true;
    result.push(trace.request_id);
  });
  return result;
}

function getContinuityModes(bundle) {
  var seen = {};
  var result = [];
  var logs = bundle && bundle.logs ? bundle.logs : [];
  logs.forEach(function(log) {
    if (!log || !log.execution_mode || seen[log.execution_mode]) return;
    seen[log.execution_mode] = true;
    result.push(log.execution_mode);
  });
  return result;
}

function getFilteredContinuityLogs(bundle) {
  var logs = bundle && bundle.logs ? bundle.logs.slice() : [];
  if (state.continuityRequestId) {
    logs = logs.filter(function(log) { return log.request_id === state.continuityRequestId; });
  }
  if (state.continuityMode) {
    logs = logs.filter(function(log) { return log.execution_mode === state.continuityMode; });
  }
  return logs;
}

function getFilteredContinuityTraces(bundle) {
  var traces = bundle && bundle.traces ? bundle.traces.slice() : [];
  if (state.continuityRequestId) {
    traces = traces.filter(function(trace) { return trace.request_id === state.continuityRequestId; });
  }
  if (state.continuityMode) {
    var allowed = {};
    (bundle && bundle.logs ? bundle.logs : []).forEach(function(log) {
      if (log.execution_mode === state.continuityMode && log.request_id) allowed[log.request_id] = true;
    });
    if (Object.keys(allowed).length > 0) {
      traces = traces.filter(function(trace) { return !!allowed[trace.request_id]; });
    }
  }
  return traces;
}

async function loadContinuityOverview() {
  if (!state.selectedWorldId) {
    state.continuityBundle = null;
    if (state.page === 'continuity') renderCurrent();
    return;
  }
  var worldID = encodeURIComponent(state.selectedWorldId);
  try {
    var results = await Promise.allSettled([
      api('GET', '/api/v1/worlds/' + worldID + '/timelines/latest'),
      api('GET', '/api/v1/worlds/' + worldID + '/timelines?limit=6'),
      api('GET', '/api/v1/worlds/' + worldID + '/state-components'),
      api('GET', '/api/v1/logs?world_id=' + worldID + '&task_type=world_tick&limit=60'),
      api('GET', '/debug/traces?world_id=' + worldID + '&limit=30'),
    ]);
    var bundle = {
      world_id: state.selectedWorldId,
      latest_timeline: null,
      timelines: [],
      state_components: [],
      logs: [],
      traces: [],
    };
    if (results[0].status === 'fulfilled' && results[0].value) bundle.latest_timeline = results[0].value.timeline || null;
    if (results[1].status === 'fulfilled' && results[1].value) bundle.timelines = results[1].value.timelines || [];
    if (results[2].status === 'fulfilled' && results[2].value) bundle.state_components = results[2].value.components || [];
    if (results[3].status === 'fulfilled' && results[3].value) bundle.logs = results[3].value || [];
    if (results[4].status === 'fulfilled' && results[4].value) bundle.traces = results[4].value.traces || [];
    state.continuityBundle = bundle;
    if (state.continuityRequestId && getContinuityRequestIds(bundle).indexOf(state.continuityRequestId) < 0) state.continuityRequestId = '';
    if (state.continuityMode && getContinuityModes(bundle).indexOf(state.continuityMode) < 0) state.continuityMode = '';
    if (state.page === 'continuity') renderCurrent();
  } catch (e) {
    state.continuityBundle = null;
    if (state.page === 'continuity') renderCurrent();
    showApiError('Failed to load continuity bundle', e);
  }
}

async function loadAutonomous() {
  if (!state.selectedNodeId) return;
  try { state.autonomous = await api('GET', '/api/v1/nodes/' + encodeURIComponent(state.selectedNodeId) + '/autonomous'); }
  catch(e) { state.autonomous = null; }
}

async function loadCurrentWorld() {
  if (!state.selectedWorldId) return;
  try {
    state.nodes = await api('GET', '/api/v1/nodes?world_id=' + encodeURIComponent(state.selectedWorldId));
    try { state.relations = await api('GET', '/api/v1/relations?world_id=' + encodeURIComponent(state.selectedWorldId)); }
    catch(e) { state.relations = []; }
    var valid = {};
    for (var i = 0; i < state.nodes.length; i++) valid[state.nodes[i].id] = true;
    state.selectedNodeIds = (state.selectedNodeIds || []).filter(function(id) { return valid[id]; });
    if (state.selectedNodeId && !valid[state.selectedNodeId]) {
      state.selectedNodeId = state.selectedNodeIds.length > 0 ? state.selectedNodeIds[0] : null;
    }
    if (state.selectionAnchorId && !valid[state.selectionAnchorId]) state.selectionAnchorId = state.selectedNodeId || null;
    renderTree();
    if (state.selectedNodeId) {
      var sn = state.nodes.find(function(x) { return x.id === state.selectedNodeId; });
      selectNode(state.selectedNodeId, sn ? sn.node_type : null, { mode: 'preserve', preserveAnchor: true });
    } else {
      state.nodeDetail = null;
      renderCurrent();
    }
  } catch(e) { showApiError('Refresh failed', e); }
}

async function loadSnapshots() {
  if (!state.selectedWorldId) {
    state.snapshots = [];
    state.snapshotMeta = null;
    state.snapshotListWorldId = null;
    return;
  }
  try {
    state.snapshotMeta = await api('GET', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/snapshot-metadata');
  } catch(e) {
    state.snapshotMeta = null;
  }
  var listWorldID = state.selectedWorldId;
  if (state.snapshotMeta && state.snapshotMeta.reason === 'save_snapshot' && state.snapshotMeta.source_world_id) {
    listWorldID = state.snapshotMeta.source_world_id;
  }
  state.snapshotListWorldId = listWorldID;
  try {
    state.snapshots = await api('GET', '/api/v1/worlds/' + encodeURIComponent(listWorldID) + '/snapshots');
  } catch(e) {
    state.snapshots = [];
  }
  if (state.page === 'snapshots') renderCurrent();
}

async function refreshSnapshots() {
  showLoading(tr('Loading snapshots...'));
  try {
    await loadSnapshots();
    hideLoading();
    toast(tr('Snapshots refreshed'), 'success');
  } catch(e) {
    hideLoading();
    showApiError('Failed', e);
  }
}

async function validateSnapshot(snapshotWorldID) {
  showLoading(tr('Validating snapshot...'));
  try {
    var result = await api('GET', '/api/v1/worlds/' + encodeURIComponent(snapshotWorldID) + '/snapshot-validation');
    hideLoading();
    openSnapshotValidationModal(result);
  } catch(e) {
    hideLoading();
    showApiError('Failed', e);
  }
}

function openSnapshotValidationModal(result) {
  var issues = result.issues || [];
  var body = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Snapshot')]), ce('span', { className: 'val mono' }, [txt(result.snapshot_world_id || '-')])]),
    ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Valid')]), ce('span', { className: 'val' }, [txt(result.valid ? tr('Yes') : tr('No'))])]),
    ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Schema')]), ce('span', { className: 'val' }, [txt(result.schema_version || '-')])]),
    ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Engine')]), ce('span', { className: 'val' }, [txt((result.engine_version || '-') + ' / ' + (result.current_engine_version || '-'))])]),
    ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Issues')]), ce('span', { className: 'val' }, [txt(String(issues.length))])]),
    issues.length > 0 ? el('pre', { style: {fontSize: '11px', whiteSpace: 'pre-wrap', maxHeight: '220px', overflow: 'auto', background: 'var(--bg-input)', padding: '8px', borderRadius: 'var(--radius)'}, textContent: JSON.stringify(issues, null, 2) }) : ce('div', { className: 'hint' }, [ttxt('No validation issues.')]),
  ]);
  openModal(tr('Snapshot Validation'), body, ce('div', {}, [el('button', { id: 'modalCloseSnapshotValidationBtn', textContent: tr('Close') })]));
  document.getElementById('modalCloseSnapshotValidationBtn').addEventListener('click', closeModal);
}

function openRestoreSnapshotModal(snapshotWorldID, snapshotName) {
  var f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'restoreSnapshotName' }, [ttxt('New World Name')]),
    el('input', { id: 'restoreSnapshotName', value: (snapshotName || '') + ' restored', style: {width: '100%'} }),
    ce('label', { className: 'checkbox-row' }, [el('input', { id: 'restoreSnapshotLockWorld', type: 'checkbox', checked: true }), ttxt('Lock source snapshot during restore')]),
  ]);
  openModal(tr('Restore Snapshot'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalRestoreSnapshotBtn' }, [ttxt('Restore')]), el('button', { id: 'modalCancelSnapshotRestoreBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalRestoreSnapshotBtn').addEventListener('click', function() { restoreSnapshot(snapshotWorldID); });
  document.getElementById('modalCancelSnapshotRestoreBtn').addEventListener('click', closeModal);
}

async function restoreSnapshot(snapshotWorldID) {
  var name = document.getElementById('restoreSnapshotName').value.trim();
  var lockWorld = document.getElementById('restoreSnapshotLockWorld').checked;
  showLoading(tr('Restoring snapshot...'));
  try {
    var result = await api('POST', '/api/v1/worlds/' + encodeURIComponent(snapshotWorldID) + '/restore', { name: name, lock_world: lockWorld });
    closeModal();
    hideLoading();
    toast(tr('Snapshot restored'), 'success');
    await loadWorlds();
    await loadSnapshots();
    if (result && result.id) selectWorld(result.id);
  } catch(e) {
    hideLoading();
    showApiError('Failed', e);
  }
}

function deleteSnapshot(snapshotWorldID) {
  var body = ce('p', { style: { color: 'var(--text)', fontSize: '12px' } }, [ttxt('Delete this snapshot world and its metadata?')]);
  var footer = ce('div', {}, [
    ce('button', { className: 'danger', id: 'modalConfirmDeleteSnapshotBtn' }, [ttxt('Delete')]),
    el('button', { id: 'modalCancelDeleteSnapshotBtn', textContent: tr('Cancel') }),
  ]);
  openModal(tr('Confirm'), body, footer);
  document.getElementById('modalConfirmDeleteSnapshotBtn').addEventListener('click', async function() {
    closeModal();
    showLoading(tr('Deleting snapshot...'));
    try {
      await api('DELETE', '/api/v1/worlds/' + encodeURIComponent(snapshotWorldID) + '/snapshot');
      hideLoading();
      toast(tr('Snapshot deleted'), 'success');
      var deletedSelected = state.selectedWorldId === snapshotWorldID;
      await loadWorlds();
      if (deletedSelected) {
        state.selectedWorldId = state.worlds.length > 0 ? state.worlds[0].id : null;
        if (state.selectedWorldId) {
          await selectWorld(state.selectedWorldId);
          return;
        }
      }
      await loadSnapshots();
      renderCurrent();
    } catch(e) {
      hideLoading();
      showApiError('Failed', e);
    }
  });
  document.getElementById('modalCancelDeleteSnapshotBtn').addEventListener('click', closeModal);
}

/* ============= Create Modals ============= */
function openCreateWorldModal() {
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'createWorldName' }, [ttxt('World Name')]),
    el('input', { id: 'createWorldName', placeholder: tr('Enter world name...'), style: {width: '100%'} }),
  ]);
  openModal(tr('Create World'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalCreateWorldBtn' }, [ttxt('Create')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalCreateWorldBtn').addEventListener('click', createWorld);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function forkWorld() {
  if (!state.selectedWorldId) { toast(tr('Select a world first'), 'error'); return; }
  const lockWorld = confirm(tr('Lock world during working-copy creation? This prevents concurrent writes.'));
  showLoading(tr('Creating working copy...'));
  try {
    const result = await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/fork', { lock_world: lockWorld });
    hideLoading(); toast(tr('Working copy created'), 'success');
    await loadWorlds();
    if (result && result.id) selectWorld(result.id);
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

async function saveSnapshot() {
  if (!state.selectedWorldId) { toast(tr('Select a world first'), 'error'); return; }
  const lockWorld = confirm(tr('Lock world during snapshot save? This prevents concurrent writes.'));
  showLoading(tr('Saving snapshot...'));
  try {
    const result = await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/snapshots', { lock_world: lockWorld });
    hideLoading(); toast(tr('Snapshot saved'), 'success');
    await loadWorlds();
    if (result && result.id) selectWorld(result.id);
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

async function createWorld() {
  const name = document.getElementById('createWorldName').value.trim();
  if (!name) { toast(tr('Enter a world name'), 'error'); return; }
  try {
    await api('POST', '/api/v1/nodes', { name: name, node_type: 'world' });
    closeModal(); toast(tr('World created'), 'success'); loadWorlds();
  } catch(e) { showApiError('Failed', e); }
}

function openEditWorldModal() {
  if (!state.selectedWorldId) { toast(tr('Select a world first'), 'error'); return; }
  var world = state.worlds.find(function(x) { return x.id === state.selectedWorldId; });
  if (!world) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'editWorldName' }, [ttxt('World Name')]),
    el('input', { id: 'editWorldName', value: world.name || '', style: {width: '100%'} }),
  ]);
  openModal(tr('Edit World'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditWorldBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelWorldEditBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalEditWorldBtn').addEventListener('click', editWorld);
  document.getElementById('modalCancelWorldEditBtn').addEventListener('click', closeModal);
}

async function editWorld() {
  if (!state.selectedWorldId) return;
  const name = document.getElementById('editWorldName').value.trim();
  if (!name) { toast(tr('Enter a world name'), 'error'); return; }
  try {
    await api('PUT', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId), { name: name });
    closeModal();
    toast(tr('World updated'), 'success');
    await loadWorlds();
    await selectWorld(state.selectedWorldId);
  } catch(e) {
    showApiError('Failed', e);
  }
}

function openCreateNodeModal(parentId) {
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'createNodeName' }, [ttxt('Node Name')]),
    el('input', { id: 'createNodeName', placeholder: tr('Enter name...'), style: {width: '100%'} }),
    ce('label', { for: 'createNodeType' }, [ttxt('Type')]),
    el('select', { id: 'createNodeType', innerHTML: '<option value="faction">faction</option><option value="location">location</option><option value="npc">npc</option><option value="item">item</option><option value="quest_line">quest_line</option><option value="event">event</option>' }),
    el('input', { id: 'createNodeParent', type: 'hidden', value: parentId || '' }),
  ]);
  openModal(tr('Create Node'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalCreateNodeBtn' }, [ttxt('Create')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalCreateNodeBtn').addEventListener('click', createNode);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function createNode() {
  const name = document.getElementById('createNodeName').value.trim();
  const nodeType = document.getElementById('createNodeType').value;
  const parentId = document.getElementById('createNodeParent').value;
  if (!name) { toast(tr('Enter a node name'), 'error'); return; }
  try {
    var body = { name: name, node_type: nodeType, world_id: state.selectedWorldId };
    if (parentId) body.parent_id = parentId;
    await api('POST', '/api/v1/nodes', body);
    closeModal(); toast(tr('Node created'), 'success'); loadCurrentWorld();
  } catch(e) { showApiError('Failed', e); }
}

function openEditNodeModal(nodeId) {
  const n = state.nodes.find(function(x) { return x.id === nodeId; });
  if (!n) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'editNodeName' }, [ttxt('Node Name')]),
    el('input', { id: 'editNodeName', value: n.name, style: {width: '100%'} }),
    ce('label', { for: 'editNodeType' }, [ttxt('Node Type')]),
    el('select', { id: 'editNodeType', innerHTML: '<option value="faction">faction</option><option value="location">location</option><option value="npc">npc</option><option value="item">item</option><option value="quest_line">quest_line</option><option value="event">event</option>' }),
  ]);
  openModal(tr('Edit Node'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditNodeBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var et = document.getElementById('editNodeType'); if (et) et.value = n.node_type;
  document.getElementById('modalEditNodeBtn').addEventListener('click', function() { editNode(nodeId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function editNode(nodeId) {
  const name = document.getElementById('editNodeName').value.trim();
  const nodeType = document.getElementById('editNodeType').value;
  if (!name) { toast(tr('Enter a node name'), 'error'); return; }
  try {
    await api('PUT', '/api/v1/nodes/' + encodeURIComponent(nodeId), { name: name, node_type: nodeType });
    closeModal(); toast(tr('Node updated'), 'success'); loadCurrentWorld();
  } catch(e) { showApiError('Failed', e); }
}

async function moveNodeParent(nodeId, newParentId) {
  try {
    await api('PUT', '/api/v1/nodes/' + encodeURIComponent(nodeId), { parent_id: newParentId });
    toast(tr('Primary parent updated'), 'success'); loadCurrentWorld();
  } catch(e) { showApiError('Failed', e); }
}

function openCreateParentNodeModal(nodeId) {
  const n = state.nodes.find(function(x) { return x.id === nodeId; });
  if (!n) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '8px' } }, [ttxt('This action inserts a new primary parent node and rewires only parent_id. It does not create a relation row.')]),
    ce('label', { for: 'createParentNodeName' }, [ttxt('Node Name')]),
    el('input', { id: 'createParentNodeName', value: (n.name || '') + ' Parent', style: {width: '100%'} }),
    ce('label', { for: 'createParentNodeType' }, [ttxt('Type')]),
    el('select', { id: 'createParentNodeType', innerHTML: '<option value="faction">faction</option><option value="location">location</option><option value="npc">npc</option><option value="item">item</option><option value="quest_line">quest_line</option><option value="event">event</option>' }),
  ]);
  openModal(tr('Create Parent Node'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalCreateParentNodeBtn' }, [ttxt('Create')]), el('button', { id: 'modalCancelCreateParentNodeBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalCreateParentNodeBtn').addEventListener('click', function() { createParentNode(nodeId); });
  document.getElementById('modalCancelCreateParentNodeBtn').addEventListener('click', closeModal);
}

async function createParentNode(nodeId) {
  const n = state.nodes.find(function(x) { return x.id === nodeId; });
  if (!n) return;
  const name = document.getElementById('createParentNodeName').value.trim();
  const nodeType = document.getElementById('createParentNodeType').value;
  if (!name) { toast(tr('Enter a node name'), 'error'); return; }
  try {
    var parentBody = { name: name, node_type: nodeType, world_id: state.selectedWorldId };
    if (n.parent_id) parentBody.parent_id = n.parent_id;
    var created = await api('POST', '/api/v1/nodes', parentBody);
    await api('PUT', '/api/v1/nodes/' + encodeURIComponent(nodeId), { parent_id: created.id });
    closeModal();
    toast(tr('Parent node created'), 'success');
    await loadCurrentWorld();
    if (created && created.id) selectNode(created.id, created.node_type, { mode: 'single' });
  } catch(e) {
    showApiError('Failed', e);
  }
}

function openAddOutgoingRelationModal(nodeId) {
  var node = state.nodes.find(function(x) { return x.id === nodeId; });
  if (!node) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '8px' } }, [txt(relationSemanticsHint())]),
    ce('label', { for: 'addOutgoingRelSourceSearch' }, [ttxt('Search Nodes')]),
    el('input', { id: 'addOutgoingRelSourceSearch', value: node.name || '', style: {width: '100%'}, disabled: true }),
    ce('label', { for: 'addOutgoingRelSource' }, [ttxt('Source Node')]),
    el('select', { id: 'addOutgoingRelSource', innerHTML: '<option value="' + node.id + '">' + node.name + ' (' + node.node_type + ')</option>', disabled: true }),
    ce('label', { for: 'addOutgoingRelTargetSearch' }, [ttxt('Search Nodes')]),
    el('input', { id: 'addOutgoingRelTargetSearch', placeholder: tr('Target node search...'), style: {width: '100%'} }),
    ce('label', { for: 'addOutgoingRelTarget' }, [ttxt('Target Node')]),
    el('select', { id: 'addOutgoingRelTarget', innerHTML: '' }),
    ce('label', { for: 'addOutgoingRelType' }, [ttxt('Relation Type')]),
    el('select', { id: 'addOutgoingRelType', innerHTML: relationTypeOptionsHTML() }),
    ce('label', { for: 'addOutgoingRelWeight' }, [ttxt('Weight')]),
    el('input', { id: 'addOutgoingRelWeight', type: 'number', value: '1', min: '0', max: '100', style: {width: '100px'} }),
    ce('label', { for: 'addOutgoingRelProps' }, [ttxt('Relation Properties')]),
    el('textarea', { id: 'addOutgoingRelProps', rows: 5, placeholder: tr('Optional notes, tags, role metadata...'), style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'} }),
    ce('label', {}, [ttxt('Relation Meaning')]),
    ce('div', { id: 'addOutgoingRelMeaning', className: 'hint', style: {padding: '8px', textAlign: 'left'} }, [txt('')]),
    ce('label', {}, [ttxt('Modeling Warnings')]),
    ce('div', { id: 'addOutgoingRelWarnings', className: 'hint', style: {padding: '8px', textAlign: 'left', display: 'none', color: 'var(--yellow)'} }, [txt('')]),
    ce('label', {}, [ttxt('Relation Preview')]),
    ce('div', { id: 'addOutgoingRelPreview', className: 'status-box' }, [txt('')]),
  ]);
  openModal(tr('Add Outgoing Relation'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalAddOutgoingRelBtn' }, [ttxt('Add')]), el('button', { id: 'modalCancelAddOutgoingRelBtn', textContent: tr('Cancel') })])
  );
  bindRelationNodeFilter('addOutgoingRelTargetSearch', 'addOutgoingRelTarget', [nodeId], null);
  var sourceSelect = document.getElementById('addOutgoingRelSource');
  var targetSelect = document.getElementById('addOutgoingRelTarget');
  var typeSelect = document.getElementById('addOutgoingRelType');
  [sourceSelect, targetSelect, typeSelect].forEach(function(elm) {
    if (elm) elm.addEventListener('change', function() { updateRelationFormPreview('addOutgoing'); });
  });
  updateRelationFormPreview('addOutgoing');
  document.getElementById('modalAddOutgoingRelBtn').addEventListener('click', function() { addOutgoingRelation(nodeId); });
  document.getElementById('modalCancelAddOutgoingRelBtn').addEventListener('click', closeModal);
}

async function addOutgoingRelation(nodeId) {
  var node = state.nodes.find(function(x) { return x.id === nodeId; });
  var sourceId = document.getElementById('addOutgoingRelSource').value;
  var targetId = document.getElementById('addOutgoingRelTarget').value;
  var relType = document.getElementById('addOutgoingRelType').value;
  var weight = parseInt(document.getElementById('addOutgoingRelWeight').value, 10);
  var properties = document.getElementById('addOutgoingRelProps').value.trim();
  if (!sourceId) { toast(tr('Select a source node'), 'error'); return; }
  if (!targetId) { toast(tr('Select a target node'), 'error'); return; }
  if (sourceId === targetId) { toast(tr('Cannot link a node to itself'), 'error'); return; }
  if (!relType) { toast(tr('Select a relation type'), 'error'); return; }
  if (relType === 'external_parent' && node && node.parent_id === targetId) { toast(tr('Target node is already the primary parent'), 'error'); return; }
  if (relType === 'located_at') {
    var otherLocatedAt = (state.relations || []).some(function(rel) {
      return rel.source_id === sourceId && rel.relation_type === 'located_at' && rel.target_id !== targetId;
    });
    if (otherLocatedAt) { toast(tr('This node already has another located_at relation. Update the existing environment edge instead of stacking multiple active locations.'), 'error'); return; }
  }
  var exists = (state.relations || []).some(function(rel) {
    return rel.source_id === sourceId && rel.target_id === targetId && rel.relation_type === relType;
  });
  if (exists) { toast(tr('This relation already exists'), 'error'); return; }
  try {
    await api('POST', '/api/v1/relations', {
      world_id: state.selectedWorldId,
      source_id: sourceId,
      target_id: targetId,
      relation_type: relType,
      weight: Number.isFinite(weight) ? weight : 1,
      properties: properties,
    });
    closeModal();
    toast(tr('Relation added'), 'success');
    await selectNode(sourceId, node && node.node_type ? node.node_type : null, { mode: 'single' });
  } catch(e) {
    showApiError('Failed', e);
  }
}

function openCopyNodeModal(nodeId) {
  const n = state.nodes.find(function(x) { return x.id === nodeId; });
  if (!n) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'copyNodeName' }, [ttxt('Node Name')]),
    el('input', { id: 'copyNodeName', value: (n.name || '') + ' (copy)', style: {width: '100%'} }),
    ce('label', { className: 'checkbox-row' }, [el('input', { id: 'copyNodeWithChildren', type: 'checkbox', checked: true }), ttxt('Copy subtree')]),
  ]);
  openModal(tr('Copy Node'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalCopyNodeBtn' }, [ttxt('Create')]), el('button', { id: 'modalCancelCopyNodeBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalCopyNodeBtn').addEventListener('click', function() { copyNode(nodeId); });
  document.getElementById('modalCancelCopyNodeBtn').addEventListener('click', closeModal);
}

async function copyNode(nodeId) {
  const name = document.getElementById('copyNodeName').value.trim();
  const includeDescendants = document.getElementById('copyNodeWithChildren').checked;
  try {
    const result = await api('POST', '/api/v1/nodes/' + encodeURIComponent(nodeId) + '/copy', { name: name, include_descendants: includeDescendants });
    closeModal();
    toast(tr('Node copied'), 'success');
    await loadCurrentWorld();
    if (result && result.id) selectNode(result.id, result.node_type);
  } catch(e) {
    showApiError('Failed', e);
  }
}

async function deleteNodeHandler(nodeId) {
  const body = ce('p', { style: { color: 'var(--text)', fontSize: '12px' } }, [ttxt('Delete this node?')]);
  const footer = ce('div', {}, [
    ce('button', { className: 'danger', id: 'modalConfirmDelNodeBtn' }, [ttxt('Delete')]),
    el('button', { id: 'modalCancelDelNodeBtn', textContent: tr('Cancel') }),
  ]);
  openModal(tr('Confirm'), body, footer);
  document.getElementById('modalConfirmDelNodeBtn').addEventListener('click', async function() {
    closeModal();
    try {
      await api('DELETE', '/api/v1/nodes/' + encodeURIComponent(nodeId));
      toast(tr('Node deleted'), 'success');
      if (state.selectedNodeId === nodeId) { state.selectedNodeId = null; state.nodeDetail = null; }
      loadCurrentWorld();
    } catch(e) { showApiError('Failed', e); }
  });
  document.getElementById('modalCancelDelNodeBtn').addEventListener('click', closeModal);
}
/* ============= Component/Memory/Relation Add Modals ============= */
function openAddComponentModal() {
  if (!state.selectedNodeId) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'addCompType' }, [ttxt('Component Type')]),
    el('select', { id: 'addCompType', innerHTML: componentTypeOptionsHTML() }),
    ce('label', { for: 'addCompData', id: 'addCompRawLabel' }, [ttxt('Component Data (JSON/Markdown)')]),
    ce('div', { id: 'addCompEditorHost', className: 'component-editor-host' }, []),
    el('textarea', { id: 'addCompData', placeholder: tr('Enter component data...'), rows: 8, style: {width: '100%', fontFamily: 'var(--font-mono)'} }),
    ce('div', { id: 'addCompHint', className: 'hint', style: {textAlign: 'left'} }, [txt('')]),
  ]);
  openModal(tr('Add Component'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalAddCompBtn' }, [ttxt('Add')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('addCompType').addEventListener('change', function() { renderComponentEditor('addCompType', 'addCompData', 'addCompEditorHost', 'addCompRawLabel', 'addCompHint'); });
  renderComponentEditor('addCompType', 'addCompData', 'addCompEditorHost', 'addCompRawLabel', 'addCompHint');
  document.getElementById('modalAddCompBtn').addEventListener('click', addComponent);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function addComponent() {
  const compType = document.getElementById('addCompType').value;
  var editorData = collectComponentEditorData('addCompType', 'addCompData', 'addCompEditorHost', 'addCompRawLabel', 'addCompHint');
  if (editorData.error) { toast(editorData.error, 'error'); return; }
  const data = editorData.data.trim();
  var validationError = validateComponentEditorData(compType, data);
  if (validationError) { toast(validationError, 'error'); return; }
  try {
    await api('POST', '/api/v1/components', { node_id: state.selectedNodeId, component_type: compType, data: data });
    closeModal(); toast(tr('Component added'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { showApiError('Failed', e); }
}

function openAddMemoryModal() {
  if (!state.selectedNodeId) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'addMemContent' }, [ttxt('Content')]),
    el('textarea', { id: 'addMemContent', placeholder: tr('Enter memory content...'), rows: 6, style: {width: '100%'} }),
    ce('label', { for: 'addMemLevel' }, [ttxt('Level')]),
    el('select', { id: 'addMemLevel', innerHTML: '<option value="short_term">short_term</option><option value="long_term">long_term</option><option value="shared">shared</option><option value="world">world</option>' }),
    ce('label', { for: 'addMemTags' }, [ttxt('Tags (comma separated)')]),
    el('input', { id: 'addMemTags', placeholder: tr('tag1,tag2...'), style: {width: '100%'} }),
  ]);
  openModal(tr('Add Memory'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalAddMemBtn' }, [ttxt('Add')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalAddMemBtn').addEventListener('click', addMemory);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function addMemory() {
  const content = document.getElementById('addMemContent').value.trim();
  if (!content) { toast(tr('Enter memory content'), 'error'); return; }
  const level = document.getElementById('addMemLevel').value;
  const tags = document.getElementById('addMemTags').value.trim();
  try {
    await api('POST', '/api/v1/memories', { node_id: state.selectedNodeId, content: content, level: level, tags: tags });
    closeModal(); toast(tr('Memory added'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { showApiError('Failed', e); }
}

function openAddRelationModal() {
  if (!state.selectedNodeId) return;
  var sourceNode = state.nodes.find(function(n) { return n.id === state.selectedNodeId; });
  const f = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '8px' } }, [txt(relationSemanticsHint())]),
    ce('label', { for: 'addRelSourceSearch' }, [ttxt('Search Nodes')]),
    el('input', { id: 'addRelSourceSearch', placeholder: tr('Source node search...'), value: sourceNode ? sourceNode.name : '', style: {width: '100%'} }),
    ce('label', { for: 'addRelSource' }, [ttxt('Source Node')]),
    el('select', { id: 'addRelSource', innerHTML: '' }),
    ce('label', { for: 'addRelTargetSearch' }, [ttxt('Search Nodes')]),
    el('input', { id: 'addRelTargetSearch', placeholder: tr('Target node search...'), style: {width: '100%'} }),
    ce('label', { for: 'addRelTarget' }, [ttxt('Target Node')]),
    el('select', { id: 'addRelTarget', innerHTML: '' }),
    ce('label', { for: 'addRelType' }, [ttxt('Relation Type')]),
    el('select', { id: 'addRelType', innerHTML: relationTypeOptionsHTML() }),
    ce('label', { for: 'addRelWeight' }, [ttxt('Weight')]),
    el('input', { id: 'addRelWeight', type: 'number', value: '5', min: '0', max: '100', style: {width: '100px'} }),
    ce('label', { for: 'addRelProps' }, [ttxt('Relation Properties')]),
    el('textarea', { id: 'addRelProps', rows: 5, placeholder: tr('Optional notes, tags, role metadata...'), style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'} }),
    ce('label', {}, [ttxt('Relation Meaning')]),
    ce('div', { id: 'addRelMeaning', className: 'hint', style: {padding: '8px', textAlign: 'left'} }, [txt('')]),
    ce('label', {}, [ttxt('Modeling Warnings')]),
    ce('div', { id: 'addRelWarnings', className: 'hint', style: {padding: '8px', textAlign: 'left', display: 'none', color: 'var(--yellow)'} }, [txt('')]),
    ce('label', {}, [ttxt('Relation Preview')]),
    ce('div', { id: 'addRelPreview', className: 'status-box' }, [txt('')]),
  ]);
  openModal(tr('Add Relation'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalAddRelBtn' }, [ttxt('Add')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  bindRelationNodeFilter('addRelSourceSearch', 'addRelSource', [], state.selectedNodeId);
  bindRelationNodeFilter('addRelTargetSearch', 'addRelTarget', [state.selectedNodeId], null);
  var sourceSelect = document.getElementById('addRelSource');
  var targetSelect = document.getElementById('addRelTarget');
  var typeSelect = document.getElementById('addRelType');
  if (sourceSelect) sourceSelect.value = state.selectedNodeId;
  [sourceSelect, targetSelect, typeSelect].forEach(function(elm) {
    if (elm) elm.addEventListener('change', function() { updateRelationFormPreview('add'); });
  });
  updateRelationFormPreview('add');
  document.getElementById('modalAddRelBtn').addEventListener('click', addRelation);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function addRelation() {
  const sourceId = document.getElementById('addRelSource').value;
  const targetId = document.getElementById('addRelTarget').value;
  const relType = document.getElementById('addRelType').value;
  const weight = parseInt(document.getElementById('addRelWeight').value) || 5;
  const properties = document.getElementById('addRelProps').value.trim();
  if (!sourceId) { toast(tr('Select a source node'), 'error'); return; }
  if (!targetId) { toast(tr('Select a target node'), 'error'); return; }
  if (sourceId === targetId) { toast(tr('Cannot link a node to itself'), 'error'); return; }
  if (!relType) { toast(tr('Select a relation type'), 'error'); return; }
  var sourceNode = getNodeById(sourceId);
  if (relType === 'external_parent' && sourceNode && sourceNode.parent_id === targetId) { toast(tr('Target node is already the primary parent'), 'error'); return; }
  if (relType === 'located_at') {
    var existingLocatedAt = (state.relations || []).some(function(rel) {
      return rel.source_id === sourceId && rel.relation_type === 'located_at' && rel.target_id !== targetId;
    });
    if (existingLocatedAt) { toast(tr('This node already has another located_at relation. Update the existing environment edge instead of stacking multiple active locations.'), 'error'); return; }
  }
  var duplicate = (state.relations || []).some(function(rel) {
    return rel.source_id === sourceId && rel.target_id === targetId && rel.relation_type === relType;
  });
  if (duplicate) { toast(tr('This relation already exists'), 'error'); return; }
  try {
    await api('POST', '/api/v1/relations', { world_id: state.selectedWorldId, source_id: sourceId, target_id: targetId, relation_type: relType, weight: weight, properties: properties });
    closeModal(); toast(tr('Relation added'), 'success'); loadCurrentWorld(); selectNode(sourceId);
  } catch(e) { showApiError('Failed', e); }
}

/* ============= Import Modal ============= */
function openImportModal() {
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'importFormat' }, [ttxt('Format')]),
    el('select', { id: 'importFormat', innerHTML: '<option value="yaml">YAML</option><option value="json">JSON</option>' }),
    ce('label', { for: 'importContent' }, [ttxt('Content')]),
    el('textarea', { id: 'importContent', placeholder: tr('Paste YAML/JSON content...'), rows: 15, style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'} }),
    ce('div', { className: 'import-checks' }, [
      ce('label', { className: 'checkbox-row' }, [el('input', { type: 'checkbox', id: 'importDryRun' }), ttxt('Dry-run')]),
      ce('label', { className: 'checkbox-row' }, [el('input', { type: 'checkbox', id: 'importReset' }), ttxt('Reset World')]),
    ]),
  ]);
  openModal(tr('Import Config'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalImportBtn' }, [ttxt('Import')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalImportBtn').addEventListener('click', importConfig);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function importConfig() {
  const format = document.getElementById('importFormat').value;
  const content = document.getElementById('importContent').value.trim();
  const dryRun = document.getElementById('importDryRun').checked;
  const reset = document.getElementById('importReset').checked;
  if (!content) { toast(tr('Enter content'), 'error'); return; }
  try {
    const res = await api('POST', '/api/v1/creator/import', { format: format, content: content, dry_run: dryRun, reset: reset });
    closeModal();
    toast(tr('Import successful'), 'success');
    if (!dryRun) loadWorlds();
  } catch(e) { showApiError('Failed', e); }
}

/* ============= Config Modal ============= */
function openConfigModal() {
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'cfgUrl' }, [ttxt('Server URL')]),
    el('input', { id: 'cfgUrl', value: cfg.url, style: {width: '100%'} }),
    ce('label', { for: 'cfgKey' }, [ttxt('API Key')]),
    el('input', { id: 'cfgKey', value: cfg.key, style: {width: '100%'} }),
  ]);
  openModal(tr('Server Config'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalSaveCfgBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalSaveCfgBtn').addEventListener('click', saveConfig);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

function saveConfig() {
  cfg.url = document.getElementById('cfgUrl').value.trim();
  cfg.key = document.getElementById('cfgKey').value.trim();
  saveCfg(cfg); closeModal(); toast(tr('Config saved'), 'success');
  checkHealth(); loadWorlds();
}

/* ============= Basic Actions ============= */
async function tickAdvance() { if (!requireWorldGuard()) return;
  if (!state.selectedWorldId) { toast(tr('Please select a world first'), 'error'); return; }
  const f = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '8px' } }, [ttxt('requested_ticks is required for flexible mode and must stay 1 in fixed mode.')]),
    ce('label', { for: 'tickAdvanceType' }, [ttxt('Tick Type')]),
    el('input', { id: 'tickAdvanceType', value: 'scheduled', style: { width: '100%' } }),
    ce('label', { for: 'tickAdvanceGameTime' }, [ttxt('Game Time')]),
    el('input', { id: 'tickAdvanceGameTime', value: '', placeholder: tr('Optional external time label'), style: { width: '100%' } }),
    ce('label', { for: 'tickAdvanceRequestedTicks' }, [ttxt('Requested Ticks')]),
    el('input', { id: 'tickAdvanceRequestedTicks', type: 'number', min: '1', value: '1', style: { width: '120px' } }),
    ce('label', { for: 'tickAdvanceAutonomousLimit' }, [ttxt('Autonomous Limit')]),
    el('input', { id: 'tickAdvanceAutonomousLimit', type: 'number', min: '0', value: '', placeholder: tr('Optional'), style: { width: '120px' } }),
  ]);
  openModal(tr('Advance Tick'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalTickAdvanceBtn' }, [ttxt('Advance Tick')]), el('button', { id: 'modalCancelTickAdvanceBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalTickAdvanceBtn').addEventListener('click', submitTickAdvance);
  document.getElementById('modalCancelTickAdvanceBtn').addEventListener('click', closeModal);
}

async function submitTickAdvance() {
  var tickType = document.getElementById('tickAdvanceType').value.trim() || 'scheduled';
  var gameTime = document.getElementById('tickAdvanceGameTime').value.trim();
  var requestedTicksRaw = document.getElementById('tickAdvanceRequestedTicks').value.trim();
  var autonomousLimitRaw = document.getElementById('tickAdvanceAutonomousLimit').value.trim();
  var requestedTicks = requestedTicksRaw ? parseInt(requestedTicksRaw, 10) : 1;
  var autonomousLimit = autonomousLimitRaw ? parseInt(autonomousLimitRaw, 10) : null;
  if (!Number.isFinite(requestedTicks) || requestedTicks <= 0) {
    toast(tr('Requested ticks must be greater than 0'), 'error');
    return;
  }
  var worldTimeSettings = state.settings && state.settings.world_time_settings ? state.settings.world_time_settings : null;
  if (worldTimeSettings && worldTimeSettings.tick_scale_mode === 'fixed' && requestedTicks !== 1) {
    toast(tr('Fixed tick scale mode only allows requested_ticks = 1'), 'error');
    return;
  }
  showLoading(tr('Advancing tick...'));
  try {
    const body = { tick_type: tickType, game_time: gameTime, requested_ticks: requestedTicks };
    if (Number.isFinite(autonomousLimit) && autonomousLimit >= 0) body.autonomous_limit = autonomousLimit;
    const res = await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/ticks/advance', body);
    hideLoading();
    closeModal();
    toast(tr('Tick advanced') + ': ' + (res.tick ? 'tick ' + res.tick.tick_number : tr('Valid')), 'success');
    await loadTimelines();
    await loadStateComponents();
    await loadContinuityOverview();
    var worldTimeState = res.world_time_state || null;
    var autonomousRuns = Array.isArray(res.autonomous_runs) ? res.autonomous_runs : [];
    var resultEl = ce('div', { className: 'modal-field' }, [
      ce('label', {}, [txt(tr('Tick Advance Result'))]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Tick')]), ce('span', { className: 'val' }, [txt(res.tick ? String(res.tick.tick_number || 0) : '-')])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Requested Ticks')]), ce('span', { className: 'val' }, [txt(String(requestedTicks))])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Advanced Ticks')]), ce('span', { className: 'val' }, [txt(String(res.advanced_ticks || 0))])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('World Time Label')]), ce('span', { className: 'val' }, [txt(worldTimeState && worldTimeState.current_time_label ? worldTimeState.current_time_label : '-')])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Autonomous Runs')]), ce('span', { className: 'val' }, [txt(String(autonomousRuns.length))])]),
      el('pre', { style: {fontSize: '11px', whiteSpace: 'pre-wrap', maxHeight: '320px', overflow: 'auto', background: 'var(--bg-input)', padding: '8px', borderRadius: 'var(--radius)'}, textContent: JSON.stringify(res, null, 2) }),
    ]);
    openModal(tr('Tick Advance Result'), resultEl, ce('div', {}, [el('button', { id: 'modalCloseTickAdvanceResultBtn', textContent: tr('Close') })]));
    document.getElementById('modalCloseTickAdvanceResultBtn').addEventListener('click', closeModal);
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

async function runAutonomous() { if (!requireBothGuard()) return;
  if (!state.selectedWorldId) { toast(tr('Please select a world first'), 'error'); return; } if (!state.selectedNodeId) { toast(tr('Please select a node first'), 'error'); return; }
  showLoading(tr('Running autonomous...'));
  try {
    await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/nodes/' + encodeURIComponent(state.selectedNodeId) + '/autonomous/run', null);
    hideLoading();
    toast(tr('Autonomous triggered'), 'success');
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

async function savePolicy() {
  const blocked = document.getElementById('policyBlocked').value.split('\n').map(function(s) { return s.trim(); }).filter(Boolean);
  const safe = document.getElementById('policySafe').value.split('\n').map(function(s) { return s.trim(); }).filter(Boolean);
  try {
    await api('PUT', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/policy', { blocked_actions: blocked, safe_actions: safe });
    toast(tr('Policy saved'), 'success'); loadPolicy();
  } catch(e) { showApiError('Failed', e); }
}

async function saveSettings() {
  try {
    var current = state.settings || {};
    var payload = {};

    function maybeSetInt(key, elementId) {
      var raw = document.getElementById(elementId).value.trim();
      if (raw === '') return;
      var next = parseInt(raw, 10);
      if (!Number.isFinite(next)) return;
      if (current[key] !== next) payload[key] = next;
    }

    function maybeSetBool(key, elementId) {
      var next = !!document.getElementById(elementId).checked;
      if (!!current[key] !== next) payload[key] = next;
    }

    function maybeSetString(key, elementId, fallback) {
      var raw = document.getElementById(elementId).value.trim();
      var next = raw || fallback;
      if ((current[key] || '') !== next) payload[key] = next;
    }

    maybeSetInt('memory_limit', 'setMemoryLimit');
    maybeSetInt('max_analysis_rounds', 'setMaxRounds');
    maybeSetInt('max_context_depth', 'setMaxDepth');
    maybeSetBool('auto_apply', 'setAutoApply');
    maybeSetString('require_review_above', 'setReviewAbove', 'critical');
    maybeSetString('pipeline_mode', 'setPipelineMode', 'full');
    maybeSetInt('propagation_max_depth', 'setPropMaxDepth');
    maybeSetInt('sub_task_max_retries', 'setSubTaskRetries');
    maybeSetInt('sub_task_timeout_secs', 'setSubTaskTimeout');
    maybeSetBool('enable_propagation_machine', 'setEnablePropMachine');

    var tickScaleModeEl = document.getElementById('setTickScaleMode');
    if (tickScaleModeEl) {
      var worldTimeSettings = collectWorldTimeSettingsFromForm();
      var worldTimeValidationError = validateWorldTimeSettingsInput(worldTimeSettings);
      if (worldTimeValidationError) {
        toast(worldTimeValidationError, 'error');
        return;
      }
      if (JSON.stringify(current.world_time_settings || null) !== JSON.stringify(worldTimeSettings)) {
        payload.world_time_settings = worldTimeSettings;
      }
    }

    if (Object.keys(payload).length === 0) {
      toast(tr('Settings saved'), 'success');
      return;
    }

    await api('PUT', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/settings', payload);
    toast(tr('Settings saved'), 'success'); loadSettings();
  } catch(e) { showApiError('Failed', e); }
}

async function deleteComponent(compId) {
  const body = ce('p', { style: { color: 'var(--text)', fontSize: '12px' } }, [ttxt('Delete this component?')]);
  const footer = ce('div', {}, [
    ce('button', { className: 'danger', id: 'modalConfirmDelBtn' }, [ttxt('Delete')]),
    el('button', { id: 'modalCancelDelBtn', textContent: tr('Cancel') }),
  ]);
  openModal(tr('Confirm'), body, footer);
  document.getElementById('modalConfirmDelBtn').addEventListener('click', async function() {
    closeModal();
    try {
      await api('DELETE', '/api/v1/components/' + encodeURIComponent(compId));
      toast(tr('Component deleted'), 'success'); selectNode(state.selectedNodeId);
    } catch(e) { showApiError('Failed', e); }
  });
  document.getElementById('modalCancelDelBtn').addEventListener('click', closeModal);
}

async function deleteMemory(memId) {
  const body = ce('p', { style: { color: 'var(--text)', fontSize: '12px' } }, [ttxt('Delete this memory?')]);
  const footer = ce('div', {}, [
    ce('button', { className: 'danger', id: 'modalConfirmDelBtn' }, [ttxt('Delete')]),
    el('button', { id: 'modalCancelDelBtn', textContent: tr('Cancel') }),
  ]);
  openModal(tr('Confirm'), body, footer);
  document.getElementById('modalConfirmDelBtn').addEventListener('click', async function() {
    closeModal();
    try {
      await api('DELETE', '/api/v1/memories/' + encodeURIComponent(memId));
      toast(tr('Memory deleted'), 'success'); selectNode(state.selectedNodeId);
    } catch(e) { showApiError('Failed', e); }
  });
  document.getElementById('modalCancelDelBtn').addEventListener('click', closeModal);
}

async function deleteRelation(relId) {
  const body = ce('p', { style: { color: 'var(--text)', fontSize: '12px' } }, [ttxt('Delete this relation?')]);
  const footer = ce('div', {}, [
    ce('button', { className: 'danger', id: 'modalConfirmDelBtn' }, [ttxt('Delete')]),
    el('button', { id: 'modalCancelDelBtn', textContent: tr('Cancel') }),
  ]);
  openModal(tr('Confirm'), body, footer);
  document.getElementById('modalConfirmDelBtn').addEventListener('click', async function() {
    closeModal();
    try {
      await api('DELETE', '/api/v1/relations/' + encodeURIComponent(relId));
      toast(tr('Relation deleted'), 'success'); selectNode(state.selectedNodeId);
    } catch(e) { showApiError('Failed', e); }
  });
  document.getElementById('modalCancelDelBtn').addEventListener('click', closeModal);
}

/* ============= Edit Modals ============= */
function openEditComponentModal(compId) {
  const nd = state.nodeDetail;
  if (!nd || !nd.components) return;
  const comp = nd.components.find(function(x) { return x.id === compId; });
  if (!comp) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'editCompType' }, [ttxt('Component Type')]),
    el('select', { id: 'editCompType', innerHTML: componentTypeOptionsHTML() }),
    ce('label', { for: 'editCompData', id: 'editCompRawLabel' }, [ttxt('Component Data')]),
    ce('div', { id: 'editCompEditorHost', className: 'component-editor-host' }, []),
    el('textarea', { id: 'editCompData', rows: 10, style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'}, textContent: comp.data || '' }),
    ce('div', { id: 'editCompHint', className: 'hint', style: {textAlign: 'left'} }, [txt('')]),
  ]);
  openModal(tr('Edit Component'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditCompBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var ec = document.getElementById('editCompType'); if (ec) ec.value = comp.component_type;
  if (ec) ec.addEventListener('change', function() { renderComponentEditor('editCompType', 'editCompData', 'editCompEditorHost', 'editCompRawLabel', 'editCompHint'); });
  renderComponentEditor('editCompType', 'editCompData', 'editCompEditorHost', 'editCompRawLabel', 'editCompHint');
  document.getElementById('modalEditCompBtn').addEventListener('click', function() { editComponent(compId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function editComponent(compId) {
  const compType = document.getElementById('editCompType').value;
  var editorData = collectComponentEditorData('editCompType', 'editCompData', 'editCompEditorHost', 'editCompRawLabel', 'editCompHint');
  if (editorData.error) { toast(editorData.error, 'error'); return; }
  const data = editorData.data.trim();
  var validationError = validateComponentEditorData(compType, data);
  if (validationError) { toast(validationError, 'error'); return; }
  try {
    await api('PUT', '/api/v1/components/' + encodeURIComponent(compId), { component_type: compType, data: data });
    closeModal(); toast(tr('Component updated'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { showApiError('Failed', e); }
}

function openEditMemoryModal(memId) {
  const nd = state.nodeDetail;
  if (!nd || !nd.memories) return;
  const mem = nd.memories.find(function(x) { return x.id === memId; });
  if (!mem) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'editMemContent' }, [ttxt('Content')]),
    el('textarea', { id: 'editMemContent', rows: 6, style: {width: '100%'}, textContent: mem.content || '' }),
    ce('label', { for: 'editMemLevel' }, [ttxt('Level')]),
    el('select', { id: 'editMemLevel', innerHTML: '<option value="short_term">short_term</option><option value="long_term">long_term</option><option value="shared">shared</option><option value="world">world</option>' }),
    ce('label', { for: 'editMemTags' }, [ttxt('Tags')]),
    el('input', { id: 'editMemTags', value: mem.tags || '', style: {width: '100%'} }),
  ]);
  openModal(tr('Edit Memory'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditMemBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var em = document.getElementById('editMemLevel'); if (em) em.value = mem.level || 'long_term';
  document.getElementById('modalEditMemBtn').addEventListener('click', function() { editMemory(memId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function editMemory(memId) {
  const content = document.getElementById('editMemContent').value.trim();
  const level = document.getElementById('editMemLevel').value;
  const tags = document.getElementById('editMemTags').value.trim();
  if (!content) { toast(tr('Enter memory content'), 'error'); return; }
  try {
    await api('PUT', '/api/v1/memories/' + encodeURIComponent(memId), { content: content, level: level, tags: tags });
    closeModal(); toast(tr('Memory updated'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { showApiError('Failed', e); }
}

function openPropagateMemoryModal(memId) {
  const nd = state.nodeDetail;
  if (!nd || !nd.memories) return;
  const mem = nd.memories.find(function(x) { return x.id === memId; });
  if (!mem) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'hint', style: { textAlign: 'left' } }, [txt((mem.content || '').slice(0, 120) + ((mem.content || '').length > 120 ? '...' : ''))]),
    ce('label', { for: 'propMemMode' }, [ttxt('Propagation Mode')]),
    el('select', { id: 'propMemMode', innerHTML: '<option value="upward">upward</option><option value="environment_scope">environment_scope</option><option value="organization_scope">organization_scope</option><option value="tag_broadcast">tag_broadcast</option><option value="targeted">targeted</option><option value="manual">manual</option>' }),
    ce('div', { id: 'propMemModeMeaning', className: 'hint', style: { padding: '8px', textAlign: 'left' } }, [txt('')]),
    ce('label', { for: 'propMemTags' }, [ttxt('Propagation Tags')]),
    el('input', { id: 'propMemTags', placeholder: tr('Tags, comma separated'), style: {width: '100%'} }),
    ce('label', { for: 'propMemTargets' }, [ttxt('Propagation Targets (node IDs)')]),
    el('input', { id: 'propMemTargets', placeholder: tr('Target node IDs, comma separated'), style: {width: '100%'} }),
    ce('label', { for: 'propMemDepth' }, [ttxt('Propagation Max Depth')]),
    el('input', { id: 'propMemDepth', type: 'number', value: '0', min: '0', style: {width: '100px'} }),
    ce('label', { className: 'checkbox-row' }, [el('input', { id: 'propMemPublishUp', type: 'checkbox' }), ttxt('Publish Up')]),
    ce('div', { className: 'hint', style: { textAlign: 'left' } }, [txt(tr('Publish Up only extends environment_scope or organization_scope beyond their scoped graph. It has no effect on targeted or manual propagation.'))]),
  ]);
  openModal(tr('Propagate Memory'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalPropMemBtn' }, [ttxt('Propagate')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var modeEl = document.getElementById('propMemMode');
  if (modeEl) modeEl.addEventListener('change', updatePropagationModePreview);
  updatePropagationModePreview();
  document.getElementById('modalPropMemBtn').addEventListener('click', function() { propagateMemory(memId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function propagateMemory(memId) {
  const mode = document.getElementById('propMemMode').value;
  const tags = document.getElementById('propMemTags').value.trim().split(',').map(function(s) { return s.trim(); }).filter(Boolean);
  const targets = document.getElementById('propMemTargets').value.trim().split(',').map(function(s) { return s.trim(); }).filter(Boolean);
  const maxDepth = parseInt(document.getElementById('propMemDepth').value, 10) || 0;
  const publishUp = document.getElementById('propMemPublishUp').checked;
  if (validPropagationModes().indexOf(mode) < 0) { toast(tr('Unsupported propagation mode'), 'error'); return; }
  if (mode === 'targeted' && targets.length === 0) { toast(tr('Targeted propagation requires at least one target node ID'), 'error'); return; }
  if (mode === 'tag_broadcast' && tags.length === 0) { toast(tr('Tag broadcast propagation requires at least one tag'), 'error'); return; }
  try {
    await api('POST', '/api/v1/memories/propagate', {
      memory_id: memId,
      mode: mode,
      tags: tags,
      target_ids: targets,
      max_depth: maxDepth,
      publish_up: publishUp,
    });
    closeModal();
    toast(tr('Propagation request submitted'), 'success');
  } catch(e) {
    showApiError('Failed', e);
  }
}

async function loadStateComponents() {
  if (!state.selectedWorldId) {
    state.stateComponents = [];
    if (state.page === 'state') renderCurrent();
    return;
  }
  try {
    var result = await api('GET', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/state-components');
    state.stateComponents = result.components || [];
    if (state.page === 'state') renderCurrent();
  } catch(e) {
    state.stateComponents = [];
    if (state.page === 'state') renderCurrent();
    showApiError('Failed to load state components', e);
  }
}

async function loadTimelines() {
  if (!state.selectedWorldId) {
    state.timelines = [];
    if (state.page === 'timelines') renderCurrent();
    return;
  }
  try {
    var result = await api('GET', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/timelines?limit=50');
    state.timelines = result.timelines || [];
    if (state.page === 'timelines') renderCurrent();
  } catch(e) {
    state.timelines = [];
    if (state.page === 'timelines') renderCurrent();
    showApiError('Failed to load timelines', e);
  }
}

function openEditStateComponentModal(componentType, payload) {
  var text = payload ? JSON.stringify(payload, null, 2) : '{}';
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'editStateType' }, [ttxt('State Component')]),
    el('input', { id: 'editStateType', value: componentType, disabled: 'disabled', style: { width: '100%' } }),
    ce('label', { for: 'editStateData', id: 'editStateRawLabel' }, [ttxt('Component Data')]),
    ce('div', { id: 'editStateEditorHost', className: 'component-editor-host' }, []),
    el('textarea', { id: 'editStateData', rows: 14, style: { width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px' }, textContent: text }),
    ce('div', { id: 'editStateHint', className: 'hint', style: { textAlign: 'left' } }, [txt('')]),
  ]);
  openModal(tr('Edit State Component'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalSaveStateBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  renderComponentEditor('editStateType', 'editStateData', 'editStateEditorHost', 'editStateRawLabel', 'editStateHint');
  var validationError = validateComponentEditorData(componentType, text);
  if (!validationError) setComponentEditorHintText(getComponentEditorContext('editStateType', 'editStateData', 'editStateEditorHost', 'editStateRawLabel', 'editStateHint'), componentType, tr('Structured world tick continuity state.'));
  document.getElementById('modalSaveStateBtn').addEventListener('click', function() { saveStateComponent(componentType); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function saveStateComponent(componentType) {
  var editorData = collectComponentEditorData('editStateType', 'editStateData', 'editStateEditorHost', 'editStateRawLabel', 'editStateHint');
  if (editorData.error) { toast(editorData.error, 'error'); return; }
  const data = editorData.data.trim();
  var validationError = validateComponentEditorData(componentType, data);
  if (validationError) { toast(validationError, 'error'); return; }
  try {
    await api('PUT', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/state-components/' + encodeURIComponent(componentType), JSON.parse(data));
    closeModal();
    toast(tr('State component saved'), 'success');
    loadStateComponents();
  } catch(e) {
    showApiError('Failed', e);
  }
}

function openEditRelationModal(relId) {
  const nd = state.nodeDetail;
  if (!nd || !nd.relations) return;
  const rel = nd.relations.find(function(x) { return x.id === relId; });
  if (!rel) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '8px' } }, [txt(relationSemanticsHint())]),
    ce('label', { for: 'editRelSourceSearch' }, [ttxt('Search Nodes')]),
    el('input', { id: 'editRelSourceSearch', placeholder: tr('Source node search...'), value: getNodeNameById(rel.source_id), style: {width: '100%'} }),
    ce('label', { for: 'editRelSource' }, [ttxt('Source Node')]),
    el('select', { id: 'editRelSource', innerHTML: '' }),
    ce('label', { for: 'editRelTargetSearch' }, [ttxt('Search Nodes')]),
    el('input', { id: 'editRelTargetSearch', placeholder: tr('Target node search...'), value: getNodeNameById(rel.target_id), style: {width: '100%'} }),
    ce('label', { for: 'editRelTarget' }, [ttxt('Target Node')]),
    el('select', { id: 'editRelTarget', innerHTML: '' }),
    ce('label', { for: 'editRelType' }, [ttxt('Relation Type')]),
    el('select', { id: 'editRelType', innerHTML: relationTypeOptionsHTML() }),
    ce('label', { for: 'editRelWeight' }, [ttxt('Weight')]),
    el('input', { id: 'editRelWeight', type: 'number', value: String(rel.weight || 5), min: '0', max: '100', style: {width: '100px'} }),
    ce('label', { for: 'editRelProps' }, [ttxt('Relation Properties')]),
    el('textarea', { id: 'editRelProps', rows: 5, placeholder: tr('Optional notes, tags, role metadata...'), style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'}, textContent: rel.properties || '' }),
    ce('label', {}, [ttxt('Relation Meaning')]),
    ce('div', { id: 'editRelMeaning', className: 'hint', style: {padding: '8px', textAlign: 'left'} }, [txt('')]),
    ce('label', {}, [ttxt('Modeling Warnings')]),
    ce('div', { id: 'editRelWarnings', className: 'hint', style: {padding: '8px', textAlign: 'left', display: 'none', color: 'var(--yellow)'} }, [txt('')]),
    ce('label', {}, [ttxt('Relation Preview')]),
    ce('div', { id: 'editRelPreview', className: 'status-box' }, [txt('')]),
  ]);
  openModal(tr('Edit Relation'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditRelBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  bindRelationNodeFilter('editRelSourceSearch', 'editRelSource', [], rel.source_id);
  bindRelationNodeFilter('editRelTargetSearch', 'editRelTarget', [], rel.target_id);
  var ers = document.getElementById('editRelSource'); if (ers) ers.value = rel.source_id;
  var ert = document.getElementById('editRelTarget'); if (ert) ert.value = rel.target_id;
  var ert2 = document.getElementById('editRelType'); if (ert2) ert2.value = rel.relation_type;
  [ers, ert, ert2].forEach(function(elm) {
    if (elm) elm.addEventListener('change', function() { updateRelationFormPreview('edit'); });
  });
  updateRelationFormPreview('edit');
  document.getElementById('modalEditRelBtn').addEventListener('click', function() { editRelation(relId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function editRelation(relId) {
  const sourceId = document.getElementById('editRelSource').value;
  const targetId = document.getElementById('editRelTarget').value;
  const relType = document.getElementById('editRelType').value;
  const weight = parseInt(document.getElementById('editRelWeight').value) || 5;
  const properties = document.getElementById('editRelProps').value.trim();
  if (!sourceId || !targetId) { toast(tr('Select source and target'), 'error'); return; }
  if (sourceId === targetId) { toast(tr('Cannot link a node to itself'), 'error'); return; }
  if (!relType) { toast(tr('Select a relation type'), 'error'); return; }
  var sourceNode = getNodeById(sourceId);
  if (relType === 'external_parent' && sourceNode && sourceNode.parent_id === targetId) { toast(tr('Target node is already the primary parent'), 'error'); return; }
  if (relType === 'located_at') {
    var existingLocatedAt = (state.relations || []).some(function(rel) {
      return rel.id !== relId && rel.source_id === sourceId && rel.relation_type === 'located_at' && rel.target_id !== targetId;
    });
    if (existingLocatedAt) { toast(tr('This node already has another located_at relation. Update the existing environment edge instead of stacking multiple active locations.'), 'error'); return; }
  }
  var duplicate = (state.relations || []).some(function(rel) {
    return rel.id !== relId && rel.source_id === sourceId && rel.target_id === targetId && rel.relation_type === relType;
  });
  if (duplicate) { toast(tr('This relation already exists'), 'error'); return; }
  try {
    await api('PUT', '/api/v1/relations/' + encodeURIComponent(relId), { source_id: sourceId, target_id: targetId, relation_type: relType, weight: weight, properties: properties });
    closeModal(); toast(tr('Relation updated'), 'success'); loadCurrentWorld(); selectNode(sourceId);
  } catch(e) { showApiError('Failed', e); }
}

/* ============= Event / Scope / Replan ============= */
function openEventImpactModal() { if (!requireWorldGuard()) return;
  if (!state.selectedWorldId) { toast(tr('Please select a world first'), 'error'); return; }
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'eiType' }, [ttxt('Event Type')]),
    el('input', { id: 'eiType', placeholder: tr('diplomatic_shift, natural_disaster...'), style: {width: '100%'} }),
    ce('label', { for: 'eiScope' }, [ttxt('Scope (node ID, optional)')]),
    el('input', { id: 'eiScope', placeholder: tr('Optional'), style: {width: '100%'} }),
    ce('label', { for: 'eiDesc' }, [ttxt('Description')]),
    el('textarea', { id: 'eiDesc', placeholder: tr('Describe the event...'), rows: 4, style: {width: '100%'} }),
    ce('label', { for: 'eiSeverity' }, [ttxt('Severity')]),
    el('select', { id: 'eiSeverity', innerHTML: '<option value="low">low</option><option value="medium">medium</option><option value="high">high</option><option value="critical">critical</option>' }),
  ]);
  openModal(tr('Event Impact Assessment'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEiBtn' }, [ttxt('Assess')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalEiBtn').addEventListener('click', eventImpact);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function eventImpact() {
  const eventType = document.getElementById('eiType').value.trim();
  const scopeId = document.getElementById('eiScope').value.trim();
  const description = document.getElementById('eiDesc').value.trim();
  const severity = document.getElementById('eiSeverity').value;
  if (!eventType || !description) { toast(tr('Enter event type and description'), 'error'); return; }
  showLoading(tr('Assessing event impact...'));
  try {
    const res = await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/events/impact', { event_type: eventType, scope_id: scopeId || undefined, description: description, severity: severity });
    hideLoading(); closeModal();
    const resultEl = ce('div', { className: 'modal-field' }, [
      ce('label', {}, [txt('Impact: ' + (res.world_change_plan ? res.world_change_plan.impact_level : '-'))]),
      ce('label', {}, [txt('Summary: ' + (res.world_change_plan ? res.world_change_plan.summary : ''))]),
      el('pre', { style: {fontSize: '11px', whiteSpace: 'pre-wrap', maxHeight: '200px', overflow: 'auto', background: 'var(--bg-input)', padding: '8px', borderRadius: 'var(--radius)'}, textContent: JSON.stringify(res, null, 2) }),
    ]);
    toast(tr('Event assessed'), 'success');
    openModal(tr('Assessment Result'), resultEl, ce('div', {}, [el('button', { id: 'modalCloseResultBtn', textContent: tr('Close') })]));
    document.getElementById('modalCloseResultBtn').addEventListener('click', closeModal);
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

async function scopeAdvance() { if (!requireBothGuard()) return; if (state.selectedNodeType === 'world') { toast(tr('Scope Advance requires a non-world node'), 'error'); return; }
  if (!state.selectedNodeId) { toast(tr('Select a node as scope'), 'error'); return; }
  if (!state.selectedWorldId) { toast(tr('Select a world'), 'error'); return; }
  showLoading(tr('Advancing scope...'));
  try {
    await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/scopes/' + encodeURIComponent(state.selectedNodeId) + '/advance', null);
    hideLoading();
    toast(tr('Scope advanced'), 'success');
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

async function timelineReplan() { if (!requireWorldGuard()) return;
  if (!state.selectedWorldId) { toast(tr('Select a world'), 'error'); return; }
  showLoading(tr('Replanning timeline...'));
  try {
    const res = await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/timeline/replan', null);
    hideLoading();
    const resultEl = ce('div', { className: 'modal-field' }, [
      el('pre', { style: {fontSize: '11px', whiteSpace: 'pre-wrap', maxHeight: '300px', overflow: 'auto', background: 'var(--bg-input)', padding: '8px', borderRadius: 'var(--radius)'}, textContent: JSON.stringify(res, null, 2) }),
    ]);
    toast(tr('Replan done'), 'success');
    openModal(tr('Replan Result'), resultEl, ce('div', {}, [el('button', { id: 'modalCloseResultBtn', textContent: tr('Close') })]));
    document.getElementById('modalCloseResultBtn').addEventListener('click', closeModal);
  } catch(e) { hideLoading(); showApiError('Failed', e); }
}

/* ============= Autonomous Config Modal ============= */
function openAutonomousConfigModal() {
  const ac = state.autonomous ? state.autonomous.config : null;
  if (!ac) { toast(tr('Load autonomous config first'), 'error'); return; }
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'autoEnabled' }, [ttxt('Enabled')]),
    el('input', { type: 'checkbox', id: 'autoEnabled', checked: !!ac.enabled }),
    ce('label', { for: 'autoTrigger' }, [ttxt('Trigger')]),
    el('select', { id: 'autoTrigger', innerHTML: '<option value="manual">manual</option><option value="world_tick_sync">world_tick_sync</option><option value="scheduled">scheduled</option>' }),
    ce('label', { for: 'autoInterval' }, [ttxt('Interval Seconds (scheduled)')]),
    el('input', { id: 'autoInterval', type: 'number', value: String(ac.interval_seconds || 300), min: '1', style: {width: '100px'} }),
    ce('label', { for: 'autoCapabilities' }, [ttxt('Capabilities (JSON array)')]),
    el('textarea', { id: 'autoCapabilities', rows: 4, style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'}, textContent: JSON.stringify(ac.capabilities || [], null, 2) }),
  ]);
  openModal(tr('Autonomous Config'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalSaveAutoBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var at = document.getElementById('autoTrigger'); if (at) at.value = (ac && ac.trigger) || 'manual';
  document.getElementById('modalSaveAutoBtn').addEventListener('click', saveAutonomousConfig);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function saveAutonomousConfig() { if (!state.selectedNodeId) { toast(tr('Please select a node first'), 'error'); return; }
  const enabled = document.getElementById('autoEnabled').checked;
  const trigger = document.getElementById('autoTrigger').value;
  const interval = parseInt(document.getElementById('autoInterval').value) || 300;
  const capsText = document.getElementById('autoCapabilities').value.trim();
  let capabilities = [];
  if (capsText) { try { capabilities = JSON.parse(capsText); } catch(e) { toast(tr('Invalid JSON'), 'error'); return; } }
  try {
    await api('PUT', '/api/v1/nodes/' + encodeURIComponent(state.selectedNodeId) + '/autonomous', { enabled: enabled, trigger: trigger, interval_seconds: interval, capabilities: capabilities });
    closeModal(); toast(tr('Autonomous config saved'), 'success'); loadAutonomous();
  } catch(e) { showApiError('Failed', e); }
}


function loadTasks(silent) {
  if (!silent) showLoading(true);
  api("GET", "/api/v1/runtime/tasks?limit=100", null).then(function(data) {
    try {
      state.tasks = data && data.tasks ? data.tasks : [];
    } catch(e) {
      state.tasks = [];
    }
    if (state.page === "tasks") renderCurrent();
    if (!silent) { showLoading(false); toast(tr("Tasks refreshed"), "success"); }
  }).catch(function(err) {
    if (!silent) { showLoading(false); toast(err.message || "Error", "error"); }
  });
}
