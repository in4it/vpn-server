import { Alert } from '@mantine/core';
import { useAuthContext } from '../../Auth/Auth';
import { Link } from 'react-router-dom';

import { AppSettings } from '../../Constants/Constants';
import { useQuery } from '@tanstack/react-query';
import { TbInfoCircle } from "react-icons/tb";

export function UpgradeAlert() {
  const {authInfo} = useAuthContext()
  const { isPending, error, data } = useQuery({
    queryKey: ['upgrade'],
    queryFn: () =>
      fetch(AppSettings.url + '/upgrade', {
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
  if (error) return ''
  if (isPending) return ''

  const alertIcon = <TbInfoCircle />

  if (!data.newVersionAvailable) return ''

  return (
    <Alert variant="light" color="yellow" title="New Version Available" icon={alertIcon} style={{ marginBottom: 20 }}>A new version is available (current version: {data.currentVersion}, new version: {data.newVersion}). <Link to="/upgrade">Click here to manage upgrades</Link></Alert>
  );
}