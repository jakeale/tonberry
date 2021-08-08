import { diff } from 'deep-object-diff';
import { ServerStatusRefresher } from './status.js';

async function main() {
  let servers = {};
  let prev = {};

  const refresher = new ServerStatusRefresher();
  setInterval(async () => {
    servers = refresher.get();
    console.log(diff(prev, servers));

    prev = JSON.parse(JSON.stringify(servers));
  }, 5000);
}

main();
