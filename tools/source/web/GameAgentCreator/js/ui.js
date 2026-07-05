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
  center.appendChild(pageBtn('policy', tr('Policy')));
  center.appendChild(pageBtn('settings', tr('Settings')));
  center.appendChild(pageBtn('logs', tr('Logs')))
  center.appendChild(pageBtn('traces', tr('Traces')));
  const right = ce('div', { className: 'topbar-right' }, [
    el('span', { id: 'connStatus', className: 'status', innerHTML: '<span class="status-dot off"></span> Disconnected' }),
    ce('button', { id: 'btnConfig', title: 'Config' }, [ttxt('\u2699')]),
  ]);
  topbar.appendChild(left);
  const rightSec = ce('div', { className: 'topbar-right' }, [ce('button', { className: 'lang-btn', onclick: toggleLang, title: 'Switch Language' }, [ce('span', { className: 'lang-icon' }, [txt(lang === 'zh' ? 'EN' : '中')])]),
    ce('button', { className: 'theme-btn', onclick: toggleTheme, title: 'Toggle Theme' }, [txt(theme === 'dark' ? '\u2600' : '\u2601')])]);
  topbar.appendChild(rightSec); topbar.appendChild(center); topbar.appendChild(right);
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
    el('input', { id: 'treeFilter', placeholder: 'Filter nodes...', value: state.treeFilter }),
    ce('button', { id: 'btnAddNode' }, [ttxt('+')]),
  ]);
  lp.appendChild(tb);
  const body = el('div', { id: 'treeBody', className: 'tree-body' });
  lp.appendChild(body);
  document.getElementById('btnAddNode').addEventListener('click', openCreateNodeModal);
  document.getElementById('btnTreeRefresh').addEventListener('click', loadCurrentWorld);
  document.getElementById('treeFilter').addEventListener('input', function() {
    state.treeFilter = this.value; renderTree();
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
      for (var ni = 0; ni < nodes.length; ni++) { if (nodes[ni].id === curId) { curNode = nodes[ni]; break; } }
      if (!curNode || !curNode.parent_id) break;
      ancestorSet[curNode.parent_id] = true;
      curId = curNode.parent_id;
    }
  }
  
  // Collapse state: treeCollapsed[parentId] = true/false
  if (!state.treeCollapsed) state.treeCollapsed = {};
  
  function renderChildren(parentId, depth, container) {
    var children = childMap[parentId] || [];
    for (var ci = 0; ci < children.length; ci++) {
      (function() {
        var n = children[ci];
        var hasChildren = childMap[n.id] && childMap[n.id].length > 0;
        var isCollapsed = state.treeCollapsed[n.id];
        var isSelected = state.selectedNodeId === n.id;
        var isAncestor = ancestorSet[n.id] && !isSelected;
        
        var cls = 'tree-node';
        if (isSelected) cls += ' selected';
        if (isAncestor) cls += ' path-ancestor';
        
        var row = ce('div', { className: cls, dataset: { id: n.id, pid: n.parent_id || '' }, draggable: true, style: { paddingLeft: (12 + depth * 16) + 'px' } }, [
          ce('span', { className: 'tree-arrow' + (hasChildren ? (isCollapsed ? '' : ' expanded') : ' invisible') }, [ttxt('\u25b8')]),
          ce('span', { className: 'tree-icon ' + n.node_type }, []),
          ce('span', { className: 'tree-name' }, [txt(n.name)]),
          ce('span', { className: 'tree-type node-type-' + n.node_type }, [txt(n.node_type)]),
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
        })(n, hasChildren));
        row.addEventListener('dragstart', (function(nn) {
          return function(e) {
            e.dataTransfer.setData('text/plain', nn.id);
            e.dataTransfer.effectAllowed = 'move';
          };
        })(n));
        row.addEventListener('dragover', (function(nn) {
          return function(e) {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
            this.style.outline = '2px dashed var(--accent)';
          };
        })(n));
        row.addEventListener('dragleave', function() {
          this.style.outline = '';
        });
        row.addEventListener('drop', (function(nn) {
          return function(e) {
            e.preventDefault();
            this.style.outline = '';
            var srcId = e.dataTransfer.getData('text/plain');
            if (!srcId || srcId === nn.id) return;
            moveNodeParent(srcId, nn.id);
          };
        })(n));
        row.addEventListener('contextmenu', (function(nn) {
          return function(e) {
            e.preventDefault(); e.stopPropagation();
            showContextMenu([
              { label: tr('Edit'), onClick: function() { openEditNodeModal(nn.id); } },
              { label: tr('Create Child'), onClick: function() { openCreateNodeModal(nn.id); } },
              { label: tr('Delete'), danger: true, onClick: function() { deleteNodeHandler(nn.id); } },
            ], e.clientX, e.clientY);
          };
        })(n));
        
        container.appendChild(row);
        
        if (hasChildren && !isCollapsed) {
          var childCont = ce('div', { className: 'tree-children' }, []);
          container.appendChild(childCont);
          renderChildren(n.id, depth + 1, childCont);
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


