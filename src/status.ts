import { scrapeServerStatus, Servers } from './scrape.js';
import { diff } from 'deep-object-diff';
import { isEmpty } from './utils.js';

export class ServerStatusRefresher {
  servers: Servers = {};
  prev_servers: Servers = {};

  /**
   * Makeshift async constructor
   * @returns New instance of refresher
   */
  static async create() {
    const refresher = new ServerStatusRefresher();
    await refresher.init();
    return refresher;
  }

  /**
   * Start continuously refreshing the server status every 30 seconds
   */
  async init() {
    this.servers = await scrapeServerStatus();
    this.prev_servers = this.servers;
    setInterval(this.refresh.bind(this), 30000);
  }

  /**
   * Refresh server status
   */
  async refresh() {
    try {
      this.servers = await scrapeServerStatus();
    } catch (e) {
      console.error(e);
    }

    checkDiff(this.prev_servers, this.servers);

    this.prev_servers = this.servers;
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
