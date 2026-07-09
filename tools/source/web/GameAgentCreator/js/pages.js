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
      if ((state.selectedNodeIds || []).length > 1) {
        selBar.appendChild(ce('span', { className: 'count' }, [txt(String(state.selectedNodeIds.length) + ' ' + tr('nodes selected'))]));
      }
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
    case 'snapshots': renderSnapshotsPage(center); break;
    case 'plans': renderPlansPage(center); break;
    case 'policy': renderPolicyPage(center); break;
    case 'settings': renderSettingsPage(center); break;
    case 'continuity': renderContinuityPage(center); break;
    case 'state': renderStatePage(center); break;
    case 'timelines': renderTimelinesPage(center); break;
    case 'logs': renderLogsPage(center); break;
    case 'traces': renderTracesPage(center); break;
    default: renderWorldsPage(center);
  }
}

function snapshotStatusColor(status) {
  switch (status) {
    case 'valid': return 'var(--green)';
    case 'invalid': return 'var(--red)';
    case 'working_copy': return 'var(--accent)';
    case 'restored_copy': return '#f59e0b';
    default: return 'var(--text-dim)';
  }
}

function snapshotReasonLabel(reason) {
  return tr(reason || 'unknown');
}

function snapshotStatusLabel(status) {
  return tr(status || 'unknown');
}

function planStatusColor(status) {
  switch (status) {
    case 'approved': return 'var(--green)';
    case 'rejected': return 'var(--red)';
    case 'pending': return '#f59e0b';
    default: return 'var(--text-dim)';
  }
}

function renderSnapshotsPage(container) {
  var selectedWorld = state.worlds.find(function(x) { return x.id === state.selectedWorldId; });
  var listWorld = state.worlds.find(function(x) { return x.id === state.snapshotListWorldId; });
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnRefreshSnapshots' }, [ttxt('Refresh Snapshots')]),
    ce('button', { id: 'btnCreateSnapshotFromPage' }, [ttxt('Save Snapshot')]),
  ]);
  container.appendChild(toolbar);

  if (!state.selectedWorldId) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a world first.')]));
    return;
  }

  if (state.snapshotMeta) {
    const meta = state.snapshotMeta;
    const metaCard = ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Current World Snapshot Metadata')]),
      ce('div', { className: 'card-bd' }, [
        ce('div', { className: 'hint' }, [ttxt(meta.reason === 'save_snapshot' ? 'Current selection is a saved snapshot world.' : 'Current selection is a copied world with snapshot lineage metadata.')]),
        renderPropRow('Snapshot Name', meta.snapshot_name || '-'),
        renderPropRow('Reason', meta.reason ? snapshotReasonLabel(meta.reason) : '-'),
        renderPropRow('Status', meta.status ? snapshotStatusLabel(meta.status) : '-'),
        renderPropRow('Restorable', meta.restorable ? tr('Yes') : tr('No')),
        renderPropRow('Source World', meta.source_world_id || '-', { mono: true }),
        renderPropRow('Validation Issues', (meta.validation_issues || []).join(', ') || '-'),
        ce('div', { className: 'policy-actions' }, [
          meta.source_world_id && meta.source_world_id !== state.selectedWorldId ? ce('button', { id: 'btnOpenSnapshotSourceWorld' }, [ttxt('Open Source World')]) : null,
          ce('button', { id: 'btnValidateCurrentSnapshot' }, [ttxt('Validate')]),
          meta.reason === 'save_snapshot' ? ce('button', { id: 'btnRestoreCurrentSnapshot' }, [ttxt('Restore')]) : null,
          meta.reason === 'save_snapshot' ? ce('button', { id: 'btnDeleteCurrentSnapshot', className: 'danger' }, [ttxt('Delete')]) : null,
        ]),
      ]),
    ]);
    container.appendChild(metaCard);
  }

  var listTitle = tr('Saved Snapshots');
  if (state.snapshotMeta && state.snapshotMeta.reason === 'save_snapshot') {
    listTitle = tr('Saved Snapshots of Source World');
  }

  const listCard = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd' }, [txt(listTitle)]),
    ce('div', { className: 'card-bd', id: 'snapshotListBody' }, [
      ce('div', { className: 'hint' }, [txt(state.snapshotMeta && state.snapshotMeta.reason === 'save_snapshot'
        ? tr('Showing all save snapshots created from the source world of the current snapshot.')
        : tr('Showing all save snapshots created from the currently selected world.'))]),
      (state.snapshotListWorldId || listWorld) ? renderPropRow('List World', (listWorld ? listWorld.name + ' ' : '') + '(' + (state.snapshotListWorldId || '-') + ')') : null,
    ]),
  ]);
  container.appendChild(listCard);

  const listBody = document.getElementById('snapshotListBody');
  const snapshots = state.snapshots || [];
  if (snapshots.length === 0) {
    listBody.appendChild(ce('div', { className: 'hint' }, [ttxt('No saved snapshots for this world yet.')]));
  } else {
    for (var i = 0; i < snapshots.length; i++) {
      (function() {
        var snap = snapshots[i];
        var statusColor = snapshotStatusColor(snap.status);
        var row = ce('div', { className: 'detail-item' }, [
          ce('div', { className: 'item-hd' }, [
            ce('span', { className: 'item-tag', style: { background: 'rgba(59,130,246,.12)', color: statusColor } }, [txt(snapshotStatusLabel(snap.status))]),
            ce('span', { style: { fontWeight: 600 } }, [txt(' ' + (snap.snapshot_name || snap.snapshot_world_id || '-'))]),
            ce('span', { style: { fontSize: '10px', color: 'var(--text-dim)' } }, [txt(' ' + (snap.created_at ? new Date(snap.created_at).toLocaleString() : ''))]),
          ]),
          ce('div', { className: 'item-body' }, [
            renderPropRow('ID', snap.snapshot_world_id || '-', { mono: true }),
            renderPropRow('Counts', 'N ' + (snap.node_count || 0) + ' / C ' + (snap.component_count || 0) + ' / M ' + (snap.memory_count || 0) + ' / R ' + (snap.relation_count || 0)),
            renderPropRow('Issues', (snap.validation_issues || []).join(', ') || '-'),
            ce('div', { className: 'policy-actions' }, [
              ce('button', { className: 'snapshot-validate-btn', dataset: { id: snap.snapshot_world_id } }, [ttxt('Validate')]),
              ce('button', { className: 'snapshot-restore-btn', dataset: { id: snap.snapshot_world_id, name: snap.snapshot_name || '' }, disabled: !snap.restorable }, [ttxt('Restore')]),
              ce('button', { className: 'snapshot-delete-btn danger', dataset: { id: snap.snapshot_world_id } }, [ttxt('Delete')]),
            ]),
          ]),
        ]);
        listBody.appendChild(row);
      })();
    }
  }

  document.getElementById('btnRefreshSnapshots').addEventListener('click', refreshSnapshots);
  document.getElementById('btnCreateSnapshotFromPage').addEventListener('click', saveSnapshot);

  var currentValidate = document.getElementById('btnValidateCurrentSnapshot');
  if (currentValidate && state.snapshotMeta) {
    currentValidate.addEventListener('click', function() { validateSnapshot(state.snapshotMeta.snapshot_world_id); });
  }
  var openSourceWorld = document.getElementById('btnOpenSnapshotSourceWorld');
  if (openSourceWorld && state.snapshotMeta && state.snapshotMeta.source_world_id) {
    openSourceWorld.addEventListener('click', function() { selectWorld(state.snapshotMeta.source_world_id); });
  }
  var currentRestore = document.getElementById('btnRestoreCurrentSnapshot');
  if (currentRestore && state.snapshotMeta) {
    currentRestore.addEventListener('click', function() { openRestoreSnapshotModal(state.snapshotMeta.snapshot_world_id, state.snapshotMeta.snapshot_name || ''); });
  }
  var currentDelete = document.getElementById('btnDeleteCurrentSnapshot');
  if (currentDelete && state.snapshotMeta) {
    currentDelete.addEventListener('click', function() { deleteSnapshot(state.snapshotMeta.snapshot_world_id); });
  }

  var validateBtns = document.querySelectorAll('.snapshot-validate-btn');
  validateBtns.forEach(function(btn) {
    btn.addEventListener('click', function() { validateSnapshot(btn.dataset.id); });
  });
  var restoreBtns = document.querySelectorAll('.snapshot-restore-btn');
  restoreBtns.forEach(function(btn) {
    btn.addEventListener('click', function() { openRestoreSnapshotModal(btn.dataset.id, btn.dataset.name || ''); });
  });
  var deleteBtns = document.querySelectorAll('.snapshot-delete-btn');
  deleteBtns.forEach(function(btn) {
    btn.addEventListener('click', function() { deleteSnapshot(btn.dataset.id); });
  });
}

function renderPlansPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnRefreshPlans' }, [ttxt('Refresh Plans')]),
  ]);
  container.appendChild(toolbar);

  if (!state.selectedWorldId) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a world first.')]));
    return;
  }

  const plans = state.plans || [];
  if (plans.length === 0) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('No pending plans for this world.')]));
  } else {
    for (var i = 0; i < plans.length; i++) {
      (function() {
        var plan = plans[i];
        var impact = plan.world_change_plan ? plan.world_change_plan.impact_level : '-';
        var summary = plan.world_change_plan ? plan.world_change_plan.summary : '-';
        var actionCount = plan.action_calls ? plan.action_calls.length : 0;
        var memoryCount = plan.memory_updates ? plan.memory_updates.length : 0;
        var row = ce('div', { className: 'detail-card' }, [
          ce('div', { className: 'card-hd' }, [
            ce('span', { className: 'item-tag', style: { background: 'rgba(59,130,246,.12)', color: planStatusColor(plan.status) } }, [txt(tr(plan.status || 'pending'))]),
            ce('span', { style: { fontWeight: 600, marginLeft: '8px' } }, [txt(summary || '-')]),
          ]),
          ce('div', { className: 'card-bd' }, [
            renderPropRow('Plan ID', plan.plan_id || '-', { mono: true }),
            renderPropRow('Task Type', plan.task_type || '-'),
            renderPropRow('Impact', impact || '-'),
            renderPropRow('Created At', plan.created_at ? new Date(plan.created_at).toLocaleString() : '-'),
            renderPropRow('Actions', String(actionCount)),
            renderPropRow('Memory Updates', String(memoryCount)),
            renderPropRow('Summary', summary || '-'),
            ce('div', { className: 'policy-actions' }, [
              ce('button', { className: 'primary plan-approve-btn', dataset: { planId: plan.plan_id, worldId: plan.world_id } }, [ttxt('Approve')]),
              ce('button', { className: 'danger plan-reject-btn', dataset: { planId: plan.plan_id, worldId: plan.world_id } }, [ttxt('Reject')]),
            ]),
          ]),
        ]);
        container.appendChild(row);
      })();
    }
  }

  document.getElementById('btnRefreshPlans').addEventListener('click', loadPlans);
  document.querySelectorAll('.plan-approve-btn').forEach(function(btn) {
    btn.addEventListener('click', function() { approvePlan(btn.dataset.worldId, btn.dataset.planId); });
  });
  document.querySelectorAll('.plan-reject-btn').forEach(function(btn) {
    btn.addEventListener('click', function() { rejectPlan(btn.dataset.worldId, btn.dataset.planId); });
  });
}

/* ============= Worlds Page ============= */
function renderWorldsPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnEditWorld' }, [ttxt('Edit World')]),
    ce('button', { id: 'btnForkWorld' }, [ttxt('Create Working Copy')]),
    ce('button', { id: 'btnSaveSnapshot' }, [ttxt('Save Snapshot')]),
    ce('button', { id: 'btnTickAdvance' }, [ttxt('Advance Tick')]),
    ce('button', { id: 'btnAutonomous' }, [ttxt('Run Autonomous')]),
    ce('button', { id: 'btnAutonomousConfig' }, [ttxt('Autonomous Config')]),
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
  document.getElementById('btnEditWorld').addEventListener('click', openEditWorldModal);
  document.getElementById('btnForkWorld').addEventListener('click', forkWorld);
  document.getElementById('btnSaveSnapshot').addEventListener('click', saveSnapshot);
  document.getElementById('btnTickAdvance').addEventListener('click', tickAdvance);
  document.getElementById('btnAutonomous').addEventListener('click', runAutonomous);
  document.getElementById('btnAutonomousConfig').addEventListener('click', openAutonomousConfigModal);
  document.getElementById('btnEventImpact').addEventListener('click', openEventImpactModal);
  document.getElementById('btnScopeAdvance').addEventListener('click', scopeAdvance);
  document.getElementById('btnReplan').addEventListener('click', timelineReplan);
  updateActionButtons();
}

/* ============= Node Detail ============= */
function relationDirectionLabel(nodeId, relation) {
  if (!relation) return '';
  if (relation.source_id === nodeId && relation.target_id === nodeId) return tr('Self');
  if (relation.source_id === nodeId) return tr('Outgoing');
  if (relation.target_id === nodeId) return tr('Incoming');
  return tr('Linked');
}

function relationDirectionArrow(nodeId, relation) {
  if (!relation) return '->';
  if (relation.source_id === nodeId && relation.target_id === nodeId) return '<>'; 
  if (relation.source_id === nodeId) return '->';
  if (relation.target_id === nodeId) return '<-';
  return '--';
}

function relationPeerId(nodeId, relation) {
  if (!relation) return '';
  if (relation.source_id === nodeId && relation.target_id === nodeId) return nodeId;
  if (relation.source_id === nodeId) return relation.target_id;
  return relation.source_id;
}

function relationPeerName(nodeId, relation) {
  var peerId = relationPeerId(nodeId, relation);
  return getNodeNameById(peerId);
}

function relationBlueprintTone(relationType) {
  switch (relationType) {
    case 'external_parent': return { bg: 'rgba(59,130,246,.12)', color: '#60a5fa' };
    case 'subordinate': return { bg: 'rgba(14,165,233,.12)', color: '#38bdf8' };
    case 'belongs_to': return { bg: 'rgba(16,185,129,.12)', color: '#34d399' };
    case 'located_at': return { bg: 'rgba(245,158,11,.12)', color: '#fbbf24' };
    case 'ally': return { bg: 'rgba(34,197,94,.12)', color: '#4ade80' };
    case 'enemy': return { bg: 'rgba(239,68,68,.12)', color: '#f87171' };
    case 'kinship': return { bg: 'rgba(168,85,247,.12)', color: '#c084fc' };
    default: return { bg: 'rgba(148,163,184,.12)', color: 'var(--text-dim)' };
  }
}

function buildRelationBlueprintCard(node, relations) {
  var wrapper = ce('div', { className: 'relation-blueprint' }, []);
  var inbound = [];
  var outbound = [];
  var selfLoop = [];
  (relations || []).forEach(function(rel) {
    if (!rel) return;
    if (rel.source_id === node.id && rel.target_id === node.id) {
      selfLoop.push(rel);
      return;
    }
    if (rel.target_id === node.id) inbound.push(rel);
    if (rel.source_id === node.id) outbound.push(rel);
  });

  function buildLane(title, items, emptyLabel) {
    var lane = ce('div', { className: 'relation-lane' }, [
      ce('div', { className: 'relation-lane-title' }, [txt(title + ' (' + items.length + ')')]),
    ]);
    if (!items.length) {
      lane.appendChild(ce('div', { className: 'relation-empty' }, [txt(emptyLabel)]));
      return lane;
    }
    items.forEach(function(rel) {
      var tone = relationBlueprintTone(rel.relation_type);
      var peerId = relationPeerId(node.id, rel);
      var peerName = relationPeerName(node.id, rel);
      lane.appendChild(ce('div', { className: 'relation-chip', style: { background: tone.bg, color: tone.color } }, [
        ce('div', { className: 'relation-chip-top' }, [
          ce('span', { className: 'relation-chip-name' }, [txt(peerName)]),
          ce('span', { className: 'relation-chip-arrow' }, [txt(relationDirectionArrow(node.id, rel))]),
        ]),
        ce('div', { className: 'relation-chip-meta' }, [
          ce('span', {}, [txt(rel.relation_type)]),
          ce('span', {}, [txt(peerId ? peerId.slice(0, 8) : '-')]),
        ]),
      ]));
    });
    return lane;
  }

  wrapper.appendChild(buildLane(tr('Incoming Relations'), inbound, tr('No incoming relations')));

  var center = ce('div', { className: 'relation-center' }, [
    ce('div', { className: 'relation-center-node' }, [
      ce('div', { className: 'relation-center-type' }, [txt(node.node_type)]),
      ce('div', { className: 'relation-center-name' }, [txt(node.name)]),
      ce('div', { className: 'relation-center-id' }, [txt(node.id.slice(0, 8))]),
    ]),
  ]);
  if (selfLoop.length > 0) {
    var loopBox = ce('div', { className: 'relation-self-loop' }, [
      ce('div', { className: 'relation-lane-title' }, [txt(tr('Self Relations') + ' (' + selfLoop.length + ')')]),
    ]);
    selfLoop.forEach(function(rel) {
      var tone = relationBlueprintTone(rel.relation_type);
      loopBox.appendChild(ce('div', { className: 'relation-chip relation-chip-compact', style: { background: tone.bg, color: tone.color } }, [
        ce('div', { className: 'relation-chip-top' }, [ce('span', { className: 'relation-chip-name' }, [txt(rel.relation_type)])]),
        ce('div', { className: 'relation-chip-meta' }, [ce('span', {}, [txt(relationTypeDescription(rel.relation_type) || '-')])]),
      ]));
    });
    center.appendChild(loopBox);
  }
  wrapper.appendChild(center);

  wrapper.appendChild(buildLane(tr('Outgoing Relations'), outbound, tr('No outgoing relations')));
  return wrapper;
}

