/* eslint-disable no-unused-expressions */
/* eslint-disable no-console */
import yaml from 'js-yaml';
import fs from 'fs';
import path from 'path';
import { Parser, FileContent } from '.';

export default function YMLParser(): Parser {
  const parseFile = (filePath: string) => {
    let config = null;
    try {
      config = yaml.load(fs.readFileSync(filePath, 'utf8')) || [];
    } catch (e) {
      // eslint-disable-next-line no-console
      console.error(e);
    }
    return config;
  };
  const parseFolder = (folderName: string) => {
    let contents: FileContent[] = [];
    const folderPath = `${process.cwd()}/${folderName}`;

    try {
      const files = fs.readdirSync(folderPath);
      files.forEach((file) => {
        // Do whatever you want to do with the .yml file
        if (path.extname(file) === '.yml') {
          const fileContent = parseFile(`${folderPath}/${file}`);
          Array.isArray(fileContent)
            ? (contents = [...contents, ...fileContent])
            : (contents = [...contents, fileContent]);
        }
      });
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log(`Unable to scan directory: ${err}`);
    }
    return contents;
  };

  return {
    parseFile,
    parseFolder
  };
}
