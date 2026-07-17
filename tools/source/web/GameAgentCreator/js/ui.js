/* ============= Topbar ============= */
function renderTopbar() {
  const topbar = document.getElementById('topbar');
  topbar.innerHTML = '';
  const left = ce('div', { className: 'topbar-left' }, [
    ce('span', { className: 'brand' }, [ttxt('GameAgentCreator')]),
    ce('button', { id: 'topbarCreateWorld', className: 'topbar-action primary', style: {fontSize:'11px', padding:'0 8px', minHeight:'24px'} }, [ttxt('Create World')]),
    ce('button', { id: 'topbarImport', className: 'topbar-action', style: {fontSize:'11px', padding:'0 8px', minHeight:'24px'} }, [ttxt('Import Config')]),
  ]);
  const center = ce('div', { className: 'topbar-center' }, []);
  const sel = el('select', { id: 'worldSelector' });
  sel.appendChild(el('option', { value: '', textContent: tr('-- Select World --') }));
  for (const w of state.worlds) {
    sel.appendChild(el('option', { value: w.id, textContent: w.name }));
  }
  if (state.selectedWorldId) sel.value = state.selectedWorldId;
  sel.addEventListener('change', function() { selectWorld(this.value); });
  center.appendChild(sel);
  function pageBtn(name, label) {
    const b = ce('button', { className: state.page === name ? 'active' : '', dataset: { page: name } }, [txt(label)]);
    b.addEventListener('click', function() { switchPage(name); });
    return b;
  }
  center.appendChild(pageBtn('worlds', tr('Worlds')));
  center.appendChild(pageBtn('snapshots', tr('Snapshots')));
  center.appendChild(pageBtn('tasks', tr('Tasks')))
  center.appendChild(pageBtn('plans', tr('Plans')));
  center.appendChild(pageBtn('policy', tr('Policy')));
  center.appendChild(pageBtn('settings', tr('Settings')));
  center.appendChild(pageBtn('continuity', tr('Continuity')));
  center.appendChild(pageBtn('state', tr('State')));
  center.appendChild(pageBtn('timelines', tr('Timelines')));
  center.appendChild(pageBtn('logs', tr('Logs')));
  center.appendChild(pageBtn('traces', tr('Traces')));
  const right = ce('div', { className: 'topbar-right' }, [
    el('span', { id: 'connStatus', className: 'status', innerHTML: '<span class="status-dot off"></span> ' + tr('Disconnected') }),
    ce('button', { id: 'btnConfig', title: tr('Server Config') }, [txt('\u2699')]),
  ]);
  topbar.appendChild(left);
  const rightSec = ce('div', { className: 'topbar-right' }, [
    ce('button', { className: 'lang-btn', onclick: toggleLang, title: tr('Switch Language') }, [ce('span', { className: 'lang-icon' }, [txt(lang === 'zh' ? 'EN' : '中')])]),
    ce('button', { className: 'theme-btn', onclick: toggleTheme, title: tr('Toggle Theme') }, [txt(theme === 'dark' ? '\u2600' : '\u2601')]),
  ]);
  topbar.appendChild(rightSec);
  topbar.appendChild(center);
  topbar.appendChild(right);
  document.getElementById('btnConfig').addEventListener('click', openConfigModal);
  document.getElementById('topbarCreateWorld').addEventListener('click', openCreateWorldModal);
  document.getElementById('topbarImport').addEventListener('click', openImportModal);
}

/* ============= Left Panel ============= */
function renderLeftPanel() {
  const lp = document.getElementById('leftPanel');
  lp.innerHTML = '';
  const hd = ce('div', { className: 'panel-hd' }, [
    ce('span', { className: 'title' }, [ttxt('World Outline')]),
    ce('button', { className: 'close', id: 'btnTreeRefresh' }, [txt('\u21bb')]),
  ]);
  lp.appendChild(hd);
  const tb = ce('div', { className: 'tree-toolbar' }, [
    el('input', { id: 'treeFilter', placeholder: tr('Filter nodes...'), value: state.treeFilter }),
    ce('button', { id: 'btnAddNode' }, [txt('+')]),
  ]);
  lp.appendChild(tb);
  lp.appendChild(ce('div', { className: 'hint', style: { padding: '6px 10px 0 10px', textAlign: 'left' } }, [
    txt(tr('Outline drag, drop, and Add New Parent only update the node primary parent. Use Relations to edit non-tree links.')),
  ]));
  const body = el('div', { id: 'treeBody', className: 'tree-body' });
  lp.appendChild(body);
  document.getElementById('btnAddNode').addEventListener('click', openCreateNodeModal);
  document.getElementById('btnTreeRefresh').addEventListener('click', loadCurrentWorld);
  document.getElementById('treeFilter').addEventListener('input', function() {
    state.treeFilter = this.value;
    invalidateTreeCache();
    if (this._debounceTimer) clearTimeout(this._debounceTimer);
    var self = this;
    self._debounceTimer = setTimeout(function() { renderTree(); }, 80);
  });
}

