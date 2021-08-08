import fetch from 'node-fetch';
import cheerio from 'cheerio';

interface Server {
  status: string;
  category: string;
  character_creation_status: string;
}

export interface Servers {
  [key: string]: Server;
}

/**
 * Fetches the server status page.
 * @returns The fetch response if it was successful.
 */
async function getStatus() {
  const resp = await fetch(
    'https://na.finalfantasyxiv.com/lodestone/worldstatus/'
  );

  if (resp.ok) {
    return resp;
  }

  throw new Error(
    `Could not fetch world status page. Reason: ${resp.statusText}`
  );
}

/**
 * Parses a server <li> element and retrieves relevant attributes.
 * @param {cheerio.Element} server
 * @param {cheerio.Root} $
 * @returns Object containing a particular server's status information
 */
function getServerInfo($: cheerio.Root, server: cheerio.Element): Server {
  const selector: cheerio.Cheerio = $(server);

  const category: string = selector
    .find('div .world-list__world_category')
    .find('p')
    .text();

  const status: string = selector
    .find('div .world-list__status_icon')
    .find('i')
    .attr('data-tooltip')!
    .trim();

  const character_creation_status: string = selector
    .find('div .world-list__create_character')
    .find('i')
    .attr('data-tooltip')!;

  return {
    status: status,
    category: category,
    character_creation_status: character_creation_status,
  };
}

/**
 * Scrapes the FFXIV Server Status page using Cheerio.
 * @returns Object containing server status information
 */
export async function scrapeStatusPage() {
  const servers: Servers = {};
  let $: cheerio.Root;

  try {
    const resp = await getStatus();
    $ = cheerio.load(await resp.text());
  } catch (error) {
    return {};
  }

  // Information for each server is nested in an
  // <li> element named `item-list`
  $('li .item-list').each((i, e: cheerio.Element) => {
    const name: string = $(e)
      .find('div .world-list__world_name')
      .find('p')
      .text();

    servers[name] = getServerInfo($, e);
  });

  return servers;
}
