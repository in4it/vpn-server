import { useAuthContext } from '../../Auth/Auth';
import { useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';
import { fetchJSON } from '../../Auth/fetchJSON';

type Props = {
    setLicenseUserCount: (newType: number) => void;
};


export function MaxUsers({setLicenseUserCount}:Props) {
  const {authInfo} = useAuthContext()
  const { isPending, error, data } = useQuery({
    queryKey: ['license'],
    queryFn: () =>
      fetchJSON(AppSettings.url + '/license', authInfo.token),
  })
  if (isPending) return '-'
  if (error) return 'cannot retrieve licensed users'
  setLicenseUserCount(data.licenseUserCount)
  return data.licenseUserCount
 }