/* ============= Tree ============= */
function renderTree() {
  var body = document.getElementById('treeBody');
  if (!body) return;
  body.innerHTML = '';
  state.visibleNodeIds = [];

  var rows = buildFlatRows();
  if (rows.length === 0) {
    body.appendChild(ce('div', { className: 'hint' }, [ttxt('No nodes. Click + to create.')]));
    return;
  }

  // Track visible node IDs for drag/drop
  state.visibleNodeIds = rows.map(function(r) { return r.nodeId; });

  // Build root drop zone
  var rootDrop = ce('div', { className: 'tree-root-drop' }, [txt(tr('Drop here to move to root'))]);
  body.appendChild(rootDrop);

  // Node container
  var treeContent = ce('div', { className: 'tree-content' }, []);
  body.appendChild(treeContent);

  var selectedSet = {};
  (state.selectedNodeIds || []).forEach(function(id) { selectedSet[id] = true; });
  var activePathSet = {};
  if (state.selectedTreePathKey) {
    state.selectedTreePathKey.split('|').forEach(function(key) { if (key) activePathSet[key] = true; });
  }

  for (var ri = 0; ri < rows.length; ri++) {
    var r = rows[ri];
    var node = _treeCache.nodeMap[r.nodeId];
    if (!node) continue;

    var isSelected = !!selectedSet[node.id] && state.selectedTreePathKey === r.pathKey;
    var isPrimarySelected = state.selectedNodeId === node.id;
    var isAliasSelected = !!selectedSet[node.id] && state.selectedTreePathKey && state.selectedTreePathKey !== r.pathKey;
    var isAncestor = !!activePathSet[r.pathKey] && !isSelected;

    var cls = 'tree-node';
    if (isSelected) cls += ' selected';
    if (isSelected && !isPrimarySelected) cls += ' multi-selected';
    if (isAliasSelected) cls += ' alias-selected';
    if (isAncestor) cls += ' path-ancestor';

    var arrowSpan = el('span', {
      className: 'tree-arrow' + (r.hasChildren ? (r.isExpanded ? ' expanded' : '') : ' invisible'),
      textContent: '\u25b8'
    });
    var iconSpan = el('span', { className: 'tree-icon ' + node.node_type });
    var nameSpan = el('span', { className: 'tree-name', textContent: node.name });
    var typeSpan = el('span', { className: 'tree-type node-type-' + node.node_type, textContent: node.node_type });

    var row = ce('div', {
      className: cls,
      dataset: { id: node.id, pid: node.parent_id || '', pathKey: r.pathKey, depth: String(r.depth), hasChildren: r.hasChildren ? '1' : '0' }
    }, [arrowSpan, iconSpan, nameSpan, typeSpan]);
    row.style.paddingLeft = r.paddingLeft + 'px';
    treeContent.appendChild(row);

    // Drag source: mark for drag
    if (node.node_type !== 'world') {
      row.draggable = true;
    }
  }

  // Event delegation: click on tree arrows
  treeContent.addEventListener('click', function(e) {
    if (state.suppressTreeClickUntil && Date.now() < state.suppressTreeClickUntil) return;
    var target = e.target;
    var row = target.closest('.tree-node');
    if (!row) return;

    var nodeId = row.dataset.id;
    var node = getNodeById(nodeId);
    if (!node) return;

    // Arrow click: expand/collapse
    if (target.classList.contains('tree-arrow') && !target.classList.contains('invisible')) {
      e.stopPropagation();
      if (!state.treeCollapsed) state.treeCollapsed = {};
      state.treeCollapsed[nodeId] = !state.treeCollapsed[nodeId];
      invalidateTreeCache();
      renderTree();
      return;
    }

    // Node selection
    if (e.shiftKey) {
      selectNode(nodeId, node.node_type, { mode: 'range', preserveAnchor: true, treePathKey: row.dataset.pathKey });
    } else if (e.ctrlKey || e.metaKey) {
      selectNode(nodeId, node.node_type, { mode: 'toggle', treePathKey: row.dataset.pathKey });
    } else {
      selectNode(nodeId, node.node_type, { mode: 'single', treePathKey: row.dataset.pathKey });
    }
  });

  // Event delegation: drag + right-click
  var dragState = null;
  function clearDrag() {
    if (dragState) {
      dragState.sourceRow.classList.remove('drag-source');
      document.body.style.userSelect = '';
      document.body.style.cursor = '';
      dragState = null;
    }
    body.classList.remove('drag-active');
    var drops = body.querySelectorAll('.drop-target');
    for (var di = 0; di < drops.length; di++) drops[di].classList.remove('drop-target');
    rootDrop.classList.remove('active');
  }

  treeContent.addEventListener('mousedown', function(e) {
    if (e.button !== 0) return;
    var row = e.target.closest('.tree-node');
    if (!row) return;
    var node = getNodeById(row.dataset.id);
    if (!node || node.node_type === 'world') return;
    if (row.querySelector('.tree-arrow') === e.target) return;

    var startX = e.clientX, startY = e.clientY;
    var started = false;

    function onMove(ev) {
      var dx = Math.abs(ev.clientX - startX), dy = Math.abs(ev.clientY - startY);
      if (!started) {
        if (Math.max(dx, dy) < 5) return;
        started = true;
        row.classList.add('drag-source');
        body.classList.add('drag-active');
        document.body.style.userSelect = 'none';
        document.body.style.cursor = 'grabbing';
        dragState = { sourceRow: row };
      }
      ev.preventDefault();

      // Clear old indicators
      rootDrop.classList.remove('active');
      var oldDrops = body.querySelectorAll('.drop-target');
      for (var di = 0; di < oldDrops.length; di++) oldDrops[di].classList.remove('drop-target');

      var hit = document.elementFromPoint(ev.clientX, ev.clientY);
      if (!hit) return;

      // Check root drop
      var rootHit = hit === rootDrop || (hit.closest && hit.closest('.tree-root-drop') === rootDrop);
      if (rootHit) { rootDrop.classList.add('active'); return; }

      var targetRow = hit.closest ? hit.closest('.tree-node') : null;
      if (targetRow && targetRow !== row) targetRow.classList.add('drop-target');
    }

    function onUp(ev) {
      window.removeEventListener('mousemove', onMove, true);
      window.removeEventListener('mouseup', onUp, true);
      if (!started) { clearDrag(); return; }
      ev.preventDefault();
      state.suppressTreeClickUntil = Date.now() + 250;

      var targetId = null;
      if (rootDrop.classList.contains('active')) { targetId = ''; }
      else {
        var tRow = body.querySelector('.drop-target');
        if (tRow) targetId = tRow.dataset.id || null;
      }

      clearDrag();
      if (targetId !== null) moveNodeParent(row.dataset.id, targetId);
    }

    window.addEventListener('mousemove', onMove, true);
    window.addEventListener('mouseup', onUp, true);
  });

  // Context menu delegation
  treeContent.addEventListener('contextmenu', async function(e) {
    e.preventDefault();
    e.stopPropagation();
    var row = e.target.closest('.tree-node');
    if (!row) return;
    var nodeId = row.dataset.id;
    var node = getNodeById(nodeId);
    if (!node) return;

    var selectedSet2 = {};
    (state.selectedNodeIds || []).forEach(function(id) { selectedSet2[id] = true; });
    if (!selectedSet2[nodeId] || state.selectedTreePathKey !== row.dataset.pathKey) {
      await selectNode(nodeId, node.node_type, { mode: 'single', treePathKey: row.dataset.pathKey });
    }
    showContextMenu([
      { label: tr('Edit'), onClick: function() { openEditNodeModal(nodeId); } },
      { label: tr('Copy Node'), onClick: function() { openCopyNodeModal(nodeId); } },
      { label: tr('Add New Parent'), tip: tr('Creates a new node and rewires only parent_id'), onClick: function() { openCreateParentNodeModal(nodeId); } },
      { label: tr('Add Outgoing Relation'), tip: tr('Relations are stored separately from the outline tree'), onClick: function() { openAddOutgoingRelationModal(nodeId); } },
      { label: tr('Create Child'), onClick: function() { openCreateNodeModal(nodeId); } },
      { label: tr('Delete'), danger: true, onClick: function() { deleteNodeHandler(nodeId); } },
    ], e.clientX, e.clientY);
  });
}

function switchPage(name) {
  state.page = name;
  var tabs = document.querySelectorAll('.topbar-center button');
  tabs.forEach(function(b) { b.classList.remove('active'); });
  var found = Array.from(tabs).find(function(b) { return b.dataset.page === name; });
  if (found) found.classList.add('active');
  renderCurrent();

  switch (name) {
    case 'tasks': loadTasks(true); break;
    case 'snapshots': loadSnapshots(); break;
    case 'plans': loadPlans(true); break;
    case 'policy': loadPolicy(); break;
    case 'settings': loadSettings(); break;
    case 'continuity': loadContinuityOverview(); break;
    case 'state': loadStateComponents(); break;
    case 'timelines': loadTimelines(); break;
    case 'logs': loadLogs(); break;
  }
}

function renderCurrent() {
  renderCenter();
  renderRightPanel();
}
