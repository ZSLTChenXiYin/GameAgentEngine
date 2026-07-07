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
  center.appendChild(pageBtn('policy', tr('Policy')));
  center.appendChild(pageBtn('settings', tr('Settings')));
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

  // Build parent-to-children map
  var childMap = {};
  for (var ni = 0; ni < nodes.length; ni++) {
    var n = nodes[ni];
    var pid = n.parent_id || '_root';
    if (!childMap[pid]) childMap[pid] = [];
    childMap[pid].push(n);
  }

  // Compute ancestor path of selected node
  var ancestorSet = {};
  if (state.selectedNodeId) {
    var curId = state.selectedNodeId;
    for (var tries = 0; tries < 100; tries++) {
      var curNode = null;
      for (var i = 0; i < nodes.length; i++) { if (nodes[i].id === curId) { curNode = nodes[i]; break; } }
      if (!curNode || !curNode.parent_id) break;
      ancestorSet[curNode.parent_id] = true;
      curId = curNode.parent_id;
    }
  }

  // Collapse state: treeCollapsed[parentId] = true/false
  if (!state.treeCollapsed) state.treeCollapsed = {};
  if (state.dragNodeId) {
    body.classList.add('drag-active');
  } else {
    body.classList.remove('drag-active');
  }

  var rootDrop = ce('div', { className: 'tree-root-drop' }, [txt('Drop here to move to root')]);
  rootDrop.addEventListener('dragover', function(e) {
    if (!state.dragNodeId) return;
    e.preventDefault();
    this.classList.add('active');
  });
  rootDrop.addEventListener('dragleave', function() {
    this.classList.remove('active');
  });
  rootDrop.addEventListener('drop', function(e) {
    e.preventDefault();
    this.classList.remove('active');
    var srcId = state.dragNodeId || (e.dataTransfer ? e.dataTransfer.getData('text/plain') : '');
    if (!srcId) return;
    moveNodeParent(srcId, '');
    state.dragNodeId = null;
    body.classList.remove('drag-active');
  });
  body.appendChild(rootDrop);

  function renderChildren(parentId, depth, container) {
    var children = childMap[parentId] || [];
    for (var ci = 0; ci < children.length; ci++) {
      (function() {
        var node = children[ci];
        var hasChildren = childMap[node.id] && childMap[node.id].length > 0;
        var isCollapsed = state.treeCollapsed[node.id];
        var isSelected = state.selectedNodeId === node.id;
        var isAncestor = ancestorSet[node.id] && !isSelected;

        var cls = 'tree-node';
        if (isSelected) cls += ' selected';
        if (isAncestor) cls += ' path-ancestor';

        var row = ce('div', { className: cls, dataset: { id: node.id, pid: node.parent_id || '' }, draggable: node.node_type !== 'world', style: { paddingLeft: (12 + depth * 16) + 'px' } }, [
          ce('span', { className: 'tree-arrow' + (hasChildren ? (isCollapsed ? '' : ' expanded') : ' invisible') }, [ttxt('\u25b8')]),
          ce('span', { className: 'tree-icon ' + node.node_type }, []),
          ce('span', { className: 'tree-name' }, [txt(node.name)]),
          ce('span', { className: 'tree-type node-type-' + node.node_type }, [txt(node.node_type)]),
        ]);

        row.addEventListener('click', (function(nn, hc) {
          return function(e) {
            if (e.target.classList.contains('tree-arrow') && hc) {
              e.stopPropagation();
              state.treeCollapsed[nn.id] = !state.treeCollapsed[nn.id];
              renderTree();
              return;
            }
            selectNode(nn.id, nn.node_type);
          };
        })(node, hasChildren));
        row.addEventListener('dragstart', (function(nn) {
          return function(e) {
            state.dragNodeId = nn.id;
            e.dataTransfer.setData('text/plain', nn.id);
            e.dataTransfer.effectAllowed = 'move';
            body.classList.add('drag-active');
            this.classList.add('drag-source');
          };
        })(node));
        row.addEventListener('dragend', function() {
          state.dragNodeId = null;
          body.classList.remove('drag-active');
          this.classList.remove('drag-source');
          this.style.outline = '';
          rootDrop.classList.remove('active');
          var activeDrops = body.querySelectorAll('.drop-target');
          activeDrops.forEach(function(item) { item.classList.remove('drop-target'); });
        });
        row.addEventListener('dragover', (function() {
          return function(e) {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
            this.classList.add('drop-target');
          };
        })());
        row.addEventListener('dragleave', function() {
          this.classList.remove('drop-target');
        });
        row.addEventListener('drop', (function(nn) {
          return function(e) {
            e.preventDefault();
            this.classList.remove('drop-target');
            var srcId = state.dragNodeId || (e.dataTransfer ? e.dataTransfer.getData('text/plain') : '');
            if (!srcId || srcId === nn.id) return;
            state.dragNodeId = null;
            body.classList.remove('drag-active');
            moveNodeParent(srcId, nn.id);
          };
        })(node));
        row.addEventListener('contextmenu', (function(nn) {
          return function(e) {
            e.preventDefault();
            e.stopPropagation();
            showContextMenu([
              { label: tr('Edit'), onClick: function() { openEditNodeModal(nn.id); } },
              { label: 'Copy', onClick: function() { openCopyNodeModal(nn.id); } },
              { label: tr('Create Child'), onClick: function() { openCreateNodeModal(nn.id); } },
              { label: tr('Delete'), danger: true, onClick: function() { deleteNodeHandler(nn.id); } },
            ], e.clientX, e.clientY);
          };
        })(node));

        container.appendChild(row);

        if (hasChildren && !isCollapsed) {
          var childCont = ce('div', { className: 'tree-children' }, []);
          container.appendChild(childCont);
          renderChildren(node.id, depth + 1, childCont);
        }
      })();
    }
  }

  var rootContainer = ce('div', {}, []);
  renderChildren('_root', 0, rootContainer);
  while (rootContainer.firstChild) body.appendChild(rootContainer.firstChild);

  if (nodes.length === 0) {
    body.appendChild(ce('div', { className: 'hint' }, [ttxt('No nodes. Click + to create.')]));
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
