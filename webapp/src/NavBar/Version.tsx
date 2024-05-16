import { useQuery } from "@tanstack/react-query"
import { useAuthContext } from "../Auth/Auth"
import { AppSettings } from "../Constants/Constants"

export function Version() {
    const {authInfo} = useAuthContext()
    const { isPending, error, data } = useQuery({
      queryKey: ['version'],
      queryFn: () =>
        fetch(AppSettings.url + '/version', {
          headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + authInfo.token
          },
        }).then((res) => {
          return res.json()
          }
          
        ),
    })
    if (isPending) return ''
    if (error) return ''
  
    return data.version
}