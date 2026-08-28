const entries = ["api-smoke", "load"];

for (const name of entries) {
  const source = `tests/k6/${name}.ts`;
  const output = `tests/k6/${name}.js`;
  const result = await Bun.build({
    entrypoints: [source],
    target: "browser",
    external: ["k6", "k6/*", "https://jslib.k6.io/*"],
    minify: false,
    write: false,
  });

  if (!result.success || result.outputs.length !== 1) {
    console.error(`Failed to build ${source}`);
    for (const log of result.logs) console.error(log);
    process.exit(1);
  }

  await Bun.write(output, await result.outputs[0].text());
  console.log(`Built ${output}`);
}
