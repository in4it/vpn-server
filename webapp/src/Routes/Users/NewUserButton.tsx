
import { Button } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { AppSettings } from "../../Constants/Constants";
import { useAuthContext } from "../../Auth/Auth";

type Props = {
    setShowNewUser: (newType: boolean) => void;
    localAuthDisabled: boolean;
  };

export function NewUserButton({setShowNewUser, localAuthDisabled}:Props) {
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
    if (error) return 'cannot retrieve licensed users'
    if (isPending) return <Button disabled>Loading...</Button>
    if(localAuthDisabled) {
        return <Button disabled>New Local User</Button>
    } else if(data.currentUserCount >= data.licenseUserCount) {
        return <Button disabled>New Local User (license user count reached)</Button>
    } else {
        return <Button onClick={() => setShowNewUser(true)}>New Local User</Button>
    }
    
}