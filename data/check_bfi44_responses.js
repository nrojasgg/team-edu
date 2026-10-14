const http = require('http');

function makeRequest(path) {
  return new Promise((resolve) => {
    http.get('http://localhost:3003/api/v1' + path, (res) => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => { try { resolve(JSON.parse(data)); } catch(e) { resolve({error:true}); } });
    });
  });
}

async function check() {
  const r = await makeRequest('/plugin/bfi44/tests/69dead572a057afd8f52496e/responses');
  console.log(JSON.stringify(r, null, 2));
}
check();