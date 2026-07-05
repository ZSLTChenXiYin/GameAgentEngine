/* ============= Right Panel ============= */
function renderRightPanel() {
  const rp = document.getElementById('rightPanel');
  rp.innerHTML = '';
  const summ = ce('div', { className: 'summ' }, []);
  summ.appendChild(ce('div', { className: 'summ-chip' }, [ce('span', { className: 'l' }, [ttxt('Worlds')]), ce('span', { className: 'v' }, [txt(String(state.worlds.length))])]));
  summ.appendChild(ce('div', { className: 'summ-chip' }, [ce('span', { className: 'l' }, [ttxt('Nodes')]), ce('span', { className: 'v' }, [txt(String(state.nodes.length))])]));
  if (state.selectedNodeId && state.nodeDetail) {
    const nd = state.nodeDetail;
    summ.appendChild(ce('div', { className: 'summ-chip' }, [ce('span', { className: 'l' }, [ttxt('Components')]), ce('span', { className: 'v' }, [txt(String(nd.components ? nd.components.length : 0))])]));
    summ.appendChild(ce('div', { className: 'summ-chip' }, [ce('span', { className: 'l' }, [ttxt('Memories')]), ce('span', { className: 'v' }, [txt(String(nd.memories ? nd.memories.length : 0))])]));
  }
  rp.appendChild(summ);
  if (state.selectedNodeId && state.nodeDetail) {
    const nd = state.nodeDetail;
    const compDock = ce('div', { className: 'dock' }, [
      ce('button', { className: 'dock-hd', 'aria-expanded': 'true' }, [function(){var c=nd.components?nd.components.length:0;return ce('span',{},[ttxt('Components'),txt(' ('+c+')')])}()]),
      ce('div', { className: 'dock-bd' }, []),
    ]);
    rp.appendChild(compDock);
    const compBody = compDock.querySelector('.dock-bd');
    if (nd.components) { for (const c of nd.components) {
      compBody.appendChild(ce('div', { className: 'dock-cnt comp-row' }, [
        ce('div', { className: 'rp-item-hd' }, [
          ce('span', { style: {color: 'var(--accent)', fontWeight: 600} }, [txt(c.component_type)]),
          ce('button', { className: 'dock-del', dataset: { id: c.id }, style: {fontSize: '10px', color: 'var(--red)', background: 'none', border: 'none', cursor: 'pointer'} }, [ttxt('\u2715')]),
        ]),
      ]));
      var _cr = compBody.lastChild;
      if (isJSON(c.data)) { _cr.appendChild(renderKV(c.data)); } else { _cr.appendChild(txt(c.data || '')); }
    } }
    const memDock = ce('div', { className: 'dock' }, [
      ce('button', { className: 'dock-hd', 'aria-expanded': 'true' }, [function(){var c=nd.memories?nd.memories.length:0;return ce('span',{},[ttxt('Memories'),txt(' ('+c+')')])}()]),
      ce('div', { className: 'dock-bd' }, []),
    ]);
    rp.appendChild(memDock);
    const memBody = memDock.querySelector('.dock-bd');
    if (nd.memories) { for (const m of nd.memories) {
      memBody.appendChild(ce('div', { className: 'dock-cnt mem-row' }, [
        ce('div', { className: 'rp-item-hd' }, [
          ce('span', { style: {color: 'var(--yellow)', fontWeight: 600} }, [txt(m.level || 'long_term')]),
          ce('button', { className: 'dock-del', dataset: { id: m.id }, style: {fontSize: '10px', color: 'var(--red)', background: 'none', border: 'none', cursor: 'pointer'} }, [ttxt('\u2715')]),
        ]),
      ]));
      var _mr = memBody.lastChild;
      if (isJSON(m.content)) { _mr.appendChild(renderKV(m.content)); } else { _mr.appendChild(txt(m.content || '')); }
    } }
    const relDock = ce('div', { className: 'dock' }, [
      ce('button', { className: 'dock-hd', 'aria-expanded': 'true' }, [function(){var c=nd.relations?nd.relations.length:0;return ce('span',{},[ttxt('Relations'),txt(' ('+c+')')])}()]),
      ce('div', { className: 'dock-bd' }, []),
    ]);
    rp.appendChild(relDock);
    const relBody = relDock.querySelector('.dock-bd');
    if (nd.relations) { for (const r of nd.relations) {
      relBody.appendChild(ce('div', { className: 'dock-cnt rel-row' }, [
        ce('div', { className: 'rp-item-hd' }, [
          ce('span', { style: {color: 'var(--green)', fontWeight: 600} }, [txt(r.relation_type)]),
          txt(' -> ' + r.target_id.substring(0, 8)),
          ce('button', { className: 'dock-del', dataset: { id: r.id }, style: {fontSize: '10px', color: 'var(--red)', background: 'none', border: 'none', cursor: 'pointer'} }, [ttxt('\u2715')]),
        ]),
      ]));
      var _rr = relBody.lastChild;
    } }
    if (state.autonomous && state.autonomous.config) {
      const ac = state.autonomous.config;
      const autoDock = ce('div', { className: 'dock' }, [
        ce('button', { className: 'dock-hd', 'aria-expanded': 'true' }, [ttxt('Autonomous')]),
        ce('div', { className: 'dock-bd' }, [
          ce('div', { className: 'dock-cnt' }, [txt('Enabled: ' + (ac.enabled ? 'Yes' : 'No') + ' | Trigger: ' + ac.trigger + (ac.interval_seconds ? ' | Interval: ' + ac.interval_seconds + 's' : ''))]),
          ce('button', { className: 'dock-edit', dataset: { id: state.selectedNodeId }, style: {fontSize: '10px', background: 'none', border: 'none', cursor: 'pointer', color: 'var(--accent)', marginLeft: 'auto'} }, [ttxt('Edit Config')]),
        ]),
      ]);
      rp.appendChild(autoDock);
    }
  }
  rp.querySelectorAll('.dock-hd').forEach(function(hd) {
    hd.addEventListener('click', function() {
      var expanded = hd.getAttribute('aria-expanded') === 'true';
      hd.setAttribute('aria-expanded', expanded ? 'false' : 'true');
    });
  });
  rp.querySelectorAll('.dock-del').forEach(function(btn) {
    btn.addEventListener('click', function(e) {
      e.stopPropagation();
      var parent = btn.closest('.dock'); if (!parent) return;
      var hdText = parent.querySelector('.dock-hd').textContent;
      if (hdText.startsWith('Components') || hdText.indexOf(tr('Components')) === 0) deleteComponent(btn.dataset.id);
      else if (hdText.startsWith('Memories') || hdText.indexOf(tr('Memories')) === 0) deleteMemory(btn.dataset.id);
      else if (hdText.startsWith('Relations') || hdText.indexOf(tr('Relations')) === 0) deleteRelation(btn.dataset.id);
    });
  });
}

