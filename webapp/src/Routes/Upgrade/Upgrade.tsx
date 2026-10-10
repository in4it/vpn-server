import { Text, Card, Container, Title, Space, Button, Alert } from '@mantine/core';
import { useAuthContext } from '../../Auth/Auth';
import classes from './Upgrade.module.css';

import { AppSettings } from '../../Constants/Constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';
import { TbInfoCircle } from "react-icons/tb";
import { useEffect, useState } from 'react';
import { fetchJSON } from '../../Auth/fetchJSON';

type UpgradeResponse = {
  currentVersion: string;
  newVersionAvailable: boolean;
  newVersion: string;
}

export function Upgrade() {
  const {authInfo} = useAuthContext()
  const queryClient = useQueryClient()
  const [message, setMessage] = useState("")
  const [upgradeError, setUpgradeError] = useState("")
  const [upgradeStatus, setUpgradeStatus] = useState("")
  const [upgradeCheckCounter, setUpgradeCheckCounter] = useState(1)
  // the version that was running when the upgrade was started
  const [versionBeforeUpgrade, setVersionBeforeUpgrade] = useState("")
  const { isPending, error, data } = useQuery<UpgradeResponse>({
    queryKey: ['upgrade'],
    queryFn: async () => {
      if(upgradeStatus !== "upgrading") {
        return fetchJSON<UpgradeResponse>(AppSettings.url + '/upgrade', authInfo.token)
      }
      // While upgrading, the server restarts: errors (502, connection refused)
      // and answers from the old server are expected. Keep the last known data
      // and only mark the upgrade completed once the new version answers.
      const previous = queryClient.getQueryData<UpgradeResponse>(['upgrade']) as UpgradeResponse
      let current: UpgradeResponse
      try {
        current = await fetchJSON<UpgradeResponse>(AppSettings.url + '/upgrade', authInfo.token)
      } catch {
        return previous
      }
      if(current.currentVersion === undefined || current.currentVersion === versionBeforeUpgrade) {
        return previous
      }
      setUpgradeStatus("completed")
      setMessage("")
      queryClient.invalidateQueries({ queryKey: ['version'] })
      return current
    },
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
        setVersionBeforeUpgrade(data?.currentVersion ?? "")
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

  if (error) return 'cannot retrieve upgrade information: ' + error.message
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