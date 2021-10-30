import cheerio from 'cheerio';
import axios, { AxiosResponse } from 'axios';

interface Server {
  status: string;
  category: string;
  characterCreationStatus: string;
}

export interface Servers {
  [key: string]: Server;
}

/**
 * Fetches the server status page.
 * @returns The fetch response if it was successful.
 */
const fetchStatus = async (): Promise<AxiosResponse<any, any>> => {
  const resp = await axios.get(
    'https://na.finalfantasyxiv.com/lodestone/worldstatus/',
    { responseType: 'text' }
  );

  if (resp.status === 200) {
    return resp;
  }

  throw new Error(`Could not fetch world status page. Reason: ${resp.status}`);
};

/**
 * Parses a server <li> element and retrieves relevant attributes.
 * @param {cheerio.Element} server
 * @param {cheerio.Root} $
 * @returns Object containing a particular server's status information
 */
const getServerInfo = ($: cheerio.Root, server: cheerio.Element): Server => {
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

  const characterCreationStatus: string = selector
    .find('div .world-list__create_character')
    .find('i')
    .attr('data-tooltip')!;

  return {
    status: status,
    category: category,
    characterCreationStatus: characterCreationStatus,
  };
};

/**
 * Scrapes the FFXIV Server Status page using Cheerio.
 * @returns Object containing server status information
 */
export const scrapeServerStatus = async () => {
  const servers: Servers = {};

  const resp = await fetchStatus();

  const $ = cheerio.load(resp.data);

  // Information for each server is nested in an
  // <li> element named `item-list`
  $('li .item-list').each((_i: number, e: cheerio.Element) => {
    const name: string = $(e)
      .find('div .world-list__world_name')
      .find('p')
      .text();

    servers[name] = getServerInfo($, e);
  });

  return servers;
};
