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
    const sel = document.getElementById('worldSelector');
    if (sel) {
      const cur = sel.value; sel.innerHTML = '';
      sel.appendChild(el('option', { value: '', textContent: '-- Select World --' }));
      for (const w of state.worlds) sel.appendChild(el('option', { value: w.id, textContent: w.name }));
      if (cur) sel.value = cur;
    }
    if (!state.selectedWorldId && state.worlds.length > 0) {
      state.selectedWorldId = state.worlds[0].id;
      selectWorld(state.selectedWorldId);
    }
  } catch(e) {
    state.worlds = [];
    toast(tr('Failed to load worlds') + ': ' + apiErrorMessage(e), 'error');
  }
}

async function selectWorld(worldId) {
  state.selectedWorldId = worldId;
  if (worldId) {
    try { state.nodes = await api('GET', '/api/v1/nodes?world_id=' + encodeURIComponent(worldId)); }
    catch(e) { state.nodes = []; toast(tr('Failed to load nodes') + ': ' + apiErrorMessage(e), 'error'); }
    state.logs = [];
    loadPolicy(); loadSettings(); loadSnapshots();
    if (state.page === 'logs') loadLogs();
  } else {
    state.nodes = []; state.selectedNodeId = null; state.nodeDetail = null; state.snapshots = []; state.snapshotMeta = null;
    state.snapshotListWorldId = null; state.logs = []; state.settings = null; state.policy = null;
  }
  renderTree(); renderCurrent();
}

