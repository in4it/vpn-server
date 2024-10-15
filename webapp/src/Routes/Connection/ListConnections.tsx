import cx from 'clsx';
import { useState } from 'react';
import { Table, ScrollArea, Button } from '@mantine/core';
import classes from './ListConnections.module.css';
import { useQueryClient, useMutation, useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';
import { TbTrash } from 'react-icons/tb';
import axios from 'axios';
import { useAuthContext } from '../../Auth/Auth';
import { Download } from './Download';

export function ListConnections() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const [scrolled, setScrolled] = useState(false);
    
    const { isPending, error, data } = useQuery({
        queryKey: ['connections'],
        queryFn: () =>
          fetch(AppSettings.url + '/vpn/connections', {
            headers: {
              "Content-Type": "application/json",
              "Authorization": "Bearer " + authInfo.token
            },
          }).then((res) => {
            return res.json()
            }
            
          ),
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
        }
      })
      
    
    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message
    
    const rows = data.map((row:OIDCProvider) => (
        <Table.Tr key={row.id}>
            <Table.Td>{row.name}</Table.Td>
            <Table.Td><Download id={row.id} name={row.name} /></Table.Td>
            <Table.Td><Button onClick={() => deleteConnection.mutate(row.id)}><TbTrash size={15} /></Button></Table.Td>
        </Table.Tr>
    ));

  
    return (
        <>
        {deleteConnection.isError ? (
            <div>An error occurred while deleting: {deleteConnection.error.message}</div>
          ) : null}
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