/* ============= Init Layout ============= */
function initLayout() {
  const app = document.getElementById('app');
  app.innerHTML = '';
  app.appendChild(el('div', { id: 'topbar', className: 'topbar' }));
  const row = el('div', { className: 'main-row' });
  row.appendChild(el('div', { id: 'leftPanel', className: 'left-panel' }));
  const rh = el('div', { className: 'resize-h', id: 'resizeH' });
  row.appendChild(rh);
  const center = el('div', { className: 'center' });
  center.appendChild(el('div', { id: 'centerContent', className: 'center-scroll' }));
  row.appendChild(center);
  row.appendChild(el('div', { className: 'resize-h', id: 'resizeRight' }));
  row.appendChild(el('div', { id: 'rightPanel', className: 'right-panel' }));
  app.appendChild(row);
  var draggingH = false;
  rh.addEventListener('mousedown', function(e) {
    draggingH = true;
    document.body.style.cursor = 'col-resize';
    document.getElementById('resizeH').classList.add('dragging');
  });
  document.addEventListener('mousemove', function(e) {
    if (!draggingH) return;
    var rect = app.getBoundingClientRect();
    var w = e.clientX - rect.left;
    if (w < 160) w = 160;
    if (w > rect.width - 360) w = rect.width - 360;
    document.querySelector('.left-panel').style.width = w + 'px';
  });
  document.addEventListener('mouseup', function() {
    if (draggingH) {
      draggingH = false;
      document.body.style.cursor = '';
      document.getElementById('resizeH').classList.remove('dragging');
    }
  });
  var draggingRight = false;
  const rh2 = document.getElementById('resizeRight');
  rh2.addEventListener('mousedown', function() { draggingRight = true; document.body.style.cursor = 'col-resize'; rh2.classList.add('dragging'); });
  document.addEventListener('mousemove', function(e) {
    if (!draggingRight) return;
    const rect = app.getBoundingClientRect();
    var rpw = rect.right - e.clientX;
    if (rpw < 180) rpw = 180;
    const maxRP = rect.width - 300;
    if (rpw > maxRP) rpw = maxRP;
    document.querySelector('.right-panel').style.width = rpw + 'px';
  });
  document.addEventListener('mouseup', function() {
    if (draggingRight) { draggingRight = false; document.body.style.cursor = ''; rh2.classList.remove('dragging'); }
  });
}

/* ============= Init ============= */
async function init() {
  document.getElementById('modalCloseBtn').addEventListener('click', closeModal);
  initLayout();
  renderTopbar();
  renderLeftPanel();
  checkHealth();
  checkEngineVersion(); // check engine version compatibility
  loadWorlds();
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') { closeModal(); hideContextMenu(); }
  });
}

document.addEventListener('DOMContentLoaded', init);
