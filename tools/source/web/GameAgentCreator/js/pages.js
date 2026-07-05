/* ============= Center Page Content ============= */
function renderCenter() {
  const center = document.getElementById('centerContent');
  center.innerHTML = '';
  const selBar = ce('div', { id: 'selBar', className: 'sel-bar' }, []);
  if (state.selectedNodeId) {
    const n = state.nodes.find(function(x) { return x.id === state.selectedNodeId; });
    if (n) {
      selBar.appendChild(ce('span', { className: 'lbl' }, [ttxt('Selected:')]));
      selBar.appendChild(ce('span', { className: 'name' }, [txt(n.name)]));
      selBar.appendChild(ce('span', { className: 'id' }, [txt(n.id.slice(0,8))]));
    }
  } else if (state.selectedWorldId) {
    const w = state.worlds.find(function(x) { return x.id === state.selectedWorldId; });
    if (w) {
      selBar.appendChild(ce('span', { className: 'lbl' }, [ttxt('World:')]));
      selBar.appendChild(ce('span', { className: 'name' }, [txt(w.name)]));
    }
  }
  if (state.selectedNodeId || state.selectedWorldId) center.appendChild(selBar);
  switch (state.page) {
    case 'worlds': renderWorldsPage(center); break;
    case 'policy': renderPolicyPage(center); break;
    case 'settings': renderSettingsPage(center); break;
    case 'logs': renderLogsPage(center); break;
    case 'traces': renderTracesPage(center); break;
    default: renderWorldsPage(center);
  }
}

/* ============= Worlds Page ============= */
function renderWorldsPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
        ce('button', { id: 'btnCloneWorld' }, [ttxt('Clone World')]),
    ce('button', { id: 'btnTickAdvance' }, [ttxt('Advance Tick')]),
    ce('button', { id: 'btnAutonomous' }, [ttxt('Run Autonomous')]),
    ce('button', { id: 'btnEventImpact' }, [ttxt('Event Impact')]),
    ce('button', { id: 'btnScopeAdvance' }, [ttxt('Scope Advance')]),
    ce('button', { id: 'btnReplan' }, [ttxt('Replan')]),
  ]);
  container.appendChild(toolbar);
  if (state.selectedNodeId && state.nodeDetail) {
    renderNodeDetail(container);
  } else if (state.selectedWorldId) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a node to inspect, or right-click in the outline.')]));
  } else {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a world to begin editing.')]));
  }
    document.getElementById('btnCloneWorld').addEventListener('click', cloneWorld);
  document.getElementById('btnTickAdvance').addEventListener('click', tickAdvance);
  document.getElementById('btnAutonomous').addEventListener('click', runAutonomous);
  document.getElementById('btnEventImpact').addEventListener('click', openEventImpactModal);
  document.getElementById('btnScopeAdvance').addEventListener('click', scopeAdvance);
  document.getElementById('btnReplan').addEventListener('click', timelineReplan);
  updateActionButtons();
}

