import { startRedis } from '../redis';

test('Redis startup', async () => {
  const client = await startRedis();

  expect(await client.ping()).toBe('PONG');
});
