import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { AppSettings } from "../../Constants/Constants"
import axios from "axios"
import { useAuthContext } from "../../Auth/Auth";
import { Button } from "@mantine/core";
import { fetchJSON } from "../../Auth/fetchJSON";
import { getErrorMessage } from "../../Auth/errorMessage";
import { ErrorAlert } from "../../ErrorAlert";

export function NewConnection() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const newConnection = useMutation({
        mutationFn: () => {
          return axios.post(AppSettings.url + '/vpn/connections', {}, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['connections'] })
            queryClient.invalidateQueries({ queryKey: ['connectionlicense'] })
        },
    })
    const { isPending, error, data } = useQuery({
        queryKey: ['connectionlicense'],
        queryFn: () => fetchJSON(AppSettings.url + '/vpn/connectionlicense', authInfo.token),
      })
    if (isPending) return <Button disabled>Loading...</Button>
    if (error) {
        return <ErrorAlert message={error.message} />
    }
    // the mutation error is reset on the next attempt (mutate clears it), and
    // shown above the button so the user can retry
    const newConnectionError = newConnection.isError ? "Error: " + getErrorMessage(newConnection.error) : ""
    if(data.licenseUserCount <= 3 && data.connectionCount >= 1) {
        return <Button disabled>Only 1 connection is allowed in the free plan</Button>
    }
    return (
        <>
        <ErrorAlert message={newConnectionError} />
        <Button onClick={() => newConnection.mutate()} loading={newConnection.isPending}>New VPN Connection</Button>
        </>
    )
}