/* ============= Node Detail ============= */
function renderNodeDetail(container) {
  const nd = state.nodeDetail;
  if (!nd || !nd.node) return;
  const n = nd.node;
  const detail = ce('div', { className: 'node-detail' }, []);
  
  // Overview card
  const overview = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd toggle-hd', 'aria-expanded': 'true' }, [ttxt('Overview')]),
    ce('div', { className: 'card-bd' }, [
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('ID')]), ce('span', { className: 'val mono' }, [txt(n.id)])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Name')]), ce('span', { className: 'val' }, [txt(n.name)])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Type')]), ce('span', { className: 'val' }, [txt(n.node_type)])]),
      n.parent_id ? ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Parent')]), ce('span', { className: 'val mono' }, [txt(n.parent_id.slice(0,8))])]) : null,
    ]),
  ]);
  detail.appendChild(overview);
  
  // Edit node button
  
  // Components — full detail
  const compSection = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd toggle-hd', 'aria-expanded': 'true' }, [function(){var c=nd.components?nd.components.length:0;return ce('span',{},[ttxt('Components'),txt(' ('+c+')')])}(), ce('button', { id: 'btnAddComponentCenter', className: 'sm' }, [ttxt('+')])]),
    ce('div', { className: 'card-bd', id: 'compDetailList' }, []),
  ]);
  detail.appendChild(compSection);
  
  // Memories — full detail
  const memSection = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd toggle-hd', 'aria-expanded': 'true' }, [function(){var c=nd.memories?nd.memories.length:0;return ce('span',{},[ttxt('Memories'),txt(' ('+c+')')])}(), ce('button', { id: 'btnAddMemoryCenter', className: 'sm' }, [ttxt('+')])]),
    ce('div', { className: 'card-bd', id: 'memDetailList' }, []),
  ]);
  detail.appendChild(memSection);
  
  // Relations — full detail
  const relSection = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd toggle-hd', 'aria-expanded': 'true' }, [function(){var c=nd.relations?nd.relations.length:0;return ce('span',{},[ttxt('Relations'),txt(' ('+c+')')])}(), ce('button', { id: 'btnAddRelationCenter', className: 'sm' }, [ttxt('+')])]),
    ce('div', { className: 'card-bd', id: 'relDetailList' }, []),
  ]);
  detail.appendChild(relSection);
  
  container.appendChild(detail);
  
  // Populate component list
  const compList = document.getElementById('compDetailList');
  if (compList && nd.components) {
    for (const c of nd.components) {
      const row = ce('div', { className: 'detail-item', dataset: { id: c.id } }, [
        ce('div', { className: 'item-hd' }, [
          ce('span', { className: 'item-tag', style: {background: 'var(--accent-soft)', color: 'var(--accent)'} }, [txt(c.component_type)]),
          ce('span', { style: {fontSize: '10px', color: 'var(--text-dim)'} }, [txt(' ' + (c.id ? c.id.substring(0,8) : ''))]),
          ce('button', { className: 'item-edit', dataset: { id: c.id } }, [ttxt('\u270e')]),
          ce('button', { className: 'item-del', dataset: { id: c.id } }, [ttxt('\u2715')]),
        ]),
        ce('div', { className: 'item-body' }, [function(){var d=c.data||'';if(isJSON(d)){var p=tryParseJSON(d);return p?renderObjectKV(p,false):txt(d)}return txt(d)}()]),
      ]);
      compList.appendChild(row);
    }
  }
  
  // Populate memory list
  const memList = document.getElementById('memDetailList');
  if (memList && nd.memories) {
    for (const m of nd.memories) {
      const levelColors = {short_term: '#f59e0b', long_term: '#3b82f6', shared: '#10b981', world: '#8b5cf6'};
      const row = ce('div', { className: 'detail-item', dataset: { id: m.id } }, [
        ce('div', { className: 'item-hd' }, [
          ce('span', { className: 'item-tag', style: {background: 'rgba(59,130,246,.12)', color: levelColors[m.level] || '#3b82f6'} }, [txt(m.level || 'long_term')]),
          ce('span', { style: {fontSize: '10px', color: 'var(--text-dim)'} }, [txt(m.tags ? ' ' + m.tags : '')]),
          ce('button', { className: 'item-edit', dataset: { id: m.id } }, [ttxt('\u270e')]),
          ce('button', { className: 'item-del', dataset: { id: m.id } }, [ttxt('\u2715')]),
        ]),
        ce('div', { className: 'item-body' }, [function(){var d=m.content||'';if(isJSON(d)){var p=tryParseJSON(d);return p?renderObjectKV(p,false):txt(d)}return txt(d)}()]),
      ]);
      memList.appendChild(row);
    }
  }
  
  // Populate relation list
  const relList = document.getElementById('relDetailList');
  if (relList && nd.relations) {
    for (const r of nd.relations) {
      var srcName = r.source_id; var tgtName = r.target_id;
      if (state.nodes) {
        var src = state.nodes.find(function(x) { return x.id === r.source_id; }); if (src) srcName = src.name;
        var tgt = state.nodes.find(function(x) { return x.id === r.target_id; }); if (tgt) tgtName = tgt.name;
      }
      const row = ce('div', { className: 'detail-item', dataset: { id: r.id } }, [
        ce('div', { className: 'item-hd' }, [
          ce('span', { className: 'item-tag', style: {background: 'rgba(16,185,129,.12)', color: 'var(--green)'} }, [txt(r.relation_type)]),
          ce('span', { style: {fontSize: '10px', color: 'var(--text-dim)'} }, [txt(' ' + r.source_id.substring(0,8) + ' -> ' + r.target_id.substring(0,8))]),
          ce('button', { className: 'item-edit', dataset: { id: r.id } }, [ttxt('\u270e')]),
          ce('button', { className: 'item-del', dataset: { id: r.id } }, [ttxt('\u2715')]),
        ]),
        ce('div', { className: 'item-body' }, [function(){if(r.properties&&isJSON(r.properties)){var p=tryParseJSON(r.properties);return p?renderObjectKV(p,false):txt(r.properties)}return txt('Weight: '+r.weight)}()]),
      ]);
      relList.appendChild(row);
    }
  }
  
  // Bind events
  document.getElementById('btnAddComponentCenter').addEventListener('click', openAddComponentModal);
  document.getElementById('btnAddMemoryCenter').addEventListener('click', openAddMemoryModal);
  document.getElementById('btnAddRelationCenter').addEventListener('click', openAddRelationModal);
  
  // Edit/delete buttons
  compList.querySelectorAll('.item-edit').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); openEditComponentModal(btn.dataset.id); }); });
  compList.querySelectorAll('.item-del').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); deleteComponent(btn.dataset.id); }); });
  memList.querySelectorAll('.item-edit').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); openEditMemoryModal(btn.dataset.id); }); });
  memList.querySelectorAll('.item-del').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); deleteMemory(btn.dataset.id); }); });
  relList.querySelectorAll('.item-edit').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); openEditRelationModal(btn.dataset.id); }); });
  relList.querySelectorAll('.item-del').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); deleteRelation(btn.dataset.id); }); });
  
  // Toggle collapse/expand for detail cards
  document.querySelectorAll('.toggle-hd').forEach(function(hd) {
    hd.addEventListener('click', function(e) {
      if (e.target.tagName === 'BUTTON') return;
      var expanded = hd.getAttribute('aria-expanded') === 'true';
      hd.setAttribute('aria-expanded', expanded ? 'false' : 'true');
    });
  });
}
/* ============= Policy Page ============= */
function renderPolicyPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('span', { style: {color: 'var(--text-dim)'} }, [ttxt('World Policy')]),
  ]);
  container.appendChild(toolbar);
  if (!state.selectedWorldId) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a world first.')]));
    return;
  }
  const policy = state.policy || { blocked_actions: [], safe_actions: [] };
  const form = ce('div', { className: 'policy-form' }, []);
  const blk = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd' }, [ttxt('Blocked Actions')]),
    ce('div', { className: 'card-bd' }, [
      el('textarea', { id: 'policyBlocked', placeholder: 'One action per line', rows: 5, style: {width: '100%', fontFamily: 'var(--font-mono)'}, textContent: (policy.blocked_actions || []).join('\n') }),
    ]),
  ]);
  form.appendChild(blk);
  const sf = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd' }, [ttxt('Safe Actions')]),
    ce('div', { className: 'card-bd' }, [
      el('textarea', { id: 'policySafe', placeholder: 'One action per line', rows: 5, style: {width: '100%', fontFamily: 'var(--font-mono)'}, textContent: (policy.safe_actions || []).join('\n') }),
    ]),
  ]);
  form.appendChild(sf);
  const btnRow = ce('div', { className: 'policy-actions' }, [
    ce('button', { className: 'primary', id: 'btnSavePolicy' }, [ttxt('Save Policy')]),
  ]);
  form.appendChild(btnRow);
  container.appendChild(form);
  document.getElementById('btnSavePolicy').addEventListener('click', savePolicy);
}

