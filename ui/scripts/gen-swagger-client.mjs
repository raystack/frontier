import path from "node:path";
import { generateApi } from "swagger-typescript-api";

const cwd = process.cwd();
const OUTPUT_PATH = path.resolve(cwd, "src", "api");
// {root}/proto/apidocs.swagger.yaml
const INPUT_PATH = path.resolve(cwd, "..", "proto", "apidocs.swagger.yaml");

async function main() {
  try {
    await generateApi({
      fileName: "frontier.ts",
      output: OUTPUT_PATH,
      input: INPUT_PATH,
      httpClientType: "axios",
    });
  } catch (error) {
    console.error("Error generating API:", error);
  }
}

main();
