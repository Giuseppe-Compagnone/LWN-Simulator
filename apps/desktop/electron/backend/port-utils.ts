import net from "net";

export function isPortAvailable(port: number) {
  return new Promise<boolean>((resolve) => {
    const server = net.createServer();

    server.once("error", () => resolve(false));

    server.once("listening", () => {
      server.close(() => resolve(true));
    });

    server.listen(port, "127.0.0.1");
  });
}

export async function findAvailablePort(startPort = 8080) {
  let port = startPort;

  while (!(await isPortAvailable(port))) {
    port++;
  }

  return port;
}
