import { Container, Button, Alert, Space } from "@mantine/core";
import { useState } from "react";
import { IconInfoCircle } from "@tabler/icons-react";
import { AppSettings } from "../../Constants/Constants";
import { useMutation } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import axios, { AxiosError } from "axios";

type RestartError = {
    error: string;
}

export function Restart() {
    const [saved, setSaved] = useState(false)
    const [saveError, setSaveError] = useState("")
    const {authInfo} = useAuthContext();
    const alertIcon = <IconInfoCircle />;
    const setupMutation = useMutation({
      mutationFn: () => {
        return axios.post(AppSettings.url + '/setup/restart-vpn', {}, {
          headers: {
              "Authorization": "Bearer " + authInfo.token
          },
        })
      },
      onSuccess: () => {
          setSaved(true)
          setSaveError("")
      },
      onError: (error:AxiosError) => {
        const errorMessage = error.response?.data as RestartError
        if(errorMessage?.error === undefined) {
            setSaveError("Error: "+ error.message)
        } else {
            setSaveError("Error: "+ errorMessage.error)
        }
      }
    })
  
    return (
        <Container my={40} size="80rem">
          <Alert variant="light" color="blue" title="Note" icon={alertIcon}>This button will reload the WireGuard® Configuration. VPN Clients will be disconnected during the reload. If the configuration has changed, clients might have to download new configuration files (for example if the port or address range has changed). The VPN Server admin UI will not be restarted.</Alert>
          <Space h="md" />
          {saved && saveError === "" ? <Alert variant="light" color="green" title="Restarted!" icon={alertIcon}>VPN Restarted!</Alert> : null}
          {saveError !== "" ? <Alert variant="light" color="red" title="Error!" icon={alertIcon} style={{marginTop: 10}}>{saveError}</Alert> : null}
            <Button type="submit" mt="md" onClick={() =>  setupMutation.mutate()}>
              Reload WireGuard® VPN
            </Button>
        </Container>

    )
}