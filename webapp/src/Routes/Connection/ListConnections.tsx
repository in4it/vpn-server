import cx from 'clsx';
import { useState } from 'react';
import { Table, ScrollArea, Button } from '@mantine/core';
import classes from './ListConnections.module.css';
import { useQueryClient, useMutation, useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';
import { fetchJSON } from '../../Auth/fetchJSON';
import { getErrorMessage } from '../../Auth/errorMessage';
import { TbTrash } from 'react-icons/tb';
import axios from 'axios';
import { useAuthContext } from '../../Auth/Auth';
import { Download } from './Download';
import { ConfirmModal } from '../../ConfirmModal';
import { ErrorAlert } from '../../ErrorAlert';

type Connection = {
    id: string;
    name: string;
}

export function ListConnections() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const [scrolled, setScrolled] = useState(false);
    const [connectionToDelete, setConnectionToDelete] = useState<Connection | null>(null)
    
    const { isPending, error, data } = useQuery({
        queryKey: ['connections'],
        queryFn: () => fetchJSON(AppSettings.url + '/vpn/connections', authInfo.token),
      })
    const deleteConnection = useMutation({
        mutationFn: (id:string) => {
          return axios.delete(AppSettings.url + '/vpn/connection/'+id, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['connections'] })
            queryClient.invalidateQueries({ queryKey: ['connectionlicense'] })
        },
        onSettled: () => {
            setConnectionToDelete(null)
        },
      })
      
    
    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message
    
    const rows = data.map((row:Connection) => (
        <Table.Tr key={row.id}>
            <Table.Td>{row.name}</Table.Td>
            <Table.Td><Download id={row.id} name={row.name} /></Table.Td>
            <Table.Td><Button aria-label={"Delete connection " + row.name} onClick={() => setConnectionToDelete(row)}><TbTrash size={15} /></Button></Table.Td>
        </Table.Tr>
    ));

  
    return (
        <>
        <ConfirmModal
            opened={connectionToDelete !== null}
            title="Delete VPN connection"
            confirmLabel="Delete"
            loading={deleteConnection.isPending}
            onClose={() => setConnectionToDelete(null)}
            onConfirm={() => connectionToDelete && deleteConnection.mutate(connectionToDelete.id)}
        >
            Are you sure you want to delete the VPN connection "{connectionToDelete?.name}"? Devices using this connection will no longer be able to connect.
        </ConfirmModal>
        {deleteConnection.isError ? <ErrorAlert message={"An error occurred while deleting: " + getErrorMessage(deleteConnection.error)} /> : null}
        <ScrollArea h={300} onScrollPositionChange={({ y }) => setScrolled(y !== 0)}>
            <Table miw={700}>
            <Table.Thead className={cx(classes.header, { [classes.scrolled]: scrolled })}>
                <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>Download (invalidates older downloaded configs)</Table.Th>
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{rows}</Table.Tbody>
            </Table>
        </ScrollArea>

        </>
    );
  
}