async function selectNode(nodeId, nodeType) {
  state.selectedNodeId = nodeId; state.selectedNodeType = nodeType || null; renderTree();
  if (!nodeId) { state.nodeDetail = null; renderCurrent(); return; }
  try { state.nodeDetail = await api('GET', '/api/v1/nodes/' + encodeURIComponent(nodeId)); }
  catch(e) { state.nodeDetail = null; toast(tr('Failed to load node details') + ': ' + apiErrorMessage(e), 'error'); }
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

async function loadLogs() {
  try {
    var q = '/api/v1/logs?limit=100';
    if (state.selectedWorldId) q += '&world_id=' + encodeURIComponent(state.selectedWorldId);
    state.logs = await api('GET', q);
    if (state.page === 'logs') renderCurrent();
  } catch(e) {
    state.logs = [];
    if (state.page === 'logs') renderCurrent();
    toast(tr('Failed to load logs') + ': ' + apiErrorMessage(e), 'error');
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
    renderTree();
    if (state.selectedNodeId) { var sn = state.nodes.find(function(x) { return x.id === state.selectedNodeId; }); selectNode(state.selectedNodeId, sn ? sn.node_type : null); }
  } catch(e) { toast(tr('Refresh failed') + ': ' + apiErrorMessage(e), 'error'); }
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
    toast(tr('Failed: ') + apiErrorMessage(e), 'error');
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
    toast(tr('Failed: ') + apiErrorMessage(e), 'error');
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
    toast(tr('Failed: ') + apiErrorMessage(e), 'error');
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
      toast(tr('Failed: ') + apiErrorMessage(e), 'error');
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
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

async function createWorld() {
  const name = document.getElementById('createWorldName').value.trim();
  if (!name) { toast(tr('Enter a world name'), 'error'); return; }
  try {
    await api('POST', '/api/v1/nodes', { name: name, node_type: 'world' });
    closeModal(); toast(tr('World created'), 'success'); loadWorlds();
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
    toast(tr('Failed: ') + apiErrorMessage(e), 'error');
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
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

async function moveNodeParent(nodeId, newParentId) {
  if (!confirm(tr('Move this node?'))) return;
  try {
    await api('PUT', '/api/v1/nodes/' + encodeURIComponent(nodeId), { parent_id: newParentId });
    toast(tr('Node moved'), 'success'); loadCurrentWorld();
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
    toast(tr('Failed: ') + apiErrorMessage(e), 'error');
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
    } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
  });
  document.getElementById('modalCancelDelNodeBtn').addEventListener('click', closeModal);
}
/* ============= Component/Memory/Relation Add Modals ============= */
function openAddComponentModal() {
  if (!state.selectedNodeId) return;
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'addCompType' }, [ttxt('Component Type')]),
    el('select', { id: 'addCompType', innerHTML: '<option value="profile">profile</option><option value="rule">rule</option><option value="timeline">timeline</option><option value="action_policy">action_policy</option><option value="prompt_profile">prompt_profile</option><option value="lore">lore</option><option value="autonomous">autonomous</option>' }),
    ce('label', { for: 'addCompData' }, [ttxt('Component Data (JSON/Markdown)')]),
    el('textarea', { id: 'addCompData', placeholder: tr('Enter component data...'), rows: 8, style: {width: '100%', fontFamily: 'var(--font-mono)'} }),
  ]);
  openModal(tr('Add Component'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalAddCompBtn' }, [ttxt('Add')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalAddCompBtn').addEventListener('click', addComponent);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function addComponent() {
  const compType = document.getElementById('addCompType').value;
  const data = document.getElementById('addCompData').value.trim();
  if (!data) { toast(tr('Enter component data'), 'error'); return; }
  try {
    await api('POST', '/api/v1/components', { node_id: state.selectedNodeId, component_type: compType, data: data });
    closeModal(); toast(tr('Component added'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

function openAddRelationModal() {
  if (!state.selectedNodeId) return;
  var options = state.nodes.map(function(n) { return '<option value="' + n.id + '">' + n.name + ' (' + n.node_type + ')</option>'; }).join('');
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'addRelTarget' }, [ttxt('Target Node')]),
    el('select', { id: 'addRelTarget', innerHTML: options }),
    ce('label', { for: 'addRelType' }, [ttxt('Relation Type')]),
    el('select', { id: 'addRelType', innerHTML: '<option value="belongs_to">belongs_to</option><option value="ally">ally</option><option value="enemy">enemy</option><option value="subordinate">subordinate</option><option value="kinship">kinship</option><option value="located_at">located_at</option>' }),
    ce('label', { for: 'addRelWeight' }, [ttxt('Weight')]),
    el('input', { id: 'addRelWeight', type: 'number', value: '5', min: '0', max: '100', style: {width: '80px'} }),
  ]);
  openModal(tr('Add Relation'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalAddRelBtn' }, [ttxt('Add')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  document.getElementById('modalAddRelBtn').addEventListener('click', addRelation);
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function addRelation() {
  const targetId = document.getElementById('addRelTarget').value;
  const relType = document.getElementById('addRelType').value;
  const weight = parseInt(document.getElementById('addRelWeight').value) || 5;
  if (!targetId) { toast(tr('Select a target node'), 'error'); return; }
  try {
    await api('POST', '/api/v1/relations', { world_id: state.selectedWorldId, source_id: state.selectedNodeId, target_id: targetId, relation_type: relType, weight: weight });
    closeModal(); toast(tr('Relation added'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  showLoading(tr('Advancing tick...'));
  try {
    const res = await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/ticks/advance', { tick_type: 'hour', game_time: '' });
    hideLoading();
    toast(tr('Tick advanced') + ': ' + (res.tick ? 'tick ' + res.tick.tick_number : tr('Valid')), 'success');
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

async function runAutonomous() { if (!requireBothGuard()) return;
  if (!state.selectedWorldId) { toast(tr('Please select a world first'), 'error'); return; } if (!state.selectedNodeId) { toast(tr('Please select a node first'), 'error'); return; }
  showLoading(tr('Running autonomous...'));
  try {
    await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/nodes/' + encodeURIComponent(state.selectedNodeId) + '/autonomous/run', null);
    hideLoading();
    toast(tr('Autonomous triggered'), 'success');
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

async function savePolicy() {
  const blocked = document.getElementById('policyBlocked').value.split('\n').map(function(s) { return s.trim(); }).filter(Boolean);
  const safe = document.getElementById('policySafe').value.split('\n').map(function(s) { return s.trim(); }).filter(Boolean);
  try {
    await api('PUT', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/policy', { blocked_actions: blocked, safe_actions: safe });
    toast(tr('Policy saved'), 'success'); loadPolicy();
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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

    if (Object.keys(payload).length === 0) {
      toast(tr('Settings saved'), 'success');
      return;
    }

    await api('PUT', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/settings', payload);
    toast(tr('Settings saved'), 'success'); loadSettings();
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
    } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
    } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
    } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
    el('select', { id: 'editCompType', innerHTML: '<option value="profile">profile</option><option value="rule">rule</option><option value="timeline">timeline</option><option value="action_policy">action_policy</option><option value="prompt_profile">prompt_profile</option><option value="lore">lore</option><option value="autonomous">autonomous</option>' }),
    ce('label', { for: 'editCompData' }, [ttxt('Component Data')]),
    el('textarea', { id: 'editCompData', rows: 10, style: {width: '100%', fontFamily: 'var(--font-mono)', fontSize: '11px'}, textContent: comp.data || '' }),
  ]);
  openModal(tr('Edit Component'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditCompBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var ec = document.getElementById('editCompType'); if (ec) ec.value = comp.component_type;
  document.getElementById('modalEditCompBtn').addEventListener('click', function() { editComponent(compId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function editComponent(compId) {
  const compType = document.getElementById('editCompType').value;
  const data = document.getElementById('editCompData').value.trim();
  if (!data) { toast(tr('Enter component data'), 'error'); return; }
  try {
    await api('PUT', '/api/v1/components/' + encodeURIComponent(compId), { component_type: compType, data: data });
    closeModal(); toast(tr('Component updated'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

function openEditRelationModal(relId) {
  const nd = state.nodeDetail;
  if (!nd || !nd.relations) return;
  const rel = nd.relations.find(function(x) { return x.id === relId; });
  if (!rel) return;
  var options = state.nodes.map(function(n) { return '<option value="' + n.id + '">' + n.name + ' (' + n.node_type + ')</option>'; }).join('');
  const f = ce('div', { className: 'modal-field' }, [
    ce('label', { for: 'editRelSource' }, [ttxt('Source Node')]),
    el('select', { id: 'editRelSource', innerHTML: options }),
    ce('label', { for: 'editRelTarget' }, [ttxt('Target Node')]),
    el('select', { id: 'editRelTarget', innerHTML: options }),
    ce('label', { for: 'editRelType' }, [ttxt('Relation Type')]),
    el('select', { id: 'editRelType', innerHTML: '<option value="belongs_to">belongs_to</option><option value="ally">ally</option><option value="enemy">enemy</option><option value="subordinate">subordinate</option><option value="kinship">kinship</option><option value="located_at">located_at</option>' }),
    ce('label', { for: 'editRelWeight' }, [ttxt('Weight')]),
    el('input', { id: 'editRelWeight', type: 'number', value: String(rel.weight || 5), min: '0', max: '100', style: {width: '80px'} }),
  ]);
  openModal(tr('Edit Relation'), f,
    ce('div', {}, [ce('button', { className: 'primary', id: 'modalEditRelBtn' }, [ttxt('Save')]), el('button', { id: 'modalCancelBtn', textContent: tr('Cancel') })])
  );
  var ers = document.getElementById('editRelSource'); if (ers) ers.value = rel.source_id;
  var ert = document.getElementById('editRelTarget'); if (ert) ert.value = rel.target_id;
  var ert2 = document.getElementById('editRelType'); if (ert2) ert2.value = rel.relation_type;
  document.getElementById('modalEditRelBtn').addEventListener('click', function() { editRelation(relId); });
  document.getElementById('modalCancelBtn').addEventListener('click', closeModal);
}

async function editRelation(relId) {
  const sourceId = document.getElementById('editRelSource').value;
  const targetId = document.getElementById('editRelTarget').value;
  const relType = document.getElementById('editRelType').value;
  const weight = parseInt(document.getElementById('editRelWeight').value) || 5;
  if (!sourceId || !targetId) { toast(tr('Select source and target'), 'error'); return; }
  try {
    await api('PUT', '/api/v1/relations/' + encodeURIComponent(relId), { source_id: sourceId, target_id: targetId, relation_type: relType, weight: weight });
    closeModal(); toast(tr('Relation updated'), 'success'); selectNode(state.selectedNodeId);
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

async function scopeAdvance() { if (!requireBothGuard()) return; if (state.selectedNodeType === 'world') { toast(tr('Scope Advance requires a non-world node'), 'error'); return; }
  if (!state.selectedNodeId) { toast(tr('Select a node as scope'), 'error'); return; }
  if (!state.selectedWorldId) { toast(tr('Select a world'), 'error'); return; }
  showLoading(tr('Advancing scope...'));
  try {
    await api('POST', '/api/v1/worlds/' + encodeURIComponent(state.selectedWorldId) + '/scopes/' + encodeURIComponent(state.selectedNodeId) + '/advance', null);
    hideLoading();
    toast(tr('Scope advanced'), 'success');
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { hideLoading(); toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
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
  } catch(e) { toast(tr('Failed: ') + apiErrorMessage(e), 'error'); }
}

