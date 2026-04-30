import { useParams, useNavigate } from 'react-router-dom';
import { PatsView } from '@raystack/frontier/react';

export default function Pats() {
  const { orgId } = useParams<{ orgId: string }>();
  const navigate = useNavigate();

  return (
    <PatsView
      onPATClick={patId => navigate(`/${orgId}/settings/pats/${patId}`)}
    />
  );
}
