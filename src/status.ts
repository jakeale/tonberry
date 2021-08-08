import { scrapeStatus, Servers } from './scrape';
import { diff } from 'deep-object-diff';
import { isEmpty } from './utils';

export class ServerStatusRefresher {
  servers: Servers = {};
  prev_servers: Servers = {};

  constructor() {
    this.start();
  }

  getServers() {
    return this.servers;
  }

  /**
   * Start continuously refreshing the server status every 30 seconds
   */
  async start() {
    this.servers = await scrapeStatus();
    this.prev_servers = this.servers;
    setInterval(this.refresh.bind(this), 30000);
  }

  /**
   * Refresh server status
   */
  async refresh() {
    try {
      this.servers = await scrapeStatus();

      checkDiff(this.prev_servers, this.servers);

      this.prev_servers = this.servers;
    } catch (e) {
      console.error(e);
    }
  }
}
/**
 * Check the difference between the last server status check and the current one.
 * @param prev `Servers` object containing the last check
 * @param current `Servers` object containing the current check
 */
const checkDiff = (prev: Servers, current: Servers) => {
  const status_diff = diff(prev, current);

  // If there's a difference, signal for bot message to be sent
  if (!isEmpty(status_diff)) {
    signal(status_diff);
  }
};

const signal = async (diff: object) => {};