function renderNodeDetail(container) {
  const nd = state.nodeDetail;
  if (!nd || !nd.node) return;
  const n = nd.node;
  const detail = ce('div', { className: 'node-detail' }, []);
  var outgoingRelations = (nd.relations || []).filter(function(rel) { return rel.source_id === n.id; });
  var externalParents = (nd.relations || []).filter(function(rel) {
    return rel.relation_type === 'external_parent' && rel.source_id === n.id;
  }).map(function(rel) {
    return getNodeNameById(rel.target_id) + ' (' + rel.target_id.slice(0, 8) + ')';
  });
  var locationTargets = outgoingRelations.filter(function(rel) {
    return rel.relation_type === 'located_at';
  }).map(function(rel) {
    return getNodeNameById(rel.target_id) + ' (' + rel.target_id.slice(0, 8) + ')';
  });
  var organizationTargets = outgoingRelations.filter(function(rel) {
    return rel.relation_type === 'belongs_to' || rel.relation_type === 'subordinate';
  }).map(function(rel) {
    return rel.relation_type + ': ' + getNodeNameById(rel.target_id) + ' (' + rel.target_id.slice(0, 8) + ')';
  });
  
  // Overview card
  const overview = ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd toggle-hd', 'aria-expanded': 'true' }, [ttxt('Overview')]),
    ce('div', { className: 'card-bd' }, [
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('ID')]), ce('span', { className: 'val mono' }, [txt(n.id)])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Name')]), ce('span', { className: 'val' }, [txt(n.name)])]),
      ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Type')]), ce('span', { className: 'val' }, [txt(n.node_type)])]),
      n.parent_id ? ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Primary Parent')]), ce('span', { className: 'val' }, [txt(getNodeNameById(n.parent_id) + ' (' + n.parent_id.slice(0,8) + ')')])]) : null,
      locationTargets.length > 0 ? ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Current Location')]), ce('span', { className: 'val' }, [txt(locationTargets.join(', '))])]) : null,
      organizationTargets.length > 0 ? ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('Organization Links')]), ce('span', { className: 'val' }, [txt(organizationTargets.join(', '))])]) : null,
      externalParents.length > 0 ? ce('div', { className: 'prop-row' }, [ce('span', { className: 'key' }, [ttxt('External Parents')]), ce('span', { className: 'val' }, [txt(externalParents.join(', '))])]) : null,
      ce('div', { className: 'hint', style: { textAlign: 'left', marginTop: '8px' } }, [ttxt('Primary Parent is the only hierarchy field shown in the outline. Use located_at for the current environment, belongs_to/subordinate for organization or control, and external_parent only for auxiliary DAG scope.')]),
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
    ce('div', { className: 'card-bd', id: 'relDetailList' }, [
      ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '8px' } }, [ttxt('Relations do not change the outline hierarchy unless you separately edit Primary Parent.')]),
      nd.relation_validation_issues && nd.relation_validation_issues.length > 0 ? ce('div', { className: 'status-box', style: { marginBottom: '8px', textAlign: 'left' } }, [
        ce('div', { style: { fontWeight: 600, marginBottom: '6px' } }, [ttxt('Relation Validation')]),
        ce('div', { className: 'hint' }, [txt(nd.relation_validation_issues.map(function(issue) { return '[' + issue.severity + '] ' + issue.code + ': ' + issue.message; }).join('\n'))]),
      ]) : null,
      nd.graph_context_preview && nd.graph_context_preview.summary && nd.graph_context_preview.summary.length > 0 ? ce('div', { className: 'status-box', style: { marginBottom: '8px', textAlign: 'left' } }, [
        ce('div', { style: { fontWeight: 600, marginBottom: '6px' } }, [ttxt('Graph Context Preview')]),
        ce('div', { className: 'hint' }, [txt(nd.graph_context_preview.summary.join('\n'))]),
      ]) : null,
      buildRelationBlueprintCard(n, nd.relations || []),
    ]),
  ]);
  detail.appendChild(relSection);
  
  container.appendChild(detail);
  
  // Populate component list
  const compList = document.getElementById('compDetailList');
  if (compList && nd.components) {
    for (const c of nd.components) {
      const row = ce('div', { className: 'detail-item', dataset: { id: c.id } }, [
        ce('div', { className: 'item-hd' }, [
          ce('span', { className: 'item-tag', style: {background: 'var(--accent-soft)', color: 'var(--accent)'} }, [txt(componentTypeDisplay(c.component_type))]),
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
          ce('button', { className: 'item-propagate', dataset: { id: m.id } }, [ttxt('Propagate')]),
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
        ce('div', { className: 'item-body' }, [
          ce('div', { className: 'hint', style: { textAlign: 'left', marginBottom: '6px' } }, [txt((srcName || r.source_id) + ' -> ' + (tgtName || r.target_id) + ' | ' + relationTypeDescription(r.relation_type))]),
          function(){if(r.properties&&isJSON(r.properties)){var p=tryParseJSON(r.properties);return p?renderObjectKV(p,false):txt(r.properties)}return txt('Weight: '+r.weight)}(),
        ]),
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
  memList.querySelectorAll('.item-propagate').forEach(function(btn) { btn.addEventListener('click', function(e) { e.stopPropagation(); openPropagateMemoryModal(btn.dataset.id); }); });
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
function renderWorldTimeSettingsEditor(initialSettings) {
  var existing = initialSettings || {};
  var initialTickUnits = Array.isArray(existing.tick_units) && existing.tick_units.length ? existing.tick_units.slice() : [''];
  var carryByFrom = {};
  var sequenceByUnit = {};
  var calendarValuesByUnit = {};
  (existing.time_scale_carry || []).forEach(function(rule) {
    if (!rule || !rule.from) return;
    carryByFrom[String(rule.from)] = Number.isInteger(rule.base) ? rule.base : '';
  });
  (existing.unit_value_sequences || []).forEach(function(seq) {
    if (!seq || !seq.unit) return;
    sequenceByUnit[String(seq.unit)] = Array.isArray(seq.values) ? seq.values.join(' | ') : '';
  });
  if (existing.time_calendar && Array.isArray(existing.time_calendar.units)) {
    existing.time_calendar.units.forEach(function(unit) {
      if (!unit || !unit.unit) return;
      calendarValuesByUnit[String(unit.unit)] = unit.value || '';
    });
  }

  var root = ce('div', { className: 'world-time-editor' }, []);
  var unitsStatus = ce('div', { id: 'worldTimeUnitsStatus', className: 'hint', style: { textAlign: 'left', padding: '0 0 8px 0' } }, [ttxt('Tick units must be ordered from largest to smallest. The smallest unit becomes Tick Min Unit automatically.')]);
  root.appendChild(unitsStatus);

  var modeRow = ce('div', { className: 'setting-row setting-row-top' }, [
    ce('label', { for: 'setTickScaleMode' }, [ttxt('Tick Scale Mode')]),
    el('select', { id: 'setTickScaleMode', innerHTML: '<option value="fixed">fixed</option><option value="flexible">flexible</option>' }),
  ]);
  root.appendChild(modeRow);

  var stepRow = ce('div', { className: 'setting-row' }, [
    ce('label', { for: 'setTickStep' }, [ttxt('Tick Step')]),
    el('input', { id: 'setTickStep', type: 'number', min: '1', max: '999999', value: String(existing.tick_step || 1) }),
    ce('div', { className: 'hint setting-inline-hint' }, [ttxt('The engine advances this many minimum tick units per inferred tick.')]),
  ]);
  root.appendChild(stepRow);

  var unitsRow = ce('div', { className: 'setting-row setting-row-block' }, [
    ce('label', {}, [ttxt('Tick Units')]),
  ]);
  var unitsBody = ce('div', { className: 'world-time-block' }, []);
  var unitsList = ce('div', { id: 'worldTimeTickUnitsList', className: 'world-time-list' }, []);
  unitsBody.appendChild(unitsList);
  unitsBody.appendChild(ce('div', { className: 'policy-actions world-time-actions' }, [
    ce('button', { type: 'button', id: 'btnAddTickUnit' }, [ttxt('Add Unit')]),
  ]));
  unitsRow.appendChild(unitsBody);
  root.appendChild(unitsRow);

  var minUnitRow = ce('div', { className: 'setting-row' }, [
    ce('label', { for: 'setTickMinUnit' }, [ttxt('Tick Min Unit')]),
    el('select', { id: 'setTickMinUnit' }),
    ce('div', { className: 'hint setting-inline-hint' }, [ttxt('Derived from the smallest configured tick unit to match Engine constraints.')]),
  ]);
  root.appendChild(minUnitRow);

  var carryRow = ce('div', { className: 'setting-row setting-row-block' }, [
    ce('label', {}, [ttxt('Time Scale Carry')]),
  ]);
  var carryBody = ce('div', { id: 'worldTimeCarryList', className: 'world-time-block' }, []);
  carryRow.appendChild(carryBody);
  root.appendChild(carryRow);

  var calendarToggleRow = ce('div', { className: 'setting-row' }, [
    ce('label', { className: 'checkbox-row' }, [
      el('input', { id: 'setCalendarEnabled', type: 'checkbox', checked: !!(existing.time_calendar && existing.time_calendar.enabled) }),
      ttxt('Enable Calendar Mode'),
    ]),
    ce('div', { className: 'hint setting-inline-hint' }, [ttxt('When enabled, calendar units, carry rules, and sequences must stay consistent.')]),
  ]);
  root.appendChild(calendarToggleRow);

  var calendarSection = ce('div', { id: 'worldTimeCalendarSection', className: 'world-time-calendar-section' }, []);
  calendarSection.appendChild(ce('div', { className: 'setting-row' }, [
    ce('label', { for: 'setCalendarName' }, [ttxt('Calendar Name')]),
    el('input', { id: 'setCalendarName', value: existing.time_calendar && existing.time_calendar.calendar_name ? existing.time_calendar.calendar_name : '', placeholder: tr('Example: 太阴'), style: { width: '100%' } }),
  ]));
  var calendarUnitsRow = ce('div', { className: 'setting-row setting-row-block' }, [
    ce('label', {}, [ttxt('Calendar Units')]),
  ]);
  var calendarUnitsBody = ce('div', { id: 'worldTimeCalendarUnitsList', className: 'world-time-block' }, []);
  calendarUnitsRow.appendChild(calendarUnitsBody);
  calendarSection.appendChild(calendarUnitsRow);
  root.appendChild(calendarSection);

  var sequenceRow = ce('div', { className: 'setting-row setting-row-block' }, [
    ce('label', {}, [ttxt('Unit Value Sequences')]),
  ]);
  var sequenceBody = ce('div', { id: 'worldTimeSequenceList', className: 'world-time-block' }, []);
  sequenceRow.appendChild(sequenceBody);
  root.appendChild(sequenceRow);

  function arrayFrom(nodeList) {
    return Array.prototype.slice.call(nodeList || []);
  }

  function normalizeUnits(values) {
    var result = [];
    var seen = {};
    (values || []).forEach(function(value) {
      var trimmed = String(value || '').trim();
      if (!trimmed || seen[trimmed]) return;
      seen[trimmed] = true;
      result.push(trimmed);
    });
    return result;
  }

  function collectDraft() {
    var tickUnits = arrayFrom(root.querySelectorAll('.world-time-unit-input')).map(function(input) { return input.value; });
    var nextCarryByFrom = {};
    arrayFrom(root.querySelectorAll('.world-time-carry-base')).forEach(function(input) {
      nextCarryByFrom[String(input.dataset.from || '')] = input.value.trim();
    });
    var nextSequenceByUnit = {};
    arrayFrom(root.querySelectorAll('.world-time-sequence-values')).forEach(function(input) {
      nextSequenceByUnit[String(input.dataset.unit || '')] = input.value;
    });
    var nextCalendarValuesByUnit = {};
    arrayFrom(root.querySelectorAll('.world-time-calendar-value')).forEach(function(input) {
      nextCalendarValuesByUnit[String(input.dataset.unit || '')] = input.value;
    });
    return {
      tick_units: tickUnits,
      carry_by_from: nextCarryByFrom,
      sequence_by_unit: nextSequenceByUnit,
      calendar_values_by_unit: nextCalendarValuesByUnit,
      calendar_enabled: !!document.getElementById('setCalendarEnabled').checked,
    };
  }

  function renderTickUnits(values) {
    unitsList.innerHTML = '';
    values.forEach(function(value, index) {
      var row = ce('div', { className: 'world-time-list-row' }, []);
      row.appendChild(ce('span', { className: 'world-time-order' }, [txt(String(index + 1))]));
      row.appendChild(el('input', {
        type: 'text',
        value: value,
        className: 'world-time-unit-input',
        placeholder: tr(index === 0 ? 'Example: 年' : 'Example: 时辰'),
        style: { width: '100%' },
      }));
      var actions = ce('div', { className: 'world-time-row-actions' }, [
        ce('button', { type: 'button', className: 'world-time-move-up', disabled: index === 0 }, [ttxt('Up')]),
        ce('button', { type: 'button', className: 'world-time-move-down', disabled: index === values.length - 1 }, [ttxt('Down')]),
        ce('button', { type: 'button', className: 'danger world-time-delete-unit', disabled: values.length === 1 }, [ttxt('Delete')]),
      ]);
      row.appendChild(actions);
      unitsList.appendChild(row);

      row.querySelector('.world-time-unit-input').addEventListener('input', syncLinkedSections);
      row.querySelector('.world-time-move-up').addEventListener('click', function() {
        var draft = collectDraft();
        var tmp = draft.tick_units[index - 1];
        draft.tick_units[index - 1] = draft.tick_units[index];
        draft.tick_units[index] = tmp;
        renderTickUnits(draft.tick_units);
        syncLinkedSections();
      });
      row.querySelector('.world-time-move-down').addEventListener('click', function() {
        var draft = collectDraft();
        var tmp = draft.tick_units[index + 1];
        draft.tick_units[index + 1] = draft.tick_units[index];
        draft.tick_units[index] = tmp;
        renderTickUnits(draft.tick_units);
        syncLinkedSections();
      });
      row.querySelector('.world-time-delete-unit').addEventListener('click', function() {
        var draft = collectDraft();
        draft.tick_units.splice(index, 1);
        if (!draft.tick_units.length) draft.tick_units.push('');
        renderTickUnits(draft.tick_units);
        syncLinkedSections();
      });
    });
  }

  function syncLinkedSections() {
    var draft = collectDraft();
    var normalizedUnits = normalizeUnits(draft.tick_units);
    var hasBlank = draft.tick_units.some(function(value) { return !String(value || '').trim(); });
    var hasDuplicate = normalizedUnits.length !== draft.tick_units.filter(function(value) { return String(value || '').trim(); }).length;

    unitsStatus.textContent = tr('Tick units must be ordered from largest to smallest. The smallest unit becomes Tick Min Unit automatically.');
    if (hasBlank) unitsStatus.textContent = tr('Tick units contain blank values. Fill them before saving.');
    if (!hasBlank && hasDuplicate) unitsStatus.textContent = tr('Tick units contain duplicates. Keep each unit unique before saving.');

    var minUnitSelect = document.getElementById('setTickMinUnit');
    minUnitSelect.innerHTML = '';
    if (normalizedUnits.length === 0) {
      minUnitSelect.appendChild(el('option', { value: '', textContent: '' }));
      minUnitSelect.value = '';
    } else {
      normalizedUnits.forEach(function(unit) {
        minUnitSelect.appendChild(el('option', { value: unit, textContent: unit }));
      });
      minUnitSelect.value = normalizedUnits[normalizedUnits.length - 1];
    }

    carryBody.innerHTML = '';
    if (normalizedUnits.length <= 1) {
      carryBody.appendChild(ce('div', { className: 'hint', style: { textAlign: 'left', padding: '0' } }, [ttxt('Add at least two tick units to configure carry rules.')]))
    } else {
      for (var i = normalizedUnits.length - 1; i > 0; i--) {
        var fromUnit = normalizedUnits[i];
        var toUnit = normalizedUnits[i - 1];
        var carryRow = ce('div', { className: 'world-time-linked-row' }, [
          ce('span', { className: 'world-time-link-label' }, [txt(fromUnit + ' -> ' + toUnit)]),
          el('input', {
            type: 'number',
            min: '1',
            step: '1',
            value: draft.carry_by_from[fromUnit] || carryByFrom[fromUnit] || '',
            className: 'world-time-carry-base',
            dataset: { from: fromUnit, to: toUnit },
          }),
          ce('span', { className: 'hint' }, [ttxt('Base')]),
        ]);
        carryBody.appendChild(carryRow);
        carryRow.querySelector('.world-time-carry-base').addEventListener('input', syncLinkedSections);
      }
    }

    sequenceBody.innerHTML = '';
    if (normalizedUnits.length <= 1) {
      sequenceBody.appendChild(ce('div', { className: 'hint', style: { textAlign: 'left', padding: '0' } }, [ttxt('Optional symbolic sequences become available after you add smaller tick units.')]))
    } else {
      for (var j = 1; j < normalizedUnits.length; j++) {
        var unit = normalizedUnits[j];
        var seqBase = draft.carry_by_from[unit] || carryByFrom[unit] || '';
        var seqRow = ce('div', { className: 'world-time-sequence-row' }, [
          ce('div', { className: 'world-time-sequence-header' }, [
            ce('span', { className: 'world-time-link-label' }, [txt(unit)]),
            ce('span', { className: 'hint' }, [txt(tr('Expected count') + ': ' + (seqBase || '-'))]),
          ]),
          el('input', {
            type: 'text',
            value: draft.sequence_by_unit[unit] || sequenceByUnit[unit] || '',
            className: 'world-time-sequence-values',
            dataset: { unit: unit },
            placeholder: tr('Example: 子 | 丑 | 寅 | 卯'),
            style: { width: '100%' },
          }),
        ]);
        sequenceBody.appendChild(seqRow);
      }
      arrayFrom(sequenceBody.querySelectorAll('.world-time-sequence-values')).forEach(function(input) {
        input.addEventListener('input', syncLinkedSections);
      });
    }

    calendarSection.classList.toggle('is-collapsed', !draft.calendar_enabled);
    calendarUnitsBody.innerHTML = '';
    if (draft.calendar_enabled) {
      if (normalizedUnits.length === 0) {
        calendarUnitsBody.appendChild(ce('div', { className: 'hint', style: { textAlign: 'left', padding: '0' } }, [ttxt('Add tick units before configuring calendar values.')]))
      } else {
        normalizedUnits.forEach(function(unit) {
          var row = ce('div', { className: 'world-time-linked-row' }, [
            ce('span', { className: 'world-time-link-label' }, [txt(unit)]),
            el('input', {
              type: 'text',
              value: draft.calendar_values_by_unit[unit] || calendarValuesByUnit[unit] || '',
              className: 'world-time-calendar-value',
              dataset: { unit: unit },
              placeholder: tr('Initial value'),
              style: { width: '100%' },
            }),
          ]);
          calendarUnitsBody.appendChild(row);
        });
      }
    }
  }

  renderTickUnits(initialTickUnits);
  document.getElementById('setTickScaleMode').value = existing.tick_scale_mode || 'fixed';
  document.getElementById('btnAddTickUnit').addEventListener('click', function() {
    var draft = collectDraft();
    draft.tick_units.push('');
    renderTickUnits(draft.tick_units);
    syncLinkedSections();
  });
  document.getElementById('setCalendarEnabled').addEventListener('change', syncLinkedSections);
  syncLinkedSections();
  return root;
}

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
  const wt = s.world_time_settings || {};
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
      var minValue = type === 'number_zero' ? '0' : '1';
      row.appendChild(el('input', { type: 'number', id: id, value: String(val), min: minValue, max: '999' }));
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
  settingRow('Propagation Max Depth', 'setPropMaxDepth', 'number_zero', s.propagation_max_depth != null ? s.propagation_max_depth : 2);
  settingRow('Enable Propagation Machine', 'setEnablePropMachine', 'bool', s.enable_propagation_machine);
  form.appendChild(ce('div', { className: 'detail-card' }, [ce('div', { className: 'card-hd' }, [ttxt('Sub-Task DAG')])]));
  settingRow('Sub-Task Max Retries', 'setSubTaskRetries', 'number_zero', s.sub_task_max_retries != null ? s.sub_task_max_retries : 2);
  settingRow('Sub-Task Timeout (sec)', 'setSubTaskTimeout', 'number_zero', s.sub_task_timeout_secs != null ? s.sub_task_timeout_secs : 60);
  form.appendChild(ce('div', { className: 'detail-card' }, [
    ce('div', { className: 'card-hd' }, [ttxt('World Time Settings')]),
    ce('div', { className: 'card-bd' }, [
      renderWorldTimeSettingsEditor(wt),
    ]),
  ]));
  const btnRow = ce('div', { className: 'policy-actions' }, [
    ce('button', { className: 'primary', id: 'btnSaveSettings' }, [ttxt('Save Settings')]),
  ]);
  form.appendChild(btnRow);
  container.appendChild(form);
  document.getElementById('btnSaveSettings').addEventListener('click', saveSettings);
}

/* ============= Logs Page ============= */
function parseLogJSON(raw) {
  if (!raw || !raw.trim || !raw.trim()) return null;
  try { return JSON.parse(raw); } catch (e) { return null; }
}

function shortID(value) {
  value = value || '';
  return value.length <= 8 ? value : value.slice(0, 8);
}

function formatInferenceLogTime(value) {
  if (!value) return '-';
  try {
    return new Date(value).toLocaleString();
  } catch (e) {
    return value;
  }
}

function getContinuityComponentData(bundle, componentType) {
  var items = bundle && bundle.state_components ? bundle.state_components : [];
  for (var i = 0; i < items.length; i++) {
    if (items[i] && items[i].component_type === componentType) return items[i].data || null;
  }
  return null;
}

function normalizeStringList(values) {
  if (!Array.isArray(values)) return [];
  var seen = {};
  var result = [];
  values.forEach(function(value) {
    if (typeof value !== 'string') return;
    var trimmed = value.trim();
    if (!trimmed || seen[trimmed]) return;
    seen[trimmed] = true;
    result.push(trimmed);
  });
  return result;
}

function diffStringLists(currentList, previousList) {
  var current = normalizeStringList(currentList);
  var previous = normalizeStringList(previousList);
  var previousSet = {};
  var currentSet = {};
  previous.forEach(function(item) { previousSet[item] = true; });
  current.forEach(function(item) { currentSet[item] = true; });
  return {
    added: current.filter(function(item) { return !previousSet[item]; }),
    removed: previous.filter(function(item) { return !currentSet[item]; }),
    stable: current.filter(function(item) { return !!previousSet[item]; }),
  };
}

function joinListPreview(values) {
  var items = normalizeStringList(values);
  return items.length > 0 ? items.join(' | ') : '-';
}

function renderPropRow(label, value, opts) {
  if (value === null || value === undefined || value === '') return null;
  var options = opts || {};
  var valueNode = value;
  if (!(valueNode instanceof Node)) {
    valueNode = ce('span', { className: 'val' + (options.mono ? ' mono' : '') }, [txt(String(value))]);
  }
  return ce('div', { className: 'prop-row' + (options.className ? ' ' + options.className : '') }, [
    ce('span', { className: 'key' }, [options.rawLabel ? txt(label) : ttxt(label)]),
    valueNode,
  ]);
}

function createToggleDetailCard(headerChildren, expanded) {
  var isExpanded = !!expanded;
  var card = ce('div', { className: 'trace-card detail-card' }, []);
  var hd = ce('div', { className: 'card-hd toggle-hd', 'aria-expanded': isExpanded ? 'true' : 'false' }, headerChildren);
  var body = ce('div', { className: 'card-bd', style: { display: isExpanded ? 'block' : 'none' } }, []);
  hd.addEventListener('click', function() {
    var expandedState = this.getAttribute('aria-expanded') === 'true';
    this.setAttribute('aria-expanded', expandedState ? 'false' : 'true');
    body.style.display = expandedState ? 'none' : 'block';
  });
  card.appendChild(hd);
  card.appendChild(body);
  return { card: card, body: body, header: hd };
}

function renderPreviewBlock(label, text, maxLength) {
  if (!text) return null;
  var preview = text;
  if (maxLength && text.length > maxLength) {
    preview = text.slice(0, maxLength) + '...';
  }
  return renderPropRow(label, ce('pre', { className: 'trace-pre' }, [txt(preview)]), { className: 'trace-block' });
}

function renderPreviewListBlock(label, items, formatter) {
  if (!items || items.length === 0) return null;
  var block = ce('div', { className: 'prop-row trace-block' }, [
    ce('span', { className: 'key' }, [txt(tr(label) + ' (' + items.length + ')')]),
  ]);
  for (var i = 0; i < items.length; i++) {
    block.appendChild(ce('div', { style: { paddingLeft: '12px', fontSize: '12px', color: 'var(--fg2)' } }, [txt(formatter(items[i]))]));
  }
  return block;
}

function renderLogDetailBlock(label, value) {
  if (value === null || value === undefined || value === '') return null;
  var content = null;
  if (typeof value === 'string') {
    content = isJSON(value) ? renderKV(value, false) : ce('pre', { className: 'trace-pre' }, [txt(value)]);
  } else if (typeof value === 'object') {
    content = renderObjectKV(value, false);
  } else {
    content = ce('span', { className: 'val' }, [txt(String(value))]);
  }
  return renderPropRow(label, content, { className: 'trace-block' });
}

function renderLogsPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnRefreshLogs' }, [ttxt('Refresh Logs')]),
  ]);
  container.appendChild(toolbar);
  const list = ce('div', { className: 'trace-container' }, []);
  if (state.logs.length === 0) {
    list.appendChild(ce('div', { className: 'hint' }, [ttxt('No logs yet.')]));
  }
  for (const log of state.logs) {
    var t = log.created_at;
    try { t = new Date(t).toLocaleString(); } catch(e) {}
    var requestData = parseLogJSON(log.request_data || '');
    var responseData = parseLogJSON(log.response_data || '');
    var pipeline = '-';
    if (responseData && (responseData.configured_pipeline_mode || responseData.effective_pipeline_mode)) {
      pipeline = (responseData.configured_pipeline_mode || '-') + ' -> ' + (responseData.effective_pipeline_mode || '-');
    } else if (requestData && requestData.pipeline_mode) {
      pipeline = requestData.pipeline_mode;
    }
    var rounds = '-';
    if (responseData && (responseData.rounds_used || responseData.max_analysis_rounds)) {
      rounds = String(responseData.rounds_used || 0) + ' / ' + String(responseData.max_analysis_rounds || 0);
    } else if (requestData && requestData.max_analysis_rounds) {
      rounds = '0 / ' + String(requestData.max_analysis_rounds);
    }

    var detailCard = createToggleDetailCard([
      ce('span', { style: { fontWeight: 600 } }, [txt(log.task_type || '-')]),
      txt(' ' + (log.duration_ms || 0) + 'ms'),
      txt(' ' + (log.tokens_used || 0) + ' tokens'),
    ], false);

    var body = detailCard.body;
    body.appendChild(renderPropRow('Time', t || '-'));
    body.appendChild(renderPropRow('World', log.world_id ? log.world_id.slice(0, 8) : '-', { mono: true }));
    body.appendChild(renderPropRow('Node', log.node_id ? log.node_id.slice(0, 8) : '-', { mono: true }));
    body.appendChild(renderPropRow('Model', log.llm_model || '-'));
    body.appendChild(renderPropRow('Tokens', String(log.tokens_used || 0)));
    body.appendChild(renderPropRow('Duration', String(log.duration_ms || 0) + 'ms'));
    body.appendChild(renderPropRow('Pipeline', pipeline));
    body.appendChild(renderPropRow('Rounds', rounds));
    if (log.category || log.event_name || log.execution_mode) {
      body.appendChild(renderPropRow('State Component', [log.category || '-', log.event_name || '-', log.execution_mode || '-'].join(' / ')));
    }
    if (log.message) {
      body.appendChild(renderPropRow('Summary', log.message));
    }

    if (requestData) {
      body.appendChild(renderLogDetailBlock('Request', requestData));
    }
    if (responseData) {
      body.appendChild(renderLogDetailBlock('Response', responseData));
    }
    if (log.detail_data) {
      body.appendChild(renderLogDetailBlock('Detail', parseLogJSON(log.detail_data || '') || log.detail_data));
    }

    list.appendChild(detailCard.card);
  }
  container.appendChild(list);
  document.getElementById('btnRefreshLogs').addEventListener('click', loadLogs);
}

function renderStatePage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnRefreshState' }, [ttxt('Refresh State')]),
  ]);
  container.appendChild(toolbar);
  const list = ce('div', { className: 'trace-container' }, []);
  if (!state.stateComponents || state.stateComponents.length === 0) {
    list.appendChild(ce('div', { className: 'hint' }, [ttxt('No state components yet.')]));
  }
  for (var i = 0; i < (state.stateComponents || []).length; i++) {
    var item = state.stateComponents[i];
    var editable = item.component_type !== 'state_snapshot';
    var isWorldTimeState = item.component_type === 'world_time_state';
    var detailCard = createToggleDetailCard([
      ce('span', { style: { fontWeight: 600 } }, [txt(componentTypeDisplay(item.component_type || '?'))]),
      txt(' ' + tr(item.component ? 'present' : 'missing')),
    ], false);
    var body = detailCard.body;
    body.appendChild(renderPropRow('Type', componentTypeDisplay(item.component_type || '-')));
    body.appendChild(renderPropRow('ID', item.component && item.component.id ? item.component.id.slice(0, 8) : '-', { mono: true }));
    body.appendChild(renderPropRow('Node', item.component && item.component.node_id ? item.component.node_id.slice(0, 8) : '-', { mono: true }));
    body.appendChild(renderPropRow('Summary', tr('Structured world tick continuity state.')));
    if (isWorldTimeState && item.data) {
      body.appendChild(renderPropRow('World Time Label', item.data.current_time_label || '-'));
      body.appendChild(renderPropRow('Advanced Ticks', String(item.data.last_advanced_ticks || 0)));
      body.appendChild(renderPropRow('Tick Scale Mode', item.data.tick_scale_mode || '-'));
    }
    if (editable) {
      var actionRow = ce('div', { className: 'world-toolbar', style: { padding: '6px 0 2px 0', borderBottom: 'none' } }, [
        ce('button', { id: 'btnEditState_' + i }, [ttxt('Edit')]),
      ]);
      body.appendChild(actionRow);
      (function(componentType, payload) {
        actionRow.querySelector('button').addEventListener('click', function() {
          openEditStateComponentModal(componentType, payload);
        });
      })(item.component_type, item.data || {});
    }
    if (item.data) {
      body.appendChild(renderLogDetailBlock('Data', item.data));
    }
    list.appendChild(detailCard.card);
  }
  container.appendChild(list);
  document.getElementById('btnRefreshState').addEventListener('click', loadStateComponents);
}

function renderContinuityPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar continuity-toolbar' }, [
    ce('button', { id: 'btnRefreshContinuity' }, [ttxt('Refresh Continuity')]),
    ce('select', { id: 'continuityRequestFilter' }, []),
    ce('select', { id: 'continuityModeFilter' }, []),
    ce('button', { id: 'btnClearContinuityFilter' }, [ttxt('Clear Filter')]),
  ]);
  container.appendChild(toolbar);

  if (!state.selectedWorldId) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('Select a world first.')]));
    return;
  }

  var bundle = state.continuityBundle;
  var requestSelect = document.getElementById('continuityRequestFilter');
  requestSelect.appendChild(el('option', { value: '', textContent: tr('All Requests') }));
  var modeSelect = document.getElementById('continuityModeFilter');
  modeSelect.appendChild(el('option', { value: '', textContent: tr('All Modes') }));

  if (!bundle) {
    container.appendChild(ce('div', { className: 'hint' }, [ttxt('No continuity artifacts yet.')]));
  } else {
    var requestOptions = getContinuityRequestIds(bundle);
    requestOptions.forEach(function(requestId) {
      requestSelect.appendChild(el('option', { value: requestId, textContent: shortID(requestId) }));
    });
    requestSelect.value = state.continuityRequestId || '';

    var modeOptions = getContinuityModes(bundle);
    modeOptions.forEach(function(mode) {
      modeSelect.appendChild(el('option', { value: mode, textContent: mode }));
    });
    modeSelect.value = state.continuityMode || '';

    var logs = getFilteredContinuityLogs(bundle);
    var traces = getFilteredContinuityTraces(bundle);
    var latest = bundle.latest_timeline || ((bundle.timelines || []).length > 0 ? bundle.timelines[0] : null);
    var previousTimeline = bundle.timelines && bundle.timelines.length > 1 ? bundle.timelines[1] : null;
    var worldState = getContinuityComponentData(bundle, 'world_state') || {};
    var worldTimeState = getContinuityComponentData(bundle, 'world_time_state') || {};
    var storyHistory = getContinuityComponentData(bundle, 'story_history') || {};
    var historyEntries = Array.isArray(storyHistory.entries) ? storyHistory.entries : [];
    var latestHistory = historyEntries.length > 0 ? historyEntries[0] : null;
    var previousHistory = historyEntries.length > 1 ? historyEntries[1] : null;
    var factDiff = diffStringLists(worldState.canonical_facts || [], previousHistory ? previousHistory.facts || [] : []);

    var summaryGrid = ce('div', { className: 'continuity-grid' }, []);
    summaryGrid.appendChild(ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Continuity Summary')]),
      ce('div', { className: 'card-bd' }, [
        renderPropRow('World', state.selectedWorldId, { mono: true }),
        renderPropRow('Latest Tick', latest ? String(latest.tick_number || 0) : '-'),
        renderPropRow('Execution Mode', state.continuityMode || tr('All Modes')),
        renderPropRow('Tracked Request', state.continuityRequestId ? shortID(state.continuityRequestId) : tr('No request filter applied.')),
        renderPropRow('Linked Logs', String(logs.length)),
        renderPropRow('Linked Traces', String(traces.length)),
        renderPropRow('World Time Label', worldTimeState.current_time_label || '-'),
        renderPropRow('Advanced Ticks', latest ? String(latest.advanced_ticks || 0) : '-'),
      ]),
    ]));

    var stateCard = ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Continuity State')]),
      ce('div', { className: 'card-bd' }, []),
    ]);
    var stateCardBody = stateCard.querySelector('.card-bd');
    if (!bundle.state_components || bundle.state_components.length === 0) {
      stateCardBody.appendChild(ce('div', { className: 'hint' }, [ttxt('No state components yet.')]));
    } else {
      bundle.state_components.forEach(function(item) {
        var status = item.component ? 'present' : 'missing';
        stateCardBody.appendChild(renderPropRow(componentTypeDisplay(item.component_type || '-'), status, { rawLabel: true }));
      });
      stateCardBody.appendChild(ce('div', { className: 'hint', style: { textAlign: 'left', marginTop: '8px' } }, [ttxt('Select a request to focus linked logs and traces.')]));
    }
    summaryGrid.appendChild(stateCard);
    container.appendChild(summaryGrid);

    var diffCard = ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Continuity Diff')]),
      ce('div', { className: 'card-bd' }, []),
    ]);
    var diffBody = diffCard.querySelector('.card-bd');
    if (!previousTimeline && !previousHistory) {
      diffBody.appendChild(ce('div', { className: 'hint' }, [ttxt('No previous tick to compare.')]));
    } else {
      diffBody.appendChild(renderPropRow('Latest Tick Summary', latest ? latest.summary || '-' : '-'));
      diffBody.appendChild(renderPropRow('Latest Advanced Ticks', latest ? String(latest.advanced_ticks || 0) : '-'));
      diffBody.appendChild(renderPropRow('Previous Advanced Ticks', previousTimeline ? String(previousTimeline.advanced_ticks || 0) : '-'));
      diffBody.appendChild(renderPropRow('Previous Tick Summary', previousTimeline ? previousTimeline.summary || '-' : '-'));
      diffBody.appendChild(renderPropRow('Latest Future Outline', latest ? latest.future_outline || '-' : '-'));
      diffBody.appendChild(renderPropRow('Previous Future Outline', previousTimeline ? previousTimeline.future_outline || '-' : '-'));
      diffBody.appendChild(renderPropRow('Latest History Summary', latestHistory ? latestHistory.summary || '-' : '-'));
      diffBody.appendChild(renderPropRow('Previous History Summary', previousHistory ? previousHistory.summary || '-' : '-'));
      diffBody.appendChild(renderPropRow('Current Canonical Facts', joinListPreview(worldState.canonical_facts || [])));
      diffBody.appendChild(renderPropRow('Previous Story Facts', joinListPreview(previousHistory ? previousHistory.facts || [] : [])));
      diffBody.appendChild(renderPropRow('Added Facts', joinListPreview(factDiff.added)));
      diffBody.appendChild(renderPropRow('Removed Facts', joinListPreview(factDiff.removed)));
      diffBody.appendChild(renderPropRow('Stable Facts', joinListPreview(factDiff.stable)));
    }
    container.appendChild(diffCard);

    var timelinesCard = ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Recent World Tick Bundle')]),
      ce('div', { className: 'card-bd' }, []),
    ]);
    var timelinesBody = timelinesCard.querySelector('.card-bd');
    if (!bundle.timelines || bundle.timelines.length === 0) {
      timelinesBody.appendChild(ce('div', { className: 'hint' }, [ttxt('No timelines yet.')]));
    } else {
      bundle.timelines.forEach(function(item) {
        var row = createToggleDetailCard([
          ce('span', { style: { fontWeight: 600 } }, [txt('#' + String(item.tick_number || 0))]),
          txt(' ' + (item.tick_type || '-')),
          txt(item.game_time ? ' ' + item.game_time : ''),
        ], false);
        row.body.appendChild(renderPropRow('Advanced Ticks', String(item.advanced_ticks || 0)));
        row.body.appendChild(renderPropRow('Summary', item.summary || '-'));
        row.body.appendChild(renderPropRow('Future Outline', item.future_outline || '-'));
        if (item.data && item.data.world_time_state) {
          row.body.appendChild(renderPropRow('World Time Label', item.data.world_time_state.current_time_label || '-'));
        }
        if (item.data) row.body.appendChild(renderLogDetailBlock('Timeline Payload', item.data));
        timelinesBody.appendChild(row.card);
      });
    }
    container.appendChild(timelinesCard);

    var logsCard = ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Recent World Tick Logs')]),
      ce('div', { className: 'card-bd' }, []),
    ]);
    var logsBody = logsCard.querySelector('.card-bd');
    if (logs.length === 0) {
      logsBody.appendChild(ce('div', { className: 'hint' }, [ttxt('No logs yet.')]));
    } else {
      logs.slice(0, 12).forEach(function(log, index) {
        var card = createToggleDetailCard([
          ce('span', { style: { fontWeight: 600 } }, [txt(log.task_type || '-')]),
          txt(' ' + (log.duration_ms || 0) + 'ms'),
          txt(' ' + (log.execution_mode || '-')),
        ], index === 0);
        card.body.appendChild(renderPropRow('Time', formatInferenceLogTime(log.created_at || '')));
        card.body.appendChild(renderPropRow('Request ID', log.request_id ? shortID(log.request_id) : '-', { mono: true }));
        card.body.appendChild(renderPropRow('Node', log.node_id ? shortID(log.node_id) : '-', { mono: true }));
        if (log.message) card.body.appendChild(renderPropRow('Summary', log.message));
        var requestData = parseLogJSON(log.request_data || '');
        var responseData = parseLogJSON(log.response_data || '');
        if (requestData) card.body.appendChild(renderLogDetailBlock('Request', requestData));
        if (responseData) card.body.appendChild(renderLogDetailBlock('Response', responseData));
        if (log.detail_data) card.body.appendChild(renderLogDetailBlock('Detail', parseLogJSON(log.detail_data || '') || log.detail_data));
        if (log.request_id) {
          card.body.appendChild(ce('div', { className: 'policy-actions continuity-actions' }, [
            ce('button', { className: 'continuity-request-btn', dataset: { requestId: log.request_id } }, [ttxt('Tracked Request')]),
          ]));
        }
        logsBody.appendChild(card.card);
      });
    }
    container.appendChild(logsCard);

    var tracesCard = ce('div', { className: 'detail-card' }, [
      ce('div', { className: 'card-hd' }, [ttxt('Recent Debug Traces')]),
      ce('div', { className: 'card-bd' }, []),
    ]);
    var tracesBody = tracesCard.querySelector('.card-bd');
    if (traces.length === 0) {
      tracesBody.appendChild(ce('div', { className: 'hint' }, [ttxt('No traces yet. Run a task in Debug mode to see traces.')]));
    } else {
      traces.slice(0, 10).forEach(function(trace, index) {
        var card = createToggleDetailCard([
          ce('span', { style: { fontWeight: 600 } }, [txt(trace.task_type || '-')]),
          txt(' ' + (trace.duration_ms || 0) + 'ms'),
          txt(' ' + shortID(trace.request_id || '')),
        ], index === 0);
        card.body.appendChild(renderPropRow('Request ID', trace.request_id ? shortID(trace.request_id) : '-', { mono: true }));
        card.body.appendChild(renderPropRow('Time', trace.timestamp ? new Date(trace.timestamp).toLocaleString() : '-'));
        card.body.appendChild(renderPropRow('Pipeline', (trace.configured_pipeline_mode || '-') + ' -> ' + (trace.effective_pipeline_mode || '-')));
        card.body.appendChild(renderPropRow('Rounds', String(trace.rounds_used || 0) + ' / ' + String(trace.max_analysis_rounds || 0)));
        if (trace.error) card.body.appendChild(renderPropRow('Error', trace.error));
        if (trace.system_prompt) card.body.appendChild(renderLogDetailBlock('System Prompt', trace.system_prompt));
        if (trace.raw_llm_response) card.body.appendChild(renderLogDetailBlock('LLM Response', trace.raw_llm_response));
        if (trace.request_id) {
          card.body.appendChild(ce('div', { className: 'policy-actions continuity-actions' }, [
            ce('button', { className: 'continuity-request-btn', dataset: { requestId: trace.request_id } }, [ttxt('Tracked Request')]),
          ]));
        }
        tracesBody.appendChild(card.card);
      });
    }
    container.appendChild(tracesCard);
  }

  document.getElementById('btnRefreshContinuity').addEventListener('click', async function() {
    await loadContinuityOverview();
    toast(tr('Continuity refreshed'), 'success');
  });
  document.getElementById('continuityRequestFilter').addEventListener('change', function() {
    state.continuityRequestId = this.value || '';
    renderCurrent();
  });
  document.getElementById('continuityModeFilter').addEventListener('change', function() {
    state.continuityMode = this.value || '';
    renderCurrent();
  });
  document.getElementById('btnClearContinuityFilter').addEventListener('click', function() {
    state.continuityRequestId = '';
    state.continuityMode = '';
    renderCurrent();
  });
  document.querySelectorAll('.continuity-request-btn').forEach(function(btn) {
    btn.addEventListener('click', function() {
      state.continuityRequestId = btn.dataset.requestId || '';
      renderCurrent();
    });
  });
}

