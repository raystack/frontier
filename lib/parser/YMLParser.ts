/* eslint-disable no-unused-expressions */
/* eslint-disable no-console */
import yaml from 'js-yaml';
import { promises as fs } from 'fs';
import path from 'path';
import { Parser, FileContent } from '.';

export default function YMLParser(): Parser {
  const parseFile = async (filePath: string) => {
    let config = null;
    try {
      config = yaml.load(await fs.readFile(filePath, 'utf8')) || [];
    } catch (e) {
      // eslint-disable-next-line no-console
      console.error(e);
    }
    return config;
  };
  const parseFolder = async (folderName: string) => {
    let contents: FileContent[] = [];
    const folderPath = path.join(process.cwd(), folderName);

    try {
      const files = await fs.readdir(folderPath);
      const fileContents = await Promise.all(
        files.map(async (file) => {
          if (
            path.extname(file) === '.yaml' ||
            path.extname(file) === '.yaml'
          ) {
            return await parseFile(`${folderPath}/${file}`);
          }
          return [];
        })
      );

      fileContents.forEach((fileContent) => {
        Array.isArray(fileContent)
          ? (contents = [...contents, ...fileContent])
          : (contents = [...contents, fileContent]);
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
