import { Text, Progress, Card, Container, Title, Space, Button } from '@mantine/core';
import { useAuthContext } from '../../Auth/Auth';
import { Link, Navigate } from 'react-router-dom';
import classes from './Home.module.css';

import { AppSettings } from '../../Constants/Constants';
import { useQuery } from '@tanstack/react-query';
import { UpgradeAlert } from './UpgradeAlert';
import { TbPaperBag } from 'react-icons/tb';
import { UserStats } from './UserStats';

export function Home() {
  const {authInfo} = useAuthContext()
  if(authInfo.role === "user") {
    return <Navigate to={"/connection"} />
  }
  const { isPending, error, data } = useQuery({
    queryKey: ['license'],
    queryFn: () =>
      fetch(AppSettings.url + '/license', {
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
  if (error) return 'cannot retrieve licensed users'

  return (
    <Container my={45} size="45rem">
    <Title ta="center" className={classes.title}>
      VPN Status
    </Title>
    <Space h="md" />

    <UpgradeAlert />

    <Card withBorder radius="md" padding="xl" bg="var(--mantine-color-body)">
      <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
        Active Users / Licensed Users
      </Text>
      <Text fz="lg" fw={500}>
        {isPending ? "-" : data.currentUserCount + " / " + data.licenseUserCount}
      </Text>
      <Progress value={isPending ? 0 : data.currentUserCount / data.licenseUserCount * 100} mt="md" size="lg" radius="xl" />
      <Text>
     
      </Text>
      {isPending || data.cloudType === "aws-marketplace" || data.cloudType === "azure" ? null : 
      <Card.Section inheritPadding mt="sm" pb="md">
          <Link to="/licenses">
            <Button leftSection={<TbPaperBag size={14} />} fz="sm" mt="md" radius="md" variant="default" size="sm">
              Get more licenses
            </Button>
          </Link>
      </Card.Section>
      }
    </Card>
    <UserStats />
    </Container>
  );
}