import { Text, Progress, Card, Container, Title, Space, Button } from '@mantine/core';
import { useAuthContext } from '../../Auth/Auth';
import { Link, Navigate } from 'react-router';
import classes from './Home.module.css';

import { AppSettings } from '../../Constants/Constants';
import { useQuery } from '@tanstack/react-query';
import { UpgradeAlert } from './UpgradeAlert';
import { TbPaperBag } from 'react-icons/tb';
import { UserStats } from './UserStats';
import { fetchJSON } from '../../Auth/fetchJSON';

type Props = {
  serverType?: string
};

export function Home({serverType = "vpn"}: Props) {
  const {authInfo} = useAuthContext()
  const { isPending, error, data } = useQuery({
    queryKey: ['license'],
    queryFn: () => fetchJSON(AppSettings.url + '/license', authInfo.token),
      enabled: authInfo.role === "admin",
  })
  if(authInfo.role === "user") {
    // regular users have no status page: send them to their landing page,
    // which differs per server type (/connection only exists on the vpn server)
    return <Navigate to={serverType === "observability" ? "/logs" : "/connection"} />
  }
  if (error) return 'cannot retrieve licensed users: ' + error.message

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