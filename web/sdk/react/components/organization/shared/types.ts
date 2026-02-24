export type OnNavigate = (to: string, params?: Record<string, unknown>) => void;

export interface BasePageProps {
  organizationId: string;
  onNavigate?: OnNavigate;
}