/* ============= Settings Page ============= */
function renderSettingsPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('span', { style: {color: 'var(--text-dim)'} }, [ttxt('World Settings')]),
  ]);
  container.appendChild(toolbar);
  if (!state.selectedWorldId) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a world first.')]));
    return;
  }
  const s = state.settings || { memory_limit: 50, max_analysis_rounds: 5, max_context_depth: 3, auto_apply: true, require_review_above: 'critical' };
  const form = ce('div', { className: 'settings-form' }, []);
  function settingRow(label, id, type, val) {
    const row = ce('div', { className: 'setting-row' }, [ce('label', { for: id }, [ttxt(label)])]);
    if (type === 'bool') {
      row.appendChild(el('input', { type: 'checkbox', id: id, checked: !!val }));
    } else if (type === 'select_pipeline') {
      const sel = el('select', { id: id });
      var opts = ['vertical','polling','full'];
      for (const o of opts) sel.appendChild(el('option', { value: o, textContent: o + ' (' + (o==='vertical'?tr('Single round'):o==='polling'?tr('Multi-round'):tr('Full features')) + ')', selected: val === o }));
      row.appendChild(sel);
    } else if (type === 'select') {
      const sel = el('select', { id: id });
      var opts = ['none','low','medium','high','critical'];
      for (const o of opts) sel.appendChild(el('option', { value: o, textContent: o, selected: val === o }));
      row.appendChild(sel);
    } else {
      row.appendChild(el('input', { type: 'number', id: id, value: String(val), min: '1', max: '999' }));
    }
    form.appendChild(row);
  }
  form.appendChild(ce('div', { className: 'detail-card' }, [ce('div', { className: 'card-hd' }, [ttxt('Inference Params')])]));
  settingRow('Memory Limit', 'setMemoryLimit', 'number', s.memory_limit);
  settingRow('Analysis Rounds', 'setMaxRounds', 'number', s.max_analysis_rounds);
  settingRow('Context Depth', 'setMaxDepth', 'number', s.max_context_depth);
  form.appendChild(ce('div', { className: 'detail-card' }, [ce('div', { className: 'card-hd' }, [ttxt('Execution Control')])]));
  settingRow('Auto Apply', 'setAutoApply', 'bool', s.auto_apply);
  settingRow('Review Threshold', 'setReviewAbove', 'select', s.require_review_above);
  form.appendChild(ce('div', { className: 'detail-card' }, [ce('div', { className: 'card-hd' }, [ttxt('Pipeline & Propagation')])]));
  settingRow('Pipeline Mode', 'setPipelineMode', 'select_pipeline', s.pipeline_mode || 'full');
  settingRow('Propagation Max Depth', 'setPropMaxDepth', 'number', s.propagation_max_depth || 2);
  settingRow('Enable Propagation Machine', 'setEnablePropMachine', 'bool', s.enable_propagation_machine);
  form.appendChild(ce('div', { className: 'detail-card' }, [ce('div', { className: 'card-hd' }, [ttxt('Sub-Task DAG')])]));
  settingRow('Sub-Task Max Retries', 'setSubTaskRetries', 'number', s.sub_task_max_retries || 2);
  settingRow('Sub-Task Timeout (sec)', 'setSubTaskTimeout', 'number', s.sub_task_timeout_secs || 60);
  const btnRow = ce('div', { className: 'policy-actions' }, [
    ce('button', { className: 'primary', id: 'btnSaveSettings' }, [ttxt('Save Settings')]),
  ]);
  form.appendChild(btnRow);
  container.appendChild(form);
  document.getElementById('btnSaveSettings').addEventListener('click', saveSettings);
}

