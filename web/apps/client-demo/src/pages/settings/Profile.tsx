import { ProfileView } from '@raystack/frontier/client';

export default function Profile() {
  return (
    <ProfileView
      enableUserDelete
      onDeleteSuccess={() => {
        window.location.href = '/login';
      }}
    />
  );
}