function renderTimelinesPage(container) {
  const toolbar = ce('div', { className: 'world-toolbar' }, [
    ce('button', { id: 'btnRefreshTimelines' }, [ttxt('Refresh Timelines')]),
  ]);
  container.appendChild(toolbar);
  const list = ce('div', { className: 'trace-container' }, []);
  if (!state.timelines || state.timelines.length === 0) {
    list.appendChild(ce('div', { className: 'hint' }, [ttxt('No timelines yet.')]));
  }
  for (var i = 0; i < (state.timelines || []).length; i++) {
    var item = state.timelines[i];
    var detailCard = createToggleDetailCard([
      ce('span', { style: { fontWeight: 600 } }, [txt('#' + String(item.tick_number || 0))]),
      txt(' ' + (item.tick_type || '-')),
      txt(item.game_time ? ' ' + item.game_time : ''),
    ], false);
    var body = detailCard.body;
    body.appendChild(renderPropRow('Tick', String(item.tick_number || 0)));
    body.appendChild(renderPropRow('Type', item.tick_type || '-'));
    body.appendChild(renderPropRow('Time', item.game_time || '-'));
    body.appendChild(renderPropRow('Advanced Ticks', String(item.advanced_ticks || 0)));
    body.appendChild(renderPropRow('Summary', item.summary || '-'));
    body.appendChild(renderPropRow('Future Outline', item.future_outline || '-'));
    if (item.data && item.data.world_time_state) {
      body.appendChild(renderPropRow('World Time Label', item.data.world_time_state.current_time_label || '-'));
    }
    if (item.data) {
      body.appendChild(renderLogDetailBlock('Timeline Payload', item.data));
    }
    list.appendChild(detailCard.card);
  }
  container.appendChild(list);
  document.getElementById('btnRefreshTimelines').addEventListener('click', loadTimelines);
}


