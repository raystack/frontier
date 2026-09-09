import { Checkbox, Flex, Link, Text } from '@raystack/apsara';
import { Fragment, ReactNode } from 'react';
import type { ConsentDocument } from '@raystack/proton/frontier';
import styles from './auth-consent.module.css';

export type ConsentLabel =
  | ReactNode
  | ((documents: ConsentDocument[]) => ReactNode);

export type AuthConsentProps = {
  documents: ConsentDocument[];
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  label?: ConsentLabel;
};

const documentLinks = (documents: ConsentDocument[]) =>
  documents.map((document, index) => (
    <Fragment key={document.id}>
      {index > 0 && (index === documents.length - 1 ? ' and ' : ', ')}
      <Link
        href={document.url}
        external
        variant="primary"
        size="micro"
        data-test-id={`frontier-sdk-consent-document-${document.id}`}
      >
        {document.title}
      </Link>
    </Fragment>
  ));

const defaultLabel = (documents: ConsentDocument[]) => (
  <>I agree to the {documentLinks(documents)}</>
);

export const AuthConsent = ({
  documents,
  checked,
  onCheckedChange,
  label
}: AuthConsentProps) => {
  const content =
    typeof label === 'function'
      ? label(documents)
      : label ?? defaultLabel(documents);

  return (
    <Flex
      gap={4}
      align="start"
      justify="center"
      className={styles.container}
      render={<label />}
    >
      <Checkbox
        size="small"
        checked={checked}
        onCheckedChange={onCheckedChange}
        data-test-id="frontier-sdk-consent-checkbox"
      />
      <Text size="micro" variant="secondary">
        {content}
      </Text>
    </Flex>
  );
};
