export type Organisation = {
  id: string;
  name: string;
  slug: string;
  state: string;
  metadata: Record<string, string>;
  created_at: Date;
  updated_at: Date;
};
