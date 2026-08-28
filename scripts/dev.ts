const processes = [
  Bun.spawn(["bun", "run", "dev"], { cwd: "apps/api", stdout: "inherit", stderr: "inherit" }),
  Bun.spawn(["bun", "run", "dev"], { cwd: "apps/web", stdout: "inherit", stderr: "inherit" }),
];

const shutdown = () => {
  for (const process of processes) process.kill();
};

process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);

const exitCodes = await Promise.all(processes.map((process) => process.exited));
shutdown();
process.exit(exitCodes.some((code) => code !== 0) ? 1 : 0);
