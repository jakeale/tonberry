import cheerio from 'cheerio';
import axios, { AxiosResponse } from 'axios';

interface Server {
  status: string;
  category: string;
  characterCreationStatus: string;
}

interface DataCentre {
  [name: string]: Server;
}

/**
 * Object that maps names of data centres to their servers
 */
interface Servers {
  [name: string]: DataCentre;
}

/**
 * Fetches the server status page.
 * @returns The fetch response if it was successful.
 */
async function fetchStatus(): Promise<AxiosResponse<any, any>> {
  const resp = await axios.get(
    'https://na.finalfantasyxiv.com/lodestone/worldstatus/',
    { responseType: 'text' }
  );

  if (resp.status === 200) {
    return resp;
  }

  throw new Error(`Could not fetch world status page. Reason: ${resp.status}`);
}

/**
 * Parses a server <li> element and retrieves relevant attributes.
 * @param {cheerio.Element} server
 * @param {cheerio.Root} $
 * @returns Object containing a particular server's status information
 */
function getServerInfo($: cheerio.Root, server: cheerio.Element): Server {
  const selector: cheerio.Cheerio = $(server).parent();

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
}

/**
 * Scrapes the FFXIV Server Status page using Cheerio.
 * @returns Object containing status information for each server, organized by which data centre they belong to
 */
export async function scrapeServerStatus(): Promise<Servers> {
  const servers: Servers = Object();

  const resp = await fetchStatus();
  const $ = cheerio.load(resp.data);

  // Information for each server is nested in an
  // <li> element named `item-list`
  $('.world-dcgroup__item').each((_i: number, e: cheerio.Element) => {
    const dataCentre: string = $(e).find('h2').text();

    servers[dataCentre] = Object();

    $(e)
      .find('.world-list__world_name')
      .each((_i: number, e: cheerio.Element) => {
        servers[dataCentre][$(e).find('p').text()] = getServerInfo($, e);
      });
  });

  return servers;
}
