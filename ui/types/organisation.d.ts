export type Organisation = {
  id: string;
  name: string;
  slug: string;
  metadata: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
};
