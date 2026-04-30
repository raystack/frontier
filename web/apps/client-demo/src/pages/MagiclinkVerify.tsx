import { Flex } from '@raystack/apsara';
import { MagicLinkVerifyView } from '@raystack/frontier/client';

export default function MagiclinkVerify() {
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <MagicLinkVerifyView />
    </Flex>
  );
}
