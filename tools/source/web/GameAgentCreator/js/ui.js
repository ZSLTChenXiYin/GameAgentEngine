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
    ce('button', { id: 'btnConfig', title: tr('Server Config') }, [ttxt('\u2699')]),
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
    ce('button', { className: 'close', id: 'btnTreeRefresh' }, [ttxt('\u21bb')]),
  ]);
  lp.appendChild(hd);
  const tb = ce('div', { className: 'tree-toolbar' }, [
    el('input', { id: 'treeFilter', placeholder: tr('Filter nodes...'), value: state.treeFilter }),
    ce('button', { id: 'btnAddNode' }, [ttxt('+')]),
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
    renderTree();
  });
}

/* ============= Tree ============= */
function renderTree() {
  const body = document.getElementById('treeBody');
  if (!body) return; body.innerHTML = '';
  const filter = state.treeFilter.toLowerCase();
  let nodes = state.nodes;
  if (filter) nodes = nodes.filter(function(n) { return n.name.toLowerCase().includes(filter) || n.node_type.includes(filter); });
  var selectedSet = {};
  (state.selectedNodeIds || []).forEach(function(id) { selectedSet[id] = true; });
  var visibleIds = [];
  var projected = buildProjectedChildMap(nodes);
  var childMap = projected.childMap;
  var nodeMap = {};
  nodes.forEach(function(n) { nodeMap[n.id] = n; });
  var activePathSet = {};
  if (state.selectedTreePathKey) {
    state.selectedTreePathKey.split('|').forEach(function(key) {
      if (key) activePathSet[key] = true;
    });
  }

  // Collapse state: treeCollapsed[parentId] = true/false
  if (!state.treeCollapsed) state.treeCollapsed = {};
  if (state.dragNodeId) {
    body.classList.add('drag-active');
  } else {
    body.classList.remove('drag-active');
  }

  function clearDropIndicators() {
    rootDrop.classList.remove('active');
    var activeDrops = body.querySelectorAll('.drop-target');
    activeDrops.forEach(function(item) { item.classList.remove('drop-target'); });
  }

  function cleanupPointerDrag(sourceRow) {
    state.dragNodeId = null;
    body.classList.remove('drag-active');
    clearDropIndicators();
    if (sourceRow) sourceRow.classList.remove('drag-source');
    document.body.style.userSelect = '';
    document.body.style.cursor = '';
  }

  var rootDrop = ce('div', { className: 'tree-root-drop' }, [txt(tr('Drop here to move to root'))]);
  body.appendChild(rootDrop);
  var treeContent = ce('div', { className: 'tree-content' }, []);
  body.appendChild(treeContent);

  function renderChildren(parentId, depth, container, ancestorIds, parentPathKey) {
    var children = childMap[parentId] || [];
    for (var ci = 0; ci < children.length; ci++) {
      (function() {
        var nodeId = children[ci];
        var node = nodeMap[nodeId];
        if (!node) return;
        if (ancestorIds.indexOf(nodeId) >= 0) return;
        var occurrenceKey = parentPathKey ? parentPathKey + '|' + nodeId : nodeId;
        var nextAncestorIds = ancestorIds.concat([nodeId]);
        var hasChildren = childMap[node.id] && childMap[node.id].length > 0;
        var isCollapsed = state.treeCollapsed[node.id];
        var isSelected = !!selectedSet[node.id] && state.selectedTreePathKey === occurrenceKey;
        var isPrimarySelected = state.selectedNodeId === node.id;
        var isAliasSelected = !!selectedSet[node.id] && state.selectedTreePathKey && state.selectedTreePathKey !== occurrenceKey;
        var isAncestor = !!activePathSet[occurrenceKey] && !isSelected;

        var cls = 'tree-node';
        if (isSelected) cls += ' selected';
        if (isSelected && !isPrimarySelected) cls += ' multi-selected';
        if (isAliasSelected) cls += ' alias-selected';
        if (isAncestor) cls += ' path-ancestor';

        var row = ce('div', { className: cls, dataset: { id: node.id, pid: node.parent_id || '' }, style: { paddingLeft: (12 + depth * 16) + 'px' } }, [
          ce('span', { className: 'tree-arrow' + (hasChildren ? (isCollapsed ? '' : ' expanded') : ' invisible') }, [ttxt('\u25b8')]),
          ce('span', { className: 'tree-icon ' + node.node_type }, []),
          ce('span', { className: 'tree-name' }, [txt(node.name)]),
          ce('span', { className: 'tree-type node-type-' + node.node_type }, [txt(node.node_type)]),
        ]);

        visibleIds.push(node.id);

        row.addEventListener('click', (function(nn, hc) {
          return function(e) {
            if (state.suppressTreeClickUntil && Date.now() < state.suppressTreeClickUntil) return;
            if (e.target.classList.contains('tree-arrow') && hc) {
              e.stopPropagation();
              state.treeCollapsed[nn.id] = !state.treeCollapsed[nn.id];
              renderTree();
              return;
            }
            if (e.shiftKey) {
              selectNode(nn.id, nn.node_type, { mode: 'range', preserveAnchor: true, treePathKey: occurrenceKey });
              return;
            }
            if (e.ctrlKey || e.metaKey) {
              selectNode(nn.id, nn.node_type, { mode: 'toggle', treePathKey: occurrenceKey });
              return;
            }
            selectNode(nn.id, nn.node_type, { mode: 'single', treePathKey: occurrenceKey });
          };
        })(node, hasChildren));

        row.addEventListener('mousedown', (function(nn) {
          return function(e) {
            if (e.button !== 0) return;
            if (nn.node_type === 'world') return;
            if (e.target.classList.contains('tree-arrow')) return;

            var sourceRow = this;
            var startX = e.clientX;
            var startY = e.clientY;
            var started = false;
            var dropTargetId = null;

            function onMove(ev) {
              var dx = Math.abs(ev.clientX - startX);
              var dy = Math.abs(ev.clientY - startY);
              if (!started) {
                if (Math.max(dx, dy) < 5) return;
                started = true;
                state.dragNodeId = nn.id;
                body.classList.add('drag-active');
                sourceRow.classList.add('drag-source');
                document.body.style.userSelect = 'none';
                document.body.style.cursor = 'grabbing';
              }

              ev.preventDefault();
              clearDropIndicators();
              dropTargetId = null;

              var hit = document.elementFromPoint(ev.clientX, ev.clientY);
              if (!hit) return;

              var rootHit = hit === rootDrop || (hit.closest && hit.closest('.tree-root-drop') === rootDrop);
              if (rootHit) {
                rootDrop.classList.add('active');
                dropTargetId = '';
                return;
              }

              var targetRow = hit.closest ? hit.closest('.tree-node') : null;
              if (!targetRow) return;
              if (targetRow.dataset.id === nn.id) return;
              targetRow.classList.add('drop-target');
              dropTargetId = targetRow.dataset.id || null;
            }

            function onUp(ev) {
              window.removeEventListener('mousemove', onMove, true);
              window.removeEventListener('mouseup', onUp, true);
              if (!started) {
                cleanupPointerDrag(sourceRow);
                return;
              }

              ev.preventDefault();
              state.suppressTreeClickUntil = Date.now() + 250;
              cleanupPointerDrag(sourceRow);

              if (dropTargetId === null) return;
              moveNodeParent(nn.id, dropTargetId);
            }

            window.addEventListener('mousemove', onMove, true);
            window.addEventListener('mouseup', onUp, true);
          };
        })(node));
        row.addEventListener('contextmenu', (function(nn) {
          return async function(e) {
            e.preventDefault();
            e.stopPropagation();
            if (!selectedSet[nn.id] || state.selectedTreePathKey !== occurrenceKey) await selectNode(nn.id, nn.node_type, { mode: 'single', treePathKey: occurrenceKey });
            showContextMenu([
              { label: tr('Edit'), onClick: function() { openEditNodeModal(nn.id); } },
              { label: tr('Copy Node'), onClick: function() { openCopyNodeModal(nn.id); } },
              { label: tr('Add New Parent'), tip: tr('Creates a new node and rewires only parent_id'), onClick: function() { openCreateParentNodeModal(nn.id); } },
              { label: tr('Add Outgoing Relation'), tip: tr('Relations are stored separately from the outline tree'), onClick: function() { openAddOutgoingRelationModal(nn.id); } },
              { label: tr('Create Child'), onClick: function() { openCreateNodeModal(nn.id); } },
              { label: tr('Delete'), danger: true, onClick: function() { deleteNodeHandler(nn.id); } },
            ], e.clientX, e.clientY);
          };
        })(node));

        container.appendChild(row);

        if (hasChildren && !isCollapsed) {
          var childCont = ce('div', { className: 'tree-children' }, []);
          container.appendChild(childCont);
          renderChildren(node.id, depth + 1, childCont, nextAncestorIds, occurrenceKey);
        }
      })();
    }
  }

  var rootContainer = ce('div', {}, []);
  renderChildren('_root', 0, rootContainer, [], '');
  while (rootContainer.firstChild) treeContent.appendChild(rootContainer.firstChild);
  state.visibleNodeIds = visibleIds;

  if (nodes.length === 0) {
    treeContent.appendChild(ce('div', { className: 'hint' }, [ttxt('No nodes. Click + to create.')]));
  }
}


/* ============= Switch Page & Render ============= */
function switchPage(name) {
  state.page = name;
  var tabs = document.querySelectorAll('.topbar-center button');
  tabs.forEach(function(b) { b.classList.remove('active'); });
  var found = Array.from(tabs).find(function(b) { return b.dataset.page === name; });
  if (found) found.classList.add('active');
  renderCurrent();
}

function renderCurrent() {
  renderCenter();
  renderRightPanel();
}
