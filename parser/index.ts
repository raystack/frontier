export interface Parser {
  parseFile: (file: string) => Promise<string>;
}
