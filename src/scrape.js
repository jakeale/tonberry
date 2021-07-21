import fetch from 'node-fetch';
import cheerio from 'cheerio';

/**
 * Fetches the server status page.
 * @returns The fetch response if it was successful.
 */
async function getStatus() {
  const resp = await fetch(
    'https://na.finalfantasyxiv.com/lodestone/worldstatus/'
  );

  if (!resp.ok) {
    throw Error('Could not fetch world status page.');
  }

  return resp;
}

/**
 * Parses a server <li> element and retrieves relevant attributes.
 * @param {CheerioAPI} $
 * @param {Element} server
 * @returns Object containing a particular server's status information
 */
function getServerInfo($, server) {
  const name = $(server).find('div .world-list__world_name').find('p').text();

  const category = $(server)
    .find('div .world-list__world_category')
    .find('p')
    .text();

  const status = $(server)
    .find('div .world-list__status_icon')
    .find('i')
    .attr('data-tooltip')
    .trim();

  return { name: name, status: status, category: category };
}

/**
 * Scrapes the FFXIV Server Status page using Cheerio.
 * @returns Object containing server status information
 */
export async function scrapeStatusPage() {
  const servers = {};

  const resp = await getStatus();

  const $ = cheerio.load(await resp.text());

  // Information for each server is nested in an
  // <li> element named `item-list`
  $('li .item-list').each((i, e) => {
    const server = getServerInfo($, e);

    servers[server.name] = {
      status: server.status,
      category: server.category,
    };
  });

  return servers;
}
