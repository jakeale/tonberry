import { createClient } from 'redis';

export const startRedis = async () => {
  const client = createClient().on('error', (err) =>
    console.log('Redis Client error', err)
  );

  await client.connect();

  return client;
};
