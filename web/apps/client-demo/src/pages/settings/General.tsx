import { GeneralView } from '@raystack/frontier/react';
import { useNavigate } from 'react-router-dom';

export default function General() {
  const navigate = useNavigate();

  return (
    <GeneralView
      onDeleteSuccess={() => {
        navigate('/');
      }}
    />
  );
}
