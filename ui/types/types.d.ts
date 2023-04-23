import { ReactNode } from "react";

export type TableColumnMetadata = {
  name: ReactNode | Element;
  key: string;
  value: string;
};