/* ============= Logs Page ============= */
function renderLogsPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnRefreshLogs' }, [ttxt('Refresh Logs')]),
  ]);
  container.appendChild(toolbar);
  const tbl = ce('div', { className: 'log-table' }, []);
  const hdr = ce('div', { className: 'log-row log-hdr' }, [
    ce('span', { className: 'log-c time' }, [ttxt('Time')]),
    ce('span', { className: 'log-c type' }, [ttxt('Type')]),
    ce('span', { className: 'log-c model' }, [ttxt('Model')]),
    ce('span', { className: 'log-c tokens' }, [ttxt('Tokens')]),
    ce('span', { className: 'log-c dur' }, [ttxt('Duration')]),
  ]);
  tbl.appendChild(hdr);
  for (const log of state.logs) {
    var t = log.created_at;
    try { t = new Date(t).toLocaleTimeString(); } catch(e) {}
    const r = ce('div', { className: 'log-row' }, [
      ce('span', { className: 'log-c time' }, [txt(t)]),
      ce('span', { className: 'log-c type' }, [txt(log.task_type || '-')]),
      ce('span', { className: 'log-c model' }, [txt(log.llm_model || '-')]),
      ce('span', { className: 'log-c tokens' }, [txt(String(log.tokens_used || 0))]),
      ce('span', { className: 'log-c dur' }, [txt(String(log.duration_ms || 0) + 'ms')]),
    ]);
    tbl.appendChild(r);
  }
  if (state.logs.length === 0) tbl.appendChild(ce('div', { className: 'hint' }, [ttxt('No logs yet.')]));
  container.appendChild(tbl);
  document.getElementById('btnRefreshLogs').addEventListener('click', loadLogs);
}


