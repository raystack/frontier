import { Dialog, Image, Separator } from '@raystack/apsara';
import styles from './styles.module.css';
import { Flex, Text } from '@raystack/apsara/v1';
import { useNavigate, useParams } from '@tanstack/react-router';
import cross from '~/react/assets/cross.svg';

export default function ManageServiceUserProjects() {
  const { id } = useParams({
    from: '/api-keys/$id/projects'
  });
  const navigate = useNavigate({ from: '/api-keys/$id/projects' });

  function onCancel() {
    navigate({
      to: '/api-keys/$id',
      params: {
        id: id
      }
    });
  }

  return (
    <Dialog open={true}>
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.addDialogContent}
      >
        <Flex justify="between" className={styles.addDialogForm}>
          <Text size={6} weight={500}>
            Manage access
          </Text>

          <Image
            alt="cross"
            style={{ cursor: 'pointer' }}
            // @ts-ignore
            src={cross}
            onClick={onCancel}
            data-test-id="frontier-sdk-service-account-manage-access-close-btn"
          />
        </Flex>
        <Separator />
      </Dialog.Content>
    </Dialog>
  );
}
