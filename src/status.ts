import { scrapeStatusPage, Servers } from './scrape.js';

export class ServerStatusRefresher {
  servers: Servers = {};

  constructor() {
    this.start();
  }

  get() {
    return this.servers;
  }

  async start() {
    this.servers = await scrapeStatusPage();
    this.continuouslyRefresh();
  }

  async continuouslyRefresh() {
    setInterval(async () => {
      this.servers = await scrapeStatusPage();
    }, 5000);
  }
}
