/* eslint-disable @typescript-eslint/ban-types */

export type FileContent = object | string | number | null | undefined;
export interface Parser {
  parseFile: (file: string) => Promise<FileContent>;
  parseFolder: (folderName: string) => Promise<Array<FileContent>>;
}
