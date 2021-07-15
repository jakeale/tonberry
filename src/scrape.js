import fetch from "node-fetch";
import cheerio from "cheerio";

/**
 * Fetches the server status page.
 * @returns The fetch response if it was successful.
 */
async function checkStatus() {
  return await fetch(
    "https://na.finalfantasyxiv.com/lodestone/worldstatus/"
  ).then((resp) => {
    if (resp.ok) {
      return resp;
    } else {
      throw Error("Couldn't fetch world status page.");
    }
  });
}

/**
 * Parses a server list element and retrieves relevant attributes.
 * @param {CheerioAPI} $
 * @param {Element} server
 * @returns Object containing a particular server's status information
 */
function getServerInfo($, server) {
  const name = $(server).find(".world-list__world_name").find("p").text();

  const category = $(server)
    .find(".world-list__world_category")
    .find("p")
    .text();

  const status = $(server)
    .find(".world-list__status_icon")
    .find("i")
    .attr("data-tooltip")
    .trim();

  return { name: name, status: status, category: category };
}

/**
 * Scrapes the FFXIV Server Status page using Cheerio.
 * @returns Object containing server status information
 */
export async function scrapeStatusPage() {
  let server_obj = {};

  const resp = await checkStatus();

  const $ = cheerio.load(await resp.text());

  // Information for each server is nested in an
  // <li> element named `item-list`
  const servers = $("li .item-list");

  servers.each((i, e) => {
    const server = getServerInfo($, e);

    server_obj[server.name] = {
      status: server.status,
      category: server.category,
    };
  });

  return server_obj;
}