function updateActionButtons() {
  var btnTick = document.getElementById('btnTickAdvance');
  var btnAuto = document.getElementById('btnAutonomous');
  var btnEvent = document.getElementById('btnEventImpact');
  var btnScope = document.getElementById('btnScopeAdvance');
  var btnReplan = document.getElementById('btnReplan');
  if (btnTick) btnTick.classList.toggle('dim', !state.selectedWorldId);
  if (btnEvent) btnEvent.classList.toggle('dim', !state.selectedWorldId);
  if (btnScope) btnScope.classList.toggle('dim', !state.selectedWorldId);
  if (btnReplan) btnReplan.classList.toggle('dim', !state.selectedWorldId);
  if (btnAuto) btnAuto.classList.toggle('dim', !state.selectedWorldId || !state.selectedNodeId);
}
/* ============= Traces Page (Debug) ============= */
async function renderTracesPage(container) {
  const toolbar = ce("div", { className: "world-toolbar" }, [
    ce("button", { id: "btnRefreshTraces" }, [ttxt("Refresh Traces")]),
  ]);
  container.appendChild(toolbar);

  const traceContainer = ce("div", { className: "trace-container" }, []);
  container.appendChild(traceContainer);

  await loadTraces(traceContainer);

  document.getElementById("btnRefreshTraces").addEventListener("click", async function() {
    await loadTraces(traceContainer);
  });
}

