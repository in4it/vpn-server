import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { AppSettings } from "../../Constants/Constants"
import { fetchJSON } from "../../Auth/fetchJSON"
import { useAuthContext } from "../../Auth/Auth";
import { ActionIcon, Group, Menu, Table, Text, rem } from "@mantine/core";
import { TbDots, TbTrash } from "react-icons/tb";
import axios from "axios";
import { useState } from "react";
import { getErrorMessage } from "../../Auth/errorMessage";
import { ConfirmModal } from "../../ConfirmModal";
import { ErrorAlert } from "../../ErrorAlert";

type Factor = {
    name: string;
    type: string;
  }

export function ListFactors() {
    const {authInfo} = useAuthContext();
    const queryClient = useQueryClient()
    const [factorToDelete, setFactorToDelete] = useState<string | null>(null)
    const { isPending, error, data } = useQuery({
        queryKey: ['factors'],
        queryFn: () => fetchJSON(AppSettings.url + '/profile/factors', authInfo.token),
    })
    const deleteFactor = useMutation({
        mutationFn: (id:string) => {
          return axios.delete(AppSettings.url + '/profile/factors/'+id, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['factors'] })
        },
        onSettled: () => {
            setFactorToDelete(null)
        },
      })


    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    const rows = data.map((item:Factor) => (
        <Table.Tr key={item.name}>
          <Table.Td>
            <Group gap="sm">
              <Text fz="sm" fw={500}>
                {item.name}
              </Text>
            </Group>
          </Table.Td>
          <Table.Td>
            <Group gap="sm">
              <Text fz="sm" fw={500}>
                {item.type}
              </Text>
            </Group>
          </Table.Td>
          <Table.Td>
                <Menu
                transitionProps={{ transition: 'pop' }}
                withArrow
                position="bottom-end"
                withinPortal
                >
                <Menu.Target>
                <ActionIcon variant="subtle" color="gray">
                    <TbDots style={{ width: rem(16), height: rem(16) }} />
                </ActionIcon>
                </Menu.Target>
                <Menu.Dropdown>
                <Menu.Item
                    leftSection={<TbTrash style={{ width: rem(16), height: rem(16) }} />}
                    color="red"
                    onClick={() => setFactorToDelete(item.name)}
                >
                    Delete Factor
                </Menu.Item>
                </Menu.Dropdown>
            </Menu>
            </Table.Td>
          </Table.Tr>
    ));

    return (
        <>
        <h2>Multifactor authentication</h2>
        <ConfirmModal
            opened={factorToDelete !== null}
            title="Delete MFA factor"
            confirmLabel="Delete"
            loading={deleteFactor.isPending}
            onClose={() => setFactorToDelete(null)}
            onConfirm={() => factorToDelete !== null && deleteFactor.mutate(factorToDelete)}
        >
            Are you sure you want to delete the factor "{factorToDelete}"? You will no longer be able to use it to log in.
        </ConfirmModal>
        {deleteFactor.isError ? <ErrorAlert message={"Could not delete factor: " + getErrorMessage(deleteFactor.error)} /> : null}
        <Table.ScrollContainer minWidth={300}>
                <Table verticalSpacing="sm">
                <Table.Thead>
                    <Table.Tr>
                    <Table.Th>Name</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>{rows}</Table.Tbody>
                </Table>
        </Table.ScrollContainer>
        </>
    )
}