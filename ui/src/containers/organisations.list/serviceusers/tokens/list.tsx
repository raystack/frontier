import {
  Button,
  Checkbox,
  Flex,
  Separator,
  Table,
  Text,
} from "@raystack/apsara/v1";
import { Dialog } from "@raystack/apsara";
import styles from "./tokens.module.css";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1ServiceUserToken } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import dayjs from "dayjs";
import { DEFAULT_DATE_FORMAT } from "~/utils/constants";
import { Cross1Icon, TrashIcon } from "@radix-ui/react-icons";
import { toast } from "sonner";
import TableLoader from "~/components/TableLoader";

interface TokensListProps {
  organisationId: string;
  serviceUserId: string;
}

interface DeleteConfirmDialogProps {
  open: boolean;
  tokenId: string;
  onConfirm: (tokenId: string) => void;
}

function DeleteConfirmDialog({
  open,
  tokenId,
  onConfirm,
}: DeleteConfirmDialogProps) {
  const [isAcknowledged, setIsAcknowledged] = useState(false);
  function onClick() {
    onConfirm(tokenId);
  }

  function onCheckedChange(value: boolean) {
    setIsAcknowledged(value);
  }
  return (
    <Dialog open={open}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: "600px", width: "100%", zIndex: "60" }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: "16px 24px" }}>
          <Text size={6} style={{ fontWeight: "500" }}>
            Verify token deletion
          </Text>
          <Cross1Icon className={styles.crossIcon} />
        </Flex>
        <Separator />
        <Flex direction="column" gap="medium" style={{ padding: "24px 32px" }}>
          <Text size={2}>
            This action <b>can not</b> be undone. This will permanently delete
            the token. All services using this token will be <b>unauthorized</b>
          </Text>
          <Flex>
            <Checkbox
              checked={isAcknowledged}
              onCheckedChange={(v) => onCheckedChange(v === true)}
            ></Checkbox>
            <Text size={2}>I acknowledge to delete the service user token</Text>
          </Flex>
          <Button
            color="danger"
            type="submit"
            disabled={!isAcknowledged}
            style={{ width: "100%" }}
            onClick={onClick}
            data-test-id="admin-ui-delete-btn"
          >
            Delete
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}

export default function TokensList({
  organisationId,
  serviceUserId,
}: TokensListProps) {
  const { client } = useFrontier();
  const [tokens, setTokens] = useState<V1Beta1ServiceUserToken[]>([]);
  const [isTokensLoading, setIsTokensLoading] = useState(false);
  const [dialogState, setDialogState] = useState({ tokenId: "", open: false });

  const fetchTokens = useCallback(
    async (userId: string) => {
      try {
        setIsTokensLoading(true);
        const resp = await client?.frontierServiceListServiceUserTokens(
          organisationId,
          userId
        );
        const tokenList = resp?.data?.tokens || [];
        setTokens(tokenList);
      } catch (err) {
        console.error(err);
      } finally {
        setIsTokensLoading(false);
      }
    },
    [client]
  );

  useEffect(() => {
    if (serviceUserId) {
      fetchTokens(serviceUserId);
    }
  }, [serviceUserId, client, fetchTokens]);

  function openDeleteDialog(tokenId: string) {
    setDialogState({
      tokenId: tokenId,
      open: true,
    });
  }

  async function deleteToken(tokenId: string) {
    const resp = await client?.frontierServiceDeleteServiceUserToken(
      organisationId,
      serviceUserId,
      tokenId
    );

    if (resp?.status === 200) {
      toast.success("Token Deleted");
      await fetchTokens(serviceUserId);
    }

    setDialogState({
      tokenId: "",
      open: false,
    });
  }

  return (
    <div className={styles.tokensListWrapper}>
      <Text size={3} weight={500}>
        Tokens
      </Text>
      <Table className={styles.tokensTable}>
        <Table.Header>
          <Table.Row>
            <Table.Head className={styles.tableCell}>ID</Table.Head>
            <Table.Head className={styles.tableCell}>Title</Table.Head>
            <Table.Head className={styles.tableCell}>Created at</Table.Head>
            <Table.Head className={styles.tableCell}>Action</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {isTokensLoading ? (
            <TableLoader cell={4} cellClassName={styles.tableCell} />
          ) : (
            tokens.map((token) => (
              <Table.Row key={token.id}>
                <Table.Cell className={styles.tableCell}>{token.id}</Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  {token.title}
                </Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  {dayjs(token?.created_at).format(DEFAULT_DATE_FORMAT)}
                </Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  <TrashIcon
                    className={styles.deleteIcon}
                    onClick={() => openDeleteDialog(token?.id || "")}
                    data-test-id="admin-ui-trash-icon"
                  />
                </Table.Cell>
              </Table.Row>
            ))
          )}
        </Table.Body>
      </Table>
      <DeleteConfirmDialog
        open={dialogState.open}
        tokenId={dialogState.tokenId}
        onConfirm={deleteToken}
      />
    </div>
  );
}
