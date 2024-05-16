import { useAuthContext } from '../../Auth/Auth';
import { useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';

type Props = {
    setUserCount: (newType: number) => void;
};

export function UserCount({setUserCount}:Props) {
  const {authInfo} = useAuthContext()
  const { isPending, error, data } = useQuery({
    queryKey: ['users'],
    queryFn: () =>
      fetch(AppSettings.url + '/users', {
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
  if (error) return 'cannot user count'
  setUserCount(data.length)
  return data.length
 }