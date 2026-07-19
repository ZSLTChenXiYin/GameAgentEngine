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
  body.tabIndex = 0;
  body.addEventListener("keydown", function(e) {
    if (!state.visibleNodeIds || state.visibleNodeIds.length === 0) return;
    var idx = state.selectedNodeId ? state.visibleNodeIds.indexOf(state.selectedNodeId) : -1;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      idx = Math.min(idx + 1, state.visibleNodeIds.length - 1);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      idx = Math.max(idx - 1, 0);
    } else if (e.key === "ArrowRight") {
      if (state.selectedNodeId && state.treeCollapsed && state.treeCollapsed[state.selectedNodeId]) {
        e.preventDefault(); state.treeCollapsed[state.selectedNodeId] = false; invalidateTreeCache(); renderTree();
      }
      return;
    } else if (e.key === "ArrowLeft") {
      if (state.selectedNodeId) {
        e.preventDefault(); state.treeCollapsed = state.treeCollapsed || {}; state.treeCollapsed[state.selectedNodeId] = true; invalidateTreeCache(); renderTree();
      }
      return;
    } else if (e.key === "Enter") {
      if (state.selectedNodeId) { e.preventDefault(); editNodeHandler(state.selectedNodeId); }
      return;
    } else { return; }
    if (idx < 0 || idx >= state.visibleNodeIds.length) return;
    var nid = state.visibleNodeIds[idx];
    var rows = _treeCache.flatRows || buildFlatRows();
    var pk = "";
    for (var ri = 0; ri < rows.length; ri++) { if (rows[ri].nodeId === nid) { pk = rows[ri].pathKey; break; } }
    selectNode(nid, pk); renderCurrent(); renderTree();
  });
}

var ROW_HEIGHT = 28;
var BUFFER = 5;

