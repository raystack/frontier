import { GeneralView } from '@raystack/frontier/client';
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
