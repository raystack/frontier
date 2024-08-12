import {
  Button,
  Checkbox,
  Dialog,
  Flex,
  Separator,
  Table,
  Text,
} from "@raystack/apsara";
import styles from "./tokens.module.css";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1ServiceUserToken } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import dayjs from "dayjs";
import { DEFAULT_DATE_FORMAT } from "~/utils/constants";
import Skeleton from "react-loading-skeleton";
import { Cross1Icon, TrashIcon } from "@radix-ui/react-icons";
import { toast } from "sonner";

interface TokensListProps {
  serviceUserId: string;
}

interface TableLoaderProps {
  row?: number;
  cell?: number;
  cellClassName?: string;
}

function TableLoader({
  row = 5,
  cell = 3,
  cellClassName = "",
}: TableLoaderProps) {
  return (
    <>
      {[...new Array(row)].map((_, i) => (
        <Table.Row key={i}>
          {[...new Array(cell)].map((_, j) => (
            <Table.Cell className={cellClassName} key={i + "-" + j}>
              <Skeleton />
            </Table.Cell>
          ))}
        </Table.Row>
      ))}
    </>
  );
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
              onCheckedChange={onCheckedChange}
            ></Checkbox>
            <Text size={2}>I acknowledge to delete the service user token</Text>
          </Flex>
          <Button
            variant="danger"
            size="medium"
            type="submit"
            disabled={!isAcknowledged}
            style={{ width: "100%" }}
            onClick={onClick}
          >
            Delete
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}

export default function TokensList({ serviceUserId }: TokensListProps) {
  const { client } = useFrontier();
  const [tokens, setTokens] = useState<V1Beta1ServiceUserToken[]>([]);
  const [isTokensLoading, setIsTokensLoading] = useState(false);
  const [dialogState, setDialogState] = useState({ tokenId: "", open: false });

  const fetchTokens = useCallback(
    async (userId: string) => {
      try {
        setIsTokensLoading(true);
        const resp = await client?.frontierServiceListServiceUserTokens(userId);
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
