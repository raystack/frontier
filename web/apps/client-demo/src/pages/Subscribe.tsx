import { Flex } from '@raystack/apsara';
import { SubscribeView } from '@raystack/frontier/client';

export default function Subscribe() {
  return (
    <Flex
      justify="center"
      direction="column"
      align="center"
      style={{ width: '100vw', height: '95vh' }}
    >
      <SubscribeView />
    </Flex>
  );
}