async function loadTraces(container) {
  container.innerHTML = ce("div", { className: "hint" }, [txt("Loading...")]);

  try {
    var q = "/debug/traces?limit=30";
    if (state.selectedWorldId) q += "&world_id=" + encodeURIComponent(state.selectedWorldId);
    var data = await api("GET", q);
    var traces = data.traces || [];
    container.innerHTML = "";

    if (traces.length === 0) {
      container.appendChild(ce("div", { className: "hint" }, [ttxt("No traces yet. Run a task in Debug mode to see traces.")]));
      return;
    }

    for (var ti = 0; ti < traces.length; ti++) {
      (function() {
        var t = traces[ti];
        var card = ce("div", { className: "trace-card detail-card" }, []);
        var hd = ce("div", { className: "card-hd toggle-hd", "aria-expanded": "false" }, [
          ce("span", { style: { fontWeight: 600 } }, [txt(t.task_type || "?")]),
          txt(" " + (t.duration_ms || 0) + "ms"),
          txt(" [" + (t.round != null ? "round " + t.round : "") + "]"),
          t.error ? ce("span", { style: { color: "var(--red)", marginLeft: "8px" } }, [txt(" ERROR")]) : null,
        ]);
        card.appendChild(hd);

        var body = ce("div", { className: "card-bd", style: { display: "none" } }, []);

        // Basic info
        body.appendChild(ce("div", { className: "prop-row" }, [ce("span", { className: "key" }, [txt("World")]), ce("span", { className: "val mono" }, [txt(t.world_id ? t.world_id.slice(0, 8) : "-")])]));
        body.appendChild(ce("div", { className: "prop-row" }, [ce("span", { className: "key" }, [txt("Node")]), ce("span", { className: "val mono" }, [txt(t.node_id ? t.node_id.slice(0, 8) : "-")])]));
        body.appendChild(ce("div", { className: "prop-row" }, [ce("span", { className: "key" }, [txt("Duration")]), ce("span", { className: "val" }, [txt(t.duration_ms + "ms")])]));

        // System Prompt (collapsible)
        if (t.system_prompt) {
          body.appendChild(ce("div", { className: "prop-row trace-block" }, [
            ce("span", { className: "key" }, [ttxt("System Prompt")]),
            ce("pre", { className: "trace-pre" }, [txt(t.system_prompt.slice(0, 1000) + (t.system_prompt.length > 1000 ? "..." : ""))]),
          ]));
        }

        // LLM Response (collapsible)
        if (t.raw_llm_response) {
          body.appendChild(ce("div", { className: "prop-row trace-block" }, [
            ce("span", { className: "key" }, [ttxt("LLM Response")]),
            ce("pre", { className: "trace-pre" }, [txt(t.raw_llm_response.slice(0, 2000) + (t.raw_llm_response.length > 2000 ? "..." : ""))]),
          ]));
        }

        // Actions
        if (t.parsed_actions && t.parsed_actions.length > 0) {
          var actSection = ce("div", { className: "prop-row trace-block" }, [
            ce("span", { className: "key" }, [txt("Actions (" + t.parsed_actions.length + ")")]),
          ]);
          for (var ai = 0; ai < t.parsed_actions.length; ai++) {
            var a = t.parsed_actions[ai];
            actSection.appendChild(ce("div", { style: { paddingLeft: "12px", fontSize: "12px", color: "var(--fg2)" } }, [txt(a.action_id + " " + (a.mode ? "[" + a.mode + "]" : ""))]));
          }
          body.appendChild(actSection);
        }

        // Memories
        if (t.parsed_memories && t.parsed_memories.length > 0) {
          var memSection = ce("div", { className: "prop-row trace-block" }, [
            ce("span", { className: "key" }, [txt("Memories (" + t.parsed_memories.length + ")")]),
          ]);
          for (var mi = 0; mi < t.parsed_memories.length; mi++) {
            var m = t.parsed_memories[mi];
            var mContent = (m.content || "").slice(0, 100);
            memSection.appendChild(ce("div", { style: { paddingLeft: "12px", fontSize: "12px", color: "var(--fg2)" } }, [txt(m.node_id.slice(0, 8) + " " + m.level + ": " + mContent + (m.content && m.content.length > 100 ? "..." : ""))]));
          }
          body.appendChild(memSection);
        }

        card.appendChild(body);

        hd.addEventListener("click", function() {
          var expanded = this.getAttribute("aria-expanded") === "true";
          this.setAttribute("aria-expanded", expanded ? "false" : "true");
          body.style.display = expanded ? "none" : "block";
        });

        container.appendChild(card);
      })();
    }
  } catch (e) {
    container.innerHTML = "";
    container.appendChild(ce("div", { className: "hint" }, [txt("Failed to load traces: " + e.message)]));
  }
}