function updateActionButtons() {
  var btnTick = document.getElementById('btnTickAdvance');
  var btnAuto = document.getElementById('btnAutonomous');
  var btnAutoConfig = document.getElementById('btnAutonomousConfig');
  var btnEvent = document.getElementById('btnEventImpact');
  var btnScope = document.getElementById('btnScopeAdvance');
  var btnReplan = document.getElementById('btnReplan');
  if (btnTick) btnTick.classList.toggle('dim', !state.selectedWorldId);
  if (btnEvent) btnEvent.classList.toggle('dim', !state.selectedWorldId);
  if (btnScope) btnScope.classList.toggle('dim', !state.selectedWorldId);
  if (btnReplan) btnReplan.classList.toggle('dim', !state.selectedWorldId);
  if (btnAuto) btnAuto.classList.toggle('dim', !state.selectedWorldId || !state.selectedNodeId);
  if (btnAutoConfig) btnAutoConfig.classList.toggle('dim', !state.selectedWorldId || !state.selectedNodeId);
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
  container.innerHTML = ce("div", { className: "hint" }, [ttxt("Loading...")]);

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
        var detailCard = createToggleDetailCard([
          ce("span", { style: { fontWeight: 600 } }, [txt(t.task_type || "?")]),
          txt(" " + (t.duration_ms || 0) + "ms"),
          txt(" [" + (t.round != null ? tr('Rounds') + ' ' + t.round : "") + "]"),
          t.error ? ce("span", { style: { color: "var(--red)", marginLeft: "8px" } }, [txt(' ' + tr('Unknown error'))]) : null,
        ], false);

        var body = detailCard.body;

        // Basic info
        body.appendChild(renderPropRow('World', t.world_id ? t.world_id.slice(0, 8) : '-', { mono: true }));
        body.appendChild(renderPropRow('Node', t.node_id ? t.node_id.slice(0, 8) : '-', { mono: true }));
        body.appendChild(renderPropRow('Duration', (t.duration_ms || 0) + 'ms'));
        body.appendChild(renderPropRow('Pipeline', (t.configured_pipeline_mode || '-') + ' -> ' + (t.effective_pipeline_mode || '-')));
        body.appendChild(renderPropRow('Rounds', String(t.rounds_used || 0) + ' / ' + String(t.max_analysis_rounds || 0)));

        // System Prompt (collapsible)
        if (t.system_prompt) {
          body.appendChild(renderPreviewBlock('System Prompt', t.system_prompt, 1000));
        }

        // LLM Response (collapsible)
        if (t.raw_llm_response) {
          body.appendChild(renderPreviewBlock('LLM Response', t.raw_llm_response, 2000));
        }

        // Actions
        if (t.parsed_actions && t.parsed_actions.length > 0) {
          body.appendChild(renderPreviewListBlock('Actions', t.parsed_actions, function(a) {
            return a.action_id + ' ' + (a.mode ? '[' + a.mode + ']' : '');
          }));
        }

        // Memories
        if (t.parsed_memories && t.parsed_memories.length > 0) {
          body.appendChild(renderPreviewListBlock('Memories', t.parsed_memories, function(m) {
            var memoryNode = (m.node_id || '').slice(0, 8);
            var memoryContent = (m.content || '').slice(0, 100);
            return memoryNode + ' ' + m.level + ': ' + memoryContent + (m.content && m.content.length > 100 ? '...' : '');
          }));
        }

        container.appendChild(detailCard.card);
      })();
    }
  } catch (e) {
    container.innerHTML = "";
    container.appendChild(ce("div", { className: "hint" }, [txt(tr("Failed to load traces: ") + e.message)]));
  }
}

