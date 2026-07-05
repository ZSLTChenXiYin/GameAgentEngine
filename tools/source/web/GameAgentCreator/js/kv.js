/* ============= JSON Key-Value Helpers ============= */
function tryParseJSON(str) {
  if (!str || str.trim() === '') return null;
  try { return JSON.parse(str); } catch(e) { return null; }
}
function isJSON(str) {
  if (!str || str.trim() === '') return false;
  var s = str.trim();
  return (s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'));
}
function renderObjectKV(obj, editable, depth) {
  if (depth === undefined) depth = 0;
  /* no depth limit */
  if (obj === null) return txt('null');
  if (Array.isArray(obj)) {
    var container = ce('div', { className: 'kv-array' + (depth > 0 ? ' kv-nested' : ''), style: {} }, []);
    for (var i = 0; i < obj.length; i++) {
      var row = ce('div', { style: {display: 'flex', gap: '6px', padding: '2px 0', alignItems: 'flex-start'} }, [
        ce('span', { style: {color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '10px', minWidth: '20px'} }, [txt('[' + i + ']')]),
      ]);
      var item = obj[i];
      if (typeof item === 'object' && item !== null) row.appendChild(renderObjectKV(item, editable, depth + 1));
      else row.appendChild(ce('span', { style: {color: typeof item === 'string' ? 'var(--yellow)' : 'var(--accent)', fontSize: '10px'} }, [txt(String(item))]));
      container.appendChild(row);
    }
    if (obj.length === 0) container.appendChild(txt('[]'));
    return container;
  }
  var table = ce('div', { className: 'kv-object' + (depth > 0 ? ' kv-nested' : ''), style: {width: '100%'} }, []);
  var keys = Object.keys(obj);
  for (var k = 0; k < keys.length; k++) {
    var key = keys[k];
    var val = obj[key];
    var row = ce('div', { className: 'kv-row' }, []);
    var keyCell = ce('div', { className: 'kv-key', style: {color: 'var(--accent)', fontFamily: 'var(--font-mono)', fontSize: '10px', fontWeight: 600, whiteSpace: 'nowrap', padding: '2px 6px 2px 0', minWidth: '60px'} }, [txt(key)]);
    row.appendChild(keyCell);
    if (typeof val === 'object' && val !== null) {
      row.appendChild(ce('div', { className: 'kv-val', style: {padding: '0', verticalAlign: 'top', minWidth: 0} }, [renderObjectKV(val, editable, depth + 1)]));
    } else {
      var valStr = String(val);
      var valColor = typeof val === 'string' ? 'var(--yellow)' : (typeof val === 'number' ? 'var(--green)' : 'var(--text-dim)');
      row.appendChild(ce('div', { style: {padding: '2px 0', color: valColor, fontFamily: 'var(--font-mono)', fontSize: '10px', wordBreak: 'break-all'} }, [txt(valStr)]));
    }
    table.appendChild(row);
  }
  if (keys.length === 0) table.appendChild(txt('{}'));
  return table;
}
function renderKV(dataStr, editable) {
  var obj = tryParseJSON(dataStr);
  if (obj && typeof obj === 'object') return renderObjectKV(obj, editable);
  return txt(dataStr || '');
}
