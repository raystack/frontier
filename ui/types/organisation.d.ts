export type Organisation = {
  id: string;
  name: string;
  slug: string;
  metadata: Record<string, string>;
  created_at: Date;
  updated_at: Date;
};
