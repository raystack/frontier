import yaml from 'js-yaml';
import fs from 'fs';
import { Parser } from '.';

export default function YMLParser(): Parser {
  return {
    parseFile: async (file: string) => {
      let config = {};
      try {
        config =
          yaml.load(fs.readFileSync(`${process.cwd()}/${file}`, 'utf8')) || {};
      } catch (e) {
        // eslint-disable-next-line no-console
        console.error(e);
      }
      return JSON.stringify(config);
    }
  };
}
