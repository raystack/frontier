import { SessionsView } from '@raystack/frontier/react';

export default function Sessions() {
  return <SessionsView onLogout={() => {
    window.location.href = '/login';
  }} />;
}
