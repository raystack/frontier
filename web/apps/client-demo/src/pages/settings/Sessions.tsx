import { SessionsView } from '@raystack/frontier/client';

export default function Sessions() {
  return <SessionsView onLogout={() => {
    window.location.href = '/login';
  }} />;
}
