import { Flex } from '@raystack/apsara';
import { UpdatesView } from '@raystack/frontier/client';

export default function Updates() {
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <UpdatesView />
    </Flex>
  );
}
