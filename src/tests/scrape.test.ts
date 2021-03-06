import { scrapeServerStatus } from '../scrape';

test('Scrape server status from Lodestone', async () => {
  const servers = await scrapeServerStatus();

  const expected_server = {
    category: expect.stringMatching(/Standard|Preferred|Congested|New/),
    characterCreationStatus: expect.stringMatching(
      /Creation of New Characters Available|Creation of New Characters Unavailable/
    ),
    status: expect.stringMatching(/Online|Offline/),
  };

  Object.values(servers).forEach((server) => {
    expect(server).toMatchObject(expected_server);
  });
});
