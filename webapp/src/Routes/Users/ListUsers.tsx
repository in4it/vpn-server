import { useState } from 'react';
import { Table, Button, rem, Group, Text, Badge, Select, PasswordInput, Menu, ActionIcon, Modal, Container, Space, Alert } from '@mantine/core';
//import classes from './ListUsers.module.css';
import { useQueryClient, useMutation, useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';
import { fetchJSON } from '../../Auth/fetchJSON';
import { getErrorMessage } from '../../Auth/errorMessage';
import { TbDots, TbInfoCircle, TbPassword, TbTrash, TbUserPause } from 'react-icons/tb';
import { useDisclosure } from '@mantine/hooks';
import axios from 'axios';
import { useAuthContext } from '../../Auth/Auth';
import { ConfirmModal } from '../../ConfirmModal';
import { ErrorAlert } from '../../ErrorAlert';

type Props = {
    localAuthDisabled: boolean;
};

type UserIDAndPassword = {
  id: string;
  password: string;
}

type Status = {
  status: string;
  color: string;
}

export function ListUsers({localAuthDisabled}:Props) {
    const [opened, { open, close }] = useDisclosure(false);
    const [newPassword, setNewPassword] = useState<string>("");
    const [passwordError, setPasswordError] = useState<string>("");
    const [userSelected, setUserSelected] = useState("")
    const [userToDelete, setUserToDelete] = useState<User | null>(null)
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const [passwordUpdated, setPasswordUpdated] = useState(false);
    const { isPending, error, data } = useQuery({
        queryKey: ['users'],
        queryFn: () => fetchJSON(AppSettings.url + '/users', authInfo.token),
      })

    const updateUser = useMutation({
        mutationFn: (user:User) => {
          return axios.patch(AppSettings.url + '/user/'+user.id, user, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        // show the new role/suspended state right away, and roll back if the
        // backend rejects the change
        onMutate: async (user:User) => {
            await queryClient.cancelQueries({ queryKey: ['users'] })
            const previousUsers = queryClient.getQueryData<User[]>(['users'])
            queryClient.setQueryData<User[]>(['users'], (users) =>
              users?.map((u) => u.id === user.id ? {...u, role: user.role, suspended: user.suspended} : u)
            )
            return { previousUsers }
        },
        onError: (_error, _user, context) => {
            if (context?.previousUsers !== undefined) {
              queryClient.setQueryData(['users'], context.previousUsers)
            }
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
        }
      })
      const changePassword = useMutation({
        mutationFn: (userIDAndPassword:UserIDAndPassword) => {
          return axios.patch(AppSettings.url + '/user/'+userIDAndPassword.id, userIDAndPassword, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            setPasswordUpdated(true)
            setNewPassword("")
        }
      })
      const deleteUser = useMutation({
        mutationFn: (id:string) => {
          return axios.delete(AppSettings.url + '/user/'+id, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] })
            queryClient.invalidateQueries({ queryKey: ['license'] })
        },
        onSettled: () => {
            setUserToDelete(null)
        },
      })
      const alertIcon = <TbInfoCircle />;
      
    
    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    const typeColor: Record<string, string> = {
        local: 'blue',
        oidc: 'cyan',
        saml: 'teal',
        provisioned: 'grape',
    };

    const rolesData = ["user", "admin"];

    // every time the modal opens it starts empty, so a password typed for
    // one user can never be submitted for another
    const openPasswordModal = (id:string) => {
        setPasswordUpdated(false)
        setNewPassword("")
        setPasswordError("")
        changePassword.reset()
        setUserSelected(id)
        open()
    }
    const closePasswordModal = () => {
        setNewPassword("")
        setPasswordError("")
        close()
    }
    const submitPassword = () => {
        if (newPassword === "") {
            setPasswordError("Password cannot be empty")
            return
        }
        setPasswordError("")
        changePassword.mutate({id: userSelected, password: newPassword})
    }
    const formatDate = (dateInput:string) => {
      const date = new Date(dateInput)
      return date.getFullYear() + "-" + (date.getMonth()+1).toString().padStart(2, '0') + "-" + date.getDate().toString().padStart(2, '0') + " " + date.getHours().toString().padStart(2, '0') + ":" + date.getMinutes().toString().padStart(2, '0') + ":" + date.getSeconds().toString().padStart(2, '0')
    }

    const getStatus = (item:User) : Status => {
      if(item.suspended) {
        return {status: "Suspended", color: "red"} 
      }
      if (localAuthDisabled && item.oidcID == "" && !item.provisioned) {
        return {status: "Local Auth Disabled", color: "yellow"} 
      }
      return {status: "Active", color: "green"}
    }
      
    const rows = data.map((item:User) => (
          <Table.Tr key={item.id}>
            <Table.Td>
              <Group gap="sm">
                <Text fz="sm" fw={500}>
                  {item.login}
                </Text>
              </Group>
            </Table.Td>
            <Table.Td>
            <Select
               data={rolesData}
               value={item.role}
               aria-label={"Role of " + item.login}
               variant="unstyled"
               allowDeselect={false}
               onChange={(event) => updateUser.mutate({...item, id: item.id, password: "", role: event === null ? "" : event, lastLogin: "0001-01-01T00:00:00Z"})}
            />
            </Table.Td>
            <Table.Td>
              {item.oidcID == "" && !item.provisioned ? 
                <Badge color={typeColor["local"]} variant="light" style={{marginRight: 5}}>
                  local user
                </Badge>
              : null}
              {item.oidcID !== "local" && item.oidcID !== "" ? 
                <Badge color={typeColor["oidc"]} variant="light" style={{marginRight: 5}}>
                  oidc user
                </Badge>
              : null}
             {item.samlID !== "" ? 
                <Badge color={typeColor["saml"]} variant="light" style={{marginRight: 5}}>
                  saml user
                </Badge>
              : null}
              {item.provisioned ? 
                <Badge color={typeColor["provisioned"]} variant="light" style={{marginRight: 5}}>
                  provisioned
                </Badge>
              : null}
            </Table.Td>
            <Table.Td>
              <Badge color={getStatus(item).color} variant="light">
                {getStatus(item).status}
              </Badge>
            </Table.Td>
            <Table.Td>
              <Group gap="sm">
                <Text fz="sm" fw={500}>
                  {item.lastLogin === "" ? "Never logged in" : (new Date(item.lastLogin)).toLocaleString() }
                  {item.oidcID == "" ? "" : 
                    item.lastTokenRenewal !== "" && item.lastTokenRenewal !== "0001-01-01T00:00:00Z" ? 
                      item.connectionsDisabledOnAuthFailure ? " (last OIDC refresh: failed)": " (last OIDC refresh: "+formatDate(item.lastTokenRenewal) + ")" : ""
                  }
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
                {item.oidcID == "" ? 
                <Menu.Item
                    leftSection={<TbPassword style={{ width: rem(16), height: rem(16) }} />}
                    onClick={() => openPasswordModal(item.id)}
                >
                    Change Password
                </Menu.Item>
                : null }
                <Menu.Item
                    leftSection={<TbUserPause style={{ width: rem(16), height: rem(16) }} />}
                    onClick={() => updateUser.mutate({...item, id: item.id, password: "", suspended: !item.suspended})}
                >
                    {item.suspended ? "Unsuspend" : "Suspend"} User
                </Menu.Item>
                <Menu.Item
                    leftSection={<TbTrash style={{ width: rem(16), height: rem(16) }} />}
                    color="red"
                    onClick={() => setUserToDelete(item)}
                >
                    Delete User
                </Menu.Item>
                </Menu.Dropdown>
            </Menu>
            </Table.Td>
          </Table.Tr>
        ));
      
        return (
          <>
            <ConfirmModal
                opened={userToDelete !== null}
                title="Delete user"
                confirmLabel="Delete"
                loading={deleteUser.isPending}
                onClose={() => setUserToDelete(null)}
                onConfirm={() => userToDelete && deleteUser.mutate(userToDelete.id)}
            >
                Are you sure you want to delete the user "{userToDelete?.login}"? Their VPN connections will be removed.
            </ConfirmModal>
            {updateUser.isError ? <ErrorAlert message={"Could not update user: " + getErrorMessage(updateUser.error)} /> : null}
            {deleteUser.isError ? <ErrorAlert message={"Could not delete user: " + getErrorMessage(deleteUser.error)} /> : null}
            <Modal opened={opened} onClose={closePasswordModal} title="Change Password">
                
                {passwordUpdated ?
                    <Container my={40}>
                        <Alert variant="light" color="green" title="Update!" icon={alertIcon}>Password Updated!</Alert>
                        <Space h="md" />
                        <Button onClick={closePasswordModal}>Close</Button>
                    </Container>
                :
                    <Container my={40}>
                        <ErrorAlert message={passwordError !== "" ? passwordError : changePassword.isError ? getErrorMessage(changePassword.error) : ""} />
                        <PasswordInput placeholder="New Password" id="your-password" value={newPassword} onChange={(event) => setNewPassword(event.currentTarget.value)} /><Space h="md" />
                        <Button onClick={submitPassword} loading={changePassword.isPending}>Change Password</Button>
                    </Container>
                }
                
            </Modal>
            <Table.ScrollContainer minWidth={900}>
                <Table verticalSpacing="sm">
                <Table.Thead>
                    <Table.Tr>
                    <Table.Th>Login</Table.Th>
                    <Table.Th>Role</Table.Th>
                    <Table.Th>Type</Table.Th>
                    <Table.Th>Status</Table.Th>
                    <Table.Th>Last Web Login</Table.Th>
                    <Table.Th />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>{rows}</Table.Tbody>
                </Table>
            </Table.ScrollContainer>
          </>
        );
      }
      