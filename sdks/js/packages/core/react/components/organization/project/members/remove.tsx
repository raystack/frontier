import { Dialog, Image, Separator } from '@raystack/apsara/*';
import styles from '../../organization.module.css';
import { Flex, Text } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';

export const RemoveProjectMember = () => {
  const navigate = useNavigate({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });

  const { projectId } = useParams({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });
  return (
    <Dialog open={true}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} weight={500}>
            Remove project member
          </Text>
          <Image
            data-test-id="frontier-sdk-remove-project-member-close-btn"
            alt="cross"
            // @ts-ignore
            src={cross}
            onClick={() =>
              navigate({
                to: '/projects/$projectId',
                params: { projectId }
              })
            }
            style={{ cursor: 'pointer' }}
          />
        </Flex>
        <Separator />
      </Dialog.Content>
    </Dialog>
  );
};
