import { useAuthContext } from '../../Auth/Auth';
import { useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';

type Props = {
    setLicenseUserCount: (newType: number) => void;
};


export function MaxUsers({setLicenseUserCount}:Props) {
  const {authInfo} = useAuthContext()
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
  })
  if (isPending) return '-'
  if (error) return 'cannot retrieve licensed users'
  setLicenseUserCount(data.licenseUserCount)
  return data.licenseUserCount
 }