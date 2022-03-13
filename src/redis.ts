import { createClient } from 'redis';

export async function startRedis() {
  const client = await createClient();

  await client.connect();

  return client;
}
