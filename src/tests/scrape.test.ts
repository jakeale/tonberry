import { scrapeServerStatus } from '../scrape';

test('Server status information is present for each server', async () => {
  const servers = await scrapeServerStatus();

  const expected_server = {
    category: expect.stringMatching(/Standard|Preferred|Congested|New/),
    characterCreationStatus: expect.stringMatching(
      /Creation of New Characters Available|Creation of New Characters Unavailable/
    ),
    status: expect.stringMatching(/Online|Offline/),
  };

  Object.values(servers).forEach((dataCentre) => {
    Object.values(dataCentre).forEach((server) => {
      expect(server).toMatchObject(expected_server);
    });
  });
});
