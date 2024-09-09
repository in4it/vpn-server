import { Text, Card, Container, Title, Space, Button, Alert } from '@mantine/core';
import { useAuthContext } from '../../Auth/Auth';
import classes from './Upgrade.module.css';

import { AppSettings } from '../../Constants/Constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';
import { TbInfoCircle } from "react-icons/tb";
import { useEffect, useState } from 'react';

export function Upgrade() {
  const {authInfo} = useAuthContext()
  const queryClient = useQueryClient()
  const [message, setMessage] = useState("")
  const [upgradeError, setUpgradeError] = useState("")
  const [upgradeStatus, setUpgradeStatus] = useState("")
  const [upgradeCheckCounter, setUpgradeCheckCounter] = useState(1)
  const { isPending, error, data } = useQuery({
    queryKey: ['upgrade'],
    queryFn: () =>
      fetch(AppSettings.url + '/upgrade', {
        headers: {
          "Content-Type": "application/json",
          "Authorization": "Bearer " + authInfo.token
        },
      }).then((res) => {
        if(upgradeStatus === "upgrading") {
          setUpgradeStatus("completed")
          setMessage("")
          queryClient.invalidateQueries({ queryKey: ['version'] })
        }
        return res.json()
        }
        
      ),
  })
  const upgradeMutation = useMutation({
    mutationFn: () => {
      return axios.post(AppSettings.url + '/upgrade', {}, {
        headers: {
            "Authorization": "Bearer " + authInfo.token
        },
      })
    },
    onSuccess: () => {
        setUpgradeStatus("upgrading")
    },
    onError: (error:AxiosError) => {
        setUpgradeError("Error: "+ error.message)
    }
  })

  // upgrade in progress tracker (every 20 seconds)
  useEffect(() => {
      if(upgradeStatus === "upgrading") {
        if(upgradeCheckCounter === 20) {
          setUpgradeStatus("failed")
          setUpgradeError("Upgrade Failed. Check the logs on the instance")
          setUpgradeCheckCounter(0)
        } else {
          setMessage("Upgrade in progress. Waiting for new version to become available ("+upgradeCheckCounter+"/20)...")
          const interval = setInterval(() => {
          setUpgradeCheckCounter(upgradeCheckCounter=>upgradeCheckCounter+1)
          setMessage("Upgrade in progress. Waiting for new version to become available ("+upgradeCheckCounter+"/20)...")
          queryClient.invalidateQueries({ queryKey: ['upgrade'] })  
          }, 20000);
          return () => {
            clearInterval(interval);
          };
        }
      }
  }, [upgradeStatus,upgradeCheckCounter]);

  if (error) return 'cannot retrieve licensed users'
  if (isPending) return 'Loading...'

  const alertIcon = <TbInfoCircle />;
  
  return (
    <Container my={40} size="40rem">
    <Title ta="center" className={classes.title}>
      Upgrade VPN Server
    </Title>
    <Space h="md" />
    {upgradeError === "" ? null : <Alert variant="light" color="red" title="Upgrade Error" icon={alertIcon} style={{marginBottom: 20}}>{upgradeError}</Alert> }
    {message === "" ? null : <Alert variant="light" color="yellow" title="Upgrade In Progress" icon={alertIcon} style={{marginBottom: 20}}>{message}</Alert> }
    {upgradeStatus === "completed" ? <Alert variant="light" color="green" title="Upgrade Completed" icon={alertIcon} style={{marginBottom: 20}}>Upgrade completed</Alert> : null }


    <Card withBorder radius="md" padding="xl" bg="var(--mantine-color-body)">
      <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
        Current Version: {data.currentVersion}
      </Text>
      <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
        {data.newVersionAvailable ?
            <>New Version available: {data.newVersion}</>
        :
            <>No New version available</>
        }
        
      </Text>
      {data.newVersionAvailable ?
      <Text fz="lg" fw={500} style={{marginTop: 20}}>
        <Button onClick={() => upgradeMutation.mutate()} disabled={upgradeStatus === "upgrading"}>Upgrade</Button>
      </Text>
      : null }
    </Card>
    </Container>
  );

}