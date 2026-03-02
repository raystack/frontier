'use client';

import { useNavigate, useParams } from '@tanstack/react-router';
import { ProjectDetailPage } from '~/react/views/projects';

export const ProjectPage = () => {
    const { projectId } = useParams({ from: '/projects/$projectId' });
    const navigate = useNavigate({ from: '/projects/$projectId' });

    return (
        <ProjectDetailPage
            projectId={projectId}
            onBack={() => navigate({ to: '/projects' })}
        />
    );
};
