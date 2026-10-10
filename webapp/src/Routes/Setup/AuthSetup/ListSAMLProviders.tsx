import cx from 'clsx';
import { useState } from 'react';
import { Table, ScrollArea, Button, Tooltip, rem, Select } from '@mantine/core';
import classes from './ListSAMLProviders.module.css';
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

export function ListSAMLProviders() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const clipboard = useClipboard();
    const [scrolled, setScrolled] = useState(false);
    const [activeRow, setActiveRow] = useState<string>("")
    const [providerToDelete, setProviderToDelete] = useState<SAMLProvider | null>(null)
    const { isPending, error, data } = useQuery({
        queryKey: ['saml'],
        queryFn: () => fetchJSON(AppSettings.url + '/saml-setup', authInfo.token),
      })
    const deleteSAML = useMutation({
        mutationFn: (id:string) => {
          return axios.delete(AppSettings.url + '/saml-setup/'+id, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['saml'] })
        },
        onSettled: () => {
            setProviderToDelete(null)
        },
    })
    const mutation = useMutation({
      mutationFn: (payload:SAMLProvider) => {
        return axios.put(AppSettings.url + '/saml-setup/'+payload.id, payload, {
          headers: {
              "Authorization": "Bearer " + authInfo.token
          },
        })
      },
      onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ['saml'] })
      }
  })
    const clipboardCopy = (rowID:string, redirectURI:string) => {
      setActiveRow(rowID)
      clipboard.copy(redirectURI)
    }
    
    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message
    
    const rows = data.map((row:SAMLProvider) => (
        <Table.Tr key={row.id}>
            <Table.Td>{row.name}</Table.Td>
            <Table.Td>
                {row.metadataURL.substring(0, 40)}...
            </Table.Td>
            <Table.Td>
            <Select
               mt="md"
               data={["False","True"]}
               value={row.allowMissingAttributes ? "True" : "False"}
               allowDeselect={false}
               aria-label={"Allow missing attributes for " + row.name}
               onChange={(_value, option) => mutation.mutate({...row, allowMissingAttributes: option.value === "True" ? true : false})}
               comboboxProps={{ width: 100, position: 'bottom-start' }}
               styles={{ wrapper: { width: 100 } }}
               required
            />
            </Table.Td>
            <Table.Td>{row.acs.substring(0, 40)}...
              <Tooltip
                label="Link copied!"
                offset={5}
                position="bottom"
                radius="xl"
                transitionProps={{ duration: 100, transition: 'slide-down' }}
                opened={clipboard.copied && activeRow === row.id + "#acs"}
                ><Button
                variant="light"
                rightSection={
                  clipboard.copied && activeRow === row.id + "#acs" ? (
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
                onClick={() => clipboardCopy(row.id + "#acs", row.acs)}
              >
                Copy
              </Button>
              </Tooltip>
            </Table.Td>
            <Table.Td>{row.audience.substring(0, 40)}...
              <Tooltip
                label="Link copied!"
                offset={5}
                position="bottom"
                radius="xl"
                transitionProps={{ duration: 100, transition: 'slide-down' }}
                opened={clipboard.copied && activeRow === row.id  + "#aud"}
                ><Button
                variant="light"
                rightSection={
                  clipboard.copied && activeRow === row.id  + "#aud" ? (
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
                onClick={() => clipboardCopy(row.id + "#aud", row.audience)}
              >
                Copy
              </Button>
              </Tooltip>
            </Table.Td>
            <Table.Td>{row.issuer.substring(0, 40)}...
              <Tooltip
                label="Link copied!"
                offset={5}
                position="bottom"
                radius="xl"
                transitionProps={{ duration: 100, transition: 'slide-down' }}
                opened={clipboard.copied && activeRow === row.id + "#issuer"}
                ><Button
                variant="light"
                rightSection={
                  clipboard.copied && activeRow === row.id + "#issuer" ? (
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
                onClick={() => clipboardCopy(row.id + "#issuer", row.issuer)}
              >
                Copy
              </Button>
              </Tooltip>
            </Table.Td>
            <Table.Td><Button aria-label={"Delete SAML provider " + row.name} onClick={() => setProviderToDelete(row)}><TbTrash size={15} /></Button></Table.Td>
        </Table.Tr>
    ));

  
    return (
        <>
        <ConfirmModal
            opened={providerToDelete !== null}
            title="Delete SAML connection"
            confirmLabel="Delete"
            loading={deleteSAML.isPending}
            onClose={() => setProviderToDelete(null)}
            onConfirm={() => providerToDelete && deleteSAML.mutate(providerToDelete.id)}
        >
            Are you sure you want to delete the SAML connection "{providerToDelete?.name}"? Users will no longer be able to log in with this provider.
        </ConfirmModal>
        {deleteSAML.isError ? <ErrorAlert message={"An error occurred while deleting: " + getErrorMessage(deleteSAML.error)} /> : null}
        {mutation.isError ? <ErrorAlert message={"Could not update SAML connection: " + getErrorMessage(mutation.error)} /> : null}
        <ScrollArea h={300} onScrollPositionChange={({ y }) => setScrolled(y !== 0)}>
            <Table miw={700}>
            <Table.Thead className={cx(classes.header, { [classes.scrolled]: scrolled })}>
                <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>Metadata URL</Table.Th>
                <Table.Th>Allow Missing Attributes</Table.Th>
                <Table.Th>ACS URL</Table.Th>
                <Table.Th>Audience</Table.Th>
                <Table.Th>Issuer</Table.Th>
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{rows}</Table.Tbody>
            </Table>
        </ScrollArea>

        </>
    );
  
}