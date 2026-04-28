import { useParams, useNavigate } from 'react-router-dom';
import { PATDetailsView } from '@raystack/frontier/client';

export default function PatDetails() {
  const { orgId, patId } = useParams<{ orgId: string; patId: string }>();
  const navigate = useNavigate();

  if (!patId) return null;

  return (
    <PATDetailsView
      patId={patId}
      onNavigateToPats={() => navigate(`/${orgId}/settings/pats`)}
      onDeleteSuccess={() => navigate(`/${orgId}/settings/pats`)}
    />
  );
}
