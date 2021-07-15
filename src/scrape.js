import fetch from "node-fetch";
import cheerio from "cheerio";

async function checkStatus() {
  return await fetch(
    "https://na.finalfantasyxiv.com/lodestone/worldstatus/"
  ).then(isOk);

  function isOk(resp) {
    if (resp.ok) {
      return resp;
    } else {
      console.log("something went wrong");
    }
  }
}

/**
 * Parses a server list element and retrieves relevant attributes
 * @param {CheerioAPI} $
 * @param {Element} server
 */
async function getServerInfo($, server) {
  const name = $(server).find(".world-list__world_name").find("p").text();

  const category = $(server)
    .find(".world-list__world_category")
    .find("p")
    .text();

  const icon = $(server)
    .find(".world-list__status_icon")
    .find("i")
    .attr("data-tooltip")
    .trim();
}

export default async function scrape() {
  const resp = await checkStatus();

  const $ = cheerio.load(await resp.text());

  const dataCenters = $("div .js--tab-content");

  dataCenters.each((i, e) => {
    const dataCenter = $(e).find(".world-dcgroup__item");

    dataCenter.each((i, e) => {
      const server = $(e).find(".item-list");

      server.each((i, e) => {
        getServerInfo($, e);
      });
    });
  });
}
