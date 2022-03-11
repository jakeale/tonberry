import { createClient } from 'redis';

export const startRedis = async () => {
  const client = await createClient();

  await client.connect();

  return client;
};
