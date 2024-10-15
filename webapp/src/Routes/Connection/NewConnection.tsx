import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { AppSettings } from "../../Constants/Constants"
import axios, { AxiosError } from "axios"
import { useAuthContext } from "../../Auth/Auth";
import { useState } from "react";
import { Alert, Button } from "@mantine/core";
import { TbInfoCircle } from "react-icons/tb";

export function NewConnection() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const [newConnectionError, setError] = useState<string>("")
    const alertIcon = <TbInfoCircle />
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
        },
        onError: (error:AxiosError) => {
            const errorMessage = error.response?.data as GenericErrorResponse
            if(errorMessage?.error === undefined) {
                setError("Error: "+ error.message)
            } else {
                setError("Error: "+ errorMessage.error)
            }
        }
    })
    const { isPending, error, data } = useQuery({
        queryKey: ['connectionlicense'],
        queryFn: () =>
          fetch(AppSettings.url + '/vpn/connectionlicense', {
            headers: {
              "Content-Type": "application/json",
              "Authorization": "Bearer " + authInfo.token
            },
          }).then((res) => {
            return res.json()
            }
            
          ),
      })
    if (isPending) return <Button disabled>Loading...</Button>
    if (newConnectionError != "") {
       return <Alert variant="light" color="red" title="Error" icon={alertIcon}>{newConnectionError}</Alert>
    }
    if (error) {
        return <Alert variant="light" color="red" title="Error" icon={alertIcon}>{error.message}</Alert>
    }
    if (isPending) {
        return <Button disabled>Loading...</Button>
    }
    if(data.licenseUserCount <= 3 && data.connectionCount >= 1) {
        return <Button disabled>Only 1 connection is allowed in the free plan</Button>
    }
    return (
        <Button onClick={() => newConnection.mutate()}>New VPN Connection</Button>
    )
}