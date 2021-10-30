import { createClient } from 'redis';

export const startRedis = async () => {
  const client = await createClient();

  // client.on('error', (err) => console.log('Redis Client error', err));

  await client.connect();

  return client;
};
