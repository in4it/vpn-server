import cx from 'clsx';
import { useState } from 'react';
import { Table, ScrollArea, Button, Tooltip, rem } from '@mantine/core';
import classes from './ListOIDCProviders.module.css';
import { useQueryClient, useMutation, useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../../Constants/Constants';
import { fetchJSON } from '../../../Auth/fetchJSON';
import { TbCheck, TbCopy, TbTrash } from 'react-icons/tb';
import { useClipboard } from '@mantine/hooks';
import axios from 'axios';
import { useAuthContext } from '../../../Auth/Auth';
import { getErrorMessage } from '../../../Auth/errorMessage';
import { ConfirmModal } from '../../../ConfirmModal';
import { ErrorAlert } from '../../../ErrorAlert';

export function ListOIDCProviders() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const clipboard = useClipboard();
    const [scrolled, setScrolled] = useState(false);
    const [activeRow, setActiveRow] = useState<string>("")
    const [providerToDelete, setProviderToDelete] = useState<OIDCProvider | null>(null)
    const { isPending, error, data } = useQuery({
        queryKey: ['oidc'],
        queryFn: () => fetchJSON(AppSettings.url + '/oidc', authInfo.token),
      })
    const mutation = useMutation({
        mutationFn: (id:string) => {
          return axios.delete(AppSettings.url + '/oidc/'+id, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['oidc'] })
        },
        onSettled: () => {
            setProviderToDelete(null)
        },
      })
    const clipboardCopy = (rowID:string, redirectURI:string) => {
      setActiveRow(rowID)
      clipboard.copy(redirectURI)
    }
    
    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message
    
    const rows = data.map((row:OIDCProvider) => (
        <Table.Tr key={row.id}>
            <Table.Td>{row.name}</Table.Td>
            <Table.Td>{row.clientId}</Table.Td>
            <Table.Td>{row.redirectURI.substring(0, 40)}...
              <Tooltip
                label="Link copied!"
                offset={5}
                position="bottom"
                radius="xl"
                transitionProps={{ duration: 100, transition: 'slide-down' }}
                opened={clipboard.copied && activeRow === row.id + "#redirectURI"}
                ><Button
                variant="light"
                rightSection={
                  clipboard.copied && activeRow === row.id + "#redirectURI" ? (
                    <TbCheck style={{ width: rem(10), height: rem(10) }} />
                  ) : (
                    <TbCopy style={{ width: rem(10), height: rem(10) }} />
                  )
                }
                radius="xl"
                size="md-1"
                styles={{
                  root: { paddingRight: rem(7), height: rem(24) },
                  section: { marginLeft: rem(22) },
                }}
                onClick={() => clipboardCopy(row.id + "#redirectURI", row.redirectURI)}
              >
                Copy
              </Button>
              </Tooltip>
            </Table.Td>
            <Table.Td>{row.loginURL.substring(0, 40)}...
              <Tooltip
                label="Link copied!"
                offset={5}
                position="bottom"
                radius="xl"
                transitionProps={{ duration: 100, transition: 'slide-down' }}
                opened={clipboard.copied && activeRow === row.id + "#loginURL"}
                ><Button
                variant="light"
                rightSection={
                  clipboard.copied && activeRow === row.id + "#loginURL" ? (
                    <TbCheck style={{ width: rem(10), height: rem(10) }} />
                  ) : (
                    <TbCopy style={{ width: rem(10), height: rem(10) }} />
                  )
                }
                radius="xl"
                size="md-1"
                styles={{
                  root: { paddingRight: rem(7), height: rem(24) },
                  section: { marginLeft: rem(22) },
                }}
                onClick={() => clipboardCopy(row.id + "#loginURL", row.loginURL)}
              >
                Copy
              </Button>
              </Tooltip>
            </Table.Td>
            <Table.Td>{row.scope}</Table.Td>
            <Table.Td><Button aria-label={"Delete OIDC provider " + row.name} onClick={() => setProviderToDelete(row)}><TbTrash size={15} /></Button></Table.Td>
        </Table.Tr>
    ));

  
    return (
        <>
        <ConfirmModal
            opened={providerToDelete !== null}
            title="Delete OIDC connection"
            confirmLabel="Delete"
            loading={mutation.isPending}
            onClose={() => setProviderToDelete(null)}
            onConfirm={() => providerToDelete && mutation.mutate(providerToDelete.id)}
        >
            Are you sure you want to delete the OIDC connection "{providerToDelete?.name}"? Users will no longer be able to log in with this provider.
        </ConfirmModal>
        {mutation.isError ? <ErrorAlert message={"An error occurred while deleting: " + getErrorMessage(mutation.error)} /> : null}
        <ScrollArea h={300} onScrollPositionChange={({ y }) => setScrolled(y !== 0)}>
            <Table miw={700}>
            <Table.Thead className={cx(classes.header, { [classes.scrolled]: scrolled })}>
                <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>ClientID</Table.Th>
                <Table.Th>RedirectURI</Table.Th>
                <Table.Th>LoginURL</Table.Th>
                <Table.Th>Scope</Table.Th>
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{rows}</Table.Tbody>
            </Table>
        </ScrollArea>

        </>
    );
  
}