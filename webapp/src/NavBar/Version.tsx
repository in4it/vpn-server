import { useQuery } from "@tanstack/react-query"
import { useAuthContext } from "../Auth/Auth"
import { AppSettings } from "../Constants/Constants"
import { fetchJSON } from "../Auth/fetchJSON"

export function Version() {
    const {authInfo} = useAuthContext()
    const { isPending, error, data } = useQuery({
      queryKey: ['version'],
      queryFn: () =>
        fetchJSON(AppSettings.url + '/vpn/version', authInfo.token),
    })
    if (isPending) return ''
    if (error) return ''
  
    return data.version
}