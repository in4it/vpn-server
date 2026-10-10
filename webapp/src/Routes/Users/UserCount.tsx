import { useAuthContext } from '../../Auth/Auth';
import { useQuery } from '@tanstack/react-query';
import { AppSettings } from '../../Constants/Constants';
import { fetchJSON } from '../../Auth/fetchJSON';

type Props = {
    setUserCount: (newType: number) => void;
};

export function UserCount({setUserCount}:Props) {
  const {authInfo} = useAuthContext()
  const { isPending, error, data } = useQuery({
    queryKey: ['users'],
    queryFn: () =>
      fetchJSON(AppSettings.url + '/users', authInfo.token),
  })
  if (isPending) return '-'
  if (error) return 'cannot retrieve user count'
  setUserCount(data.length)
  return data.length
 }