function renderTree() {
  var body = document.getElementById("treeBody");
  if (!body) return;
  var savedScrollTop = body.scrollTop;
  body.innerHTML = "";
  state.visibleNodeIds = [];

  applyLargeTreeDegradation();
  var rows = buildFlatRows();
  if (rows.length === 0) {
    body.appendChild(ce("div", { className: "hint" }, [ttxt("No nodes. Click + to create.")]));
    return;
  }

  // Root drop zone — always visible
  var rootDrop = ce("div", { className: "tree-root-drop" }, [txt(tr("Drop here to move to root"))]);
  body.appendChild(rootDrop);

  // Virtual-scroll spacer + viewport
  var totalHeight = rows.length * ROW_HEIGHT + 24;
  var spacer = ce("div", { style: { height: totalHeight + "px", position: "relative" } }, []);
  body.appendChild(spacer);
  var viewport = ce("div", { className: "tree-content", style: { position: "absolute", top: "0", left: "0", right: "0" } }, []);
  spacer.appendChild(viewport);

  var selectedSet = {};
  (state.selectedNodeIds || []).forEach(function (id) { selectedSet[id] = true; });
  var activePathSet = {};
  if (state.selectedTreePathKey) {
    state.selectedTreePathKey.split("|").forEach(function (key) { if (key) activePathSet[key] = true; });
  }

  if (state.dragNodeId) { body.classList.add("drag-active"); }
  else { body.classList.remove("drag-active"); }

  function clearDropIndicators() {
    rootDrop.classList.remove("active");
    var activeDrops = body.querySelectorAll(".drop-target");
    activeDrops.forEach(function (item) { item.classList.remove("drop-target"); });
  }

  body.scrollTop = Math.min(savedScrollTop, (rows.length * ROW_HEIGHT) + 24);

  function renderSlice() {
    var st = body.scrollTop;
    var ch = body.clientHeight || 300;
    var startIdx = Math.max(0, Math.floor(st / ROW_HEIGHT) - BUFFER);
    var endIdx = Math.min(rows.length, Math.ceil((st + ch) / ROW_HEIGHT) + BUFFER);

    viewport.innerHTML = "";
    viewport.style.top = (startIdx * ROW_HEIGHT) + "px";
    state.visibleNodeIds = rows.slice(startIdx, endIdx).map(function (r) { return r.nodeId; });

    for (var ri = startIdx; ri < endIdx; ri++) {
      var r = rows[ri];
      var node = _treeCache.nodeMap[r.nodeId];
      if (!node) continue;

      var isSelected = !!selectedSet[node.id] && state.selectedTreePathKey === r.pathKey;
      var isPrimarySelected = state.selectedNodeId === node.id;
      var isAliasSelected = !!selectedSet[node.id] && state.selectedTreePathKey && state.selectedTreePathKey !== r.pathKey;
      var isAncestor = !!activePathSet[r.pathKey] && !isSelected;

      var classes = ["tree-node"];
      if (isSelected) classes.push("selected");
      if (isPrimarySelected) classes.push("primary-selected");
      if (isAliasSelected) classes.push("alias-selected");
      if (isAncestor) classes.push("ancestor");
      if (r.nodeId === state.dragNodeId) classes.push("dragging");

      var rowEl = ce("div", { className: classes.join(" "), id: "tn_" + r.nodeId, style: { paddingLeft: r.paddingLeft + "px" } }, []);
      rowEl.dataset.nodeId = r.nodeId;
      rowEl.dataset.pathKey = r.pathKey || "";

      // Expand / collapse arrow
      var arrow = ce("span", { className: "tree-arrow" + (r.hasChildren ? "" : " invisible") }, [txt(r.isExpanded ? "\u25BC" : "\u25B6")]);
      arrow.addEventListener("click", function (ev) { ev.stopPropagation(); toggleNode(this.parentElement.dataset.nodeId); renderTree(); });
      rowEl.appendChild(arrow);

      // Type icon + name
      var icon = ce("span", { className: "node-icon node-type-" + (node.node_type || "unknown") }, []);
      rowEl.appendChild(icon);
      var nameSpan = ce("span", { className: "node-name" }, [txt(node.name || node.id)]);
      rowEl.appendChild(nameSpan);

      // Click: select node
      rowEl.addEventListener("click", function () {
        var nid = this.dataset.nodeId;
        var pk = this.dataset.pathKey;
        selectNode(nid, pk);
        renderCurrent();
        renderTree();
      });

      // Context menu
      rowEl.addEventListener("contextmenu", function (e) {
        e.preventDefault();
        e.stopPropagation();
        var nid = this.dataset.nodeId;
        if (state.selectedNodeId !== nid) { selectNode(nid, this.dataset.pathKey); renderCurrent(); }
        showContextMenu([
          { label: tr("Add Child"), onClick: function() { addNodeHandler(nid); } },
          { label: tr("Edit"), onClick: function() { editNodeHandler(nid); } },
          { label: tr("Duplicate"), onClick: function() { duplicateNodeHandler(nid); } },
          { label: tr("Delete"), danger: true, onClick: function() { deleteNodeHandler(nid); } },
        ], e.clientX, e.clientY);
      });

      // Drag / drop
      rowEl.draggable = true;
      rowEl.addEventListener("dragstart", function (e) {
        e.dataTransfer.setData("text/plain", this.dataset.nodeId);
        state.dragNodeId = this.dataset.nodeId;
        body.classList.add("drag-active");
        renderTree();
      });
      rowEl.addEventListener("dragend", function () {
        state.dragNodeId = null;
        body.classList.remove("drag-active");
        clearDropIndicators();
      });
      rowEl.addEventListener("dragover", function (e) {
        e.preventDefault();
        clearDropIndicators();
        this.classList.add("drop-target");
      });
      rowEl.addEventListener("dragleave", function () { this.classList.remove("drop-target"); });
      rowEl.addEventListener("drop", function (e) {
        e.preventDefault();
        clearDropIndicators();
        var srcId = e.dataTransfer.getData("text/plain");
        var tgtId = this.dataset.nodeId;
        if (srcId && tgtId && srcId !== tgtId) moveNode(srcId, tgtId);
      });

      viewport.appendChild(rowEl);
    }
  }

  renderSlice();
  body.onscroll = function () { renderSlice(); };
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
