import path from "node:path";
import { generateApi } from "swagger-typescript-api";

const cwd = process.cwd();
const OUTPUT_PATH = path.resolve(cwd, "api-client");
// {root}/proto/apidocs.swagger.yaml
const INPUT_PATH = path.resolve(cwd, "..", "..", "..", "..", "proto", "apidocs.swagger.yaml");

async function main() {
  try {
    await generateApi({
      output: OUTPUT_PATH,
      input: INPUT_PATH,
      modular: true
    });
  } catch (error) {
    console.error("Error generating API:", error);
  }
}

main();
