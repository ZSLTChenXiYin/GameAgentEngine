const http = require('http');

const port = Number(process.env.GAME_PORT || 9000);
const callbackToken = process.env.CALLBACK_TOKEN || 'dev-callback-token';
const engineBaseUrl = process.env.ENGINE_BASE_URL || 'http://127.0.0.1:8080';
const expectedAuth = `Bearer ${process.env.GAME_HTTP_BEARER_TOKEN || 'local-test-token'}`;

function readJson(req) {
  return new Promise((resolve, reject) => {
    let data = '';
    req.setEncoding('utf8');
    req.on('data', (chunk) => { data += chunk; });
    req.on('end', () => {
      try {
        resolve(data ? JSON.parse(data) : {});
      } catch (err) {
        reject(err);
      }
    });
    req.on('error', reject);
  });
}

async function callbackFetch(body, requestId) {
  const resp = await fetch(engineBaseUrl + '/api/v1/actions/callback', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Callback-Token': callbackToken,
      'X-Callback-Request-Id': requestId,
    },
    body: JSON.stringify(body),
  });
  const text = await resp.text();
  if (!resp.ok) {
    throw new Error(`callback failed: ${resp.status} ${text}`);
  }
  return text;
}

http.createServer(async (req, res) => {
  if (req.method !== 'POST' || req.url !== '/api/v1/runtime/dispatch') {
    res.statusCode = 404;
    res.end('not found');
    return;
  }

  if ((req.headers.authorization || '') !== expectedAuth) {
    res.statusCode = 401;
    res.end('unauthorized');
    return;
  }

  const body = await readJson(req);
  console.log('dispatch', JSON.stringify(body));

  if (body.callback_id) {
    const result = {
      source: 'runtime_task_push_receiver_example',
      echoed_payload: body.payload || null,
    };
    setTimeout(() => {
      callbackFetch({ callback_id: body.callback_id, status: 'success', result }, `cb-${body.task_id || Date.now()}`)
        .then((text) => console.log('callback', text))
        .catch((err) => console.error('callback error', err));
    }, 250);
  }

  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify({ status: 200, accepted: true }));
}).listen(port, '127.0.0.1', () => {
  console.log(`runtime task push receiver listening on ${port}`);
});
