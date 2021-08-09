import { ServerStatusRefresher } from './status.js';
import { Client } from 'discord.js';

const start = async () => {
  const refresher = await ServerStatusRefresher.create();
};

start();
