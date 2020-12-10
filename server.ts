import Composer from './config/composer';

const main = async () => {
  try {
    // Start the server
    const server = await Composer();
    await server.start();

    // Handle process exit
    process.on('SIGINT', async () => {
      await server.stop();
      process.exit(0);
    });
  } catch (err) {
    process.exit(1);
  }
};

main();
