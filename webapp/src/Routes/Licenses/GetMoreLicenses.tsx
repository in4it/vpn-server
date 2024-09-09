import { Text, Card, Container, Title, Space, Button, Alert } from '@mantine/core';
import { useAuthContext } from '../../Auth/Auth';
import classes from './GetMoreLicenses.module.css';

import { AppSettings } from '../../Constants/Constants';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { TbInfoCircle } from "react-icons/tb";
import { useState } from 'react';

export function GetMoreLicenses() {
  const {authInfo} = useAuthContext()
  const queryClient = useQueryClient()
  const [message, setMessage] = useState("")
  const { isPending, error, data } = useQuery({
    queryKey: ['license'],
    queryFn: () =>
      fetch(AppSettings.url + '/license/get-more', {
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer " + authInfo.token
        },
      }).then((res) => {
        return res.json()
        }
        
      ),
      enabled: authInfo.role === "admin",
  })

  const refreshLicense = () => {
    const now = new Date()
    queryClient.invalidateQueries({ queryKey: ['license'] })
    setMessage("Refreshed licenses ("+ now.toLocaleString() + ")")
  }

  if (error) return 'cannot retrieve licensed users'
  if (isPending) return 'Loading...'

  const alertIcon = <TbInfoCircle />;
  
  return (
    <Container my={40} size="40rem">
    <Title ta="center" className={classes.title}>
      Get More Licenses
    </Title>
    <Space h="md" />
    {message === "" ? null : <Alert variant="light" color="green" title="Licenses" icon={alertIcon} style={{marginBottom: 20}}>{message}</Alert> }


    <Card withBorder radius="md" padding="xl" bg="var(--mantine-color-body)">
      <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
        Current Users: {data.currentUserCount}
      </Text>
      <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
        Current User licenses: {data.licenseUserCount}
      </Text>
      <Text fz="sm" >
        Licenses are refreshed every 24h. You can refresh them manually by clicking the "Refresh License" button below.
      </Text>
      <Text fz="lg" fw={500} style={{marginTop: 20}}>
        <a href={"https://buy.stripe.com/8wM6oA0fW6od1Yk3cc?client_reference_id="+data.key} target="_blank">
          <Button style={{marginRight: 20}}>Buy More Licenses</Button>
        </a>
        <a href="https://billing.stripe.com/p/login/bIYeVq90G0Df0hy6oo" target="_blank">
          <Button style={{marginRight: 20}}>Manage Subscriptions</Button>
        </a>
        <Button variant="default" onClick={() => refreshLicense()}>Refresh License</Button>

      </Text>
    </Card>
    </Container>
  );

}