
import { Container, TextInput, Alert, InputWrapper, Button, Space, UnstyledButton, Checkbox, Text } from "@mantine/core";
import { useEffect, useState } from "react";
import classes from './Setup.module.css';
import { IconInfoCircle } from "@tabler/icons-react";
import { AppSettings } from "../../Constants/Constants";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { useForm } from '@mantine/form';
import axios, { AxiosError } from "axios";


type VPNSetupError = {
    error: string;
}

type VPNSetupRequest = {
    routes: string;
    vpnEndpoint: string;
    addressRange: string,
    clientAddressPrefix: string,
    port: string,
    externalInterface: string,
    nameservers: string,
    disableNAT: boolean,
};
export function VPNSetup() {
    const [saved, setSaved] = useState(false)
    const [saveError, setSaveError] = useState("")
    const {authInfo} = useAuthContext();
    const queryClient = useQueryClient()
    const { isPending, error, data, isSuccess } = useQuery({
      queryKey: ['vpn-setup'],
      queryFn: () =>
        fetch(AppSettings.url + '/setup/vpn', {
          headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + authInfo.token
          },
        }).then((res) => {
          return res.json()
          }
          
        ),
    })
    const form = useForm({
      mode: 'uncontrolled',
      initialValues: {
        routes: "",
        vpnEndpoint: "",
        addressRange: "",
        clientAddressPrefix: "",
        port: "",
        externalInterface: "",
        nameservers: "",
        disableNAT: false,   
      },
    });
    const setupMutation = useMutation({
      mutationFn: (setupRequest: VPNSetupRequest) => {
        return axios.post(AppSettings.url + '/setup/vpn', setupRequest, {
          headers: {
              "Authorization": "Bearer " + authInfo.token
          },
        })
      },
      onSuccess: () => {
          setSaved(true)
          setSaveError("")
          queryClient.invalidateQueries({ queryKey: ['vpn-setup'] })
          window.scrollTo(0, 0)
      },
      onError: (error:AxiosError) => {
        const errorMessage = error.response?.data as VPNSetupError
        if(errorMessage?.error === undefined) {
            setSaveError("Error: "+ error.message)
        } else {
            setSaveError("Error: "+ errorMessage.error)
        }      
      }
    })

    const alertIcon = <IconInfoCircle />;

    useEffect(() => {
      if (isSuccess) {
        form.setValues({ ...data });
      }
    }, [isSuccess]); 
  

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    return (
        <Container my={40} size="40rem">
            <Alert variant="light" color="blue" title="Note!" icon={alertIcon}>Changes to Address Range, Port, External Interface, or NAT will need a wireguard reload. You can click the "Reload Wireguard" button in the Restart tab after submitting the changes. This will disconnect active VPN clients, and if the Address Range or Port is changed, all clients will need to download a new VPN Config.</Alert>
            {saved && saveError === "" ? <Alert variant="light" color="green" title="Update!" icon={alertIcon} style={{marginTop: 10}}>Settings Saved!</Alert> : null}
            {saveError !== "" ? <Alert variant="light" color="red" title="Error!" icon={alertIcon} style={{marginTop: 10}}>{saveError}</Alert> : null}

            <form onSubmit={form.onSubmit((values: VPNSetupRequest) => setupMutation.mutate(values))}>
                <InputWrapper
                id="input-vpn-endpoint"
                label="VPN Endpoint to use"
                description="VPN clients will have this hostname configured in their configuration file. Usually the same as the VPN Server Hostname in the general tab."
                style={{marginTop: 10}}
                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="hostname"
                key={form.key('vpnEndpoint')}
                {...form.getInputProps('vpnEndpoint')}
                />
                </InputWrapper>

                <InputWrapper
                id="input-route-input"
                label="VPN Client Routes for clients to use"
                description="Network address should be comma separated. Enter '0.0.0.0/0, ::/0' to route all traffic."
                style={{marginTop: 10}}
                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="list of comma separated routes"
                key={form.key('routes')}
                {...form.getInputProps('routes')}
                />
                </InputWrapper>

                <InputWrapper
                id="input-addressrange-input"
                label="Address range"
                description="Should be an address range in the format address/prefix. This is the address range that the VPN will use. It needs to be large enough to contain all IP addresses for every client assigned."
                style={{marginTop: 10}}

                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="1.2.3.4/21"
                key={form.key('addressRange')}
                {...form.getInputProps('addressRange')}
                />
                </InputWrapper>

                <InputWrapper
                id="input-client-address-prefix-input"
                label="Client Address Network Prefix"
                description="Network prefix for the VPN Client to use. /32 means only one IP address for a client."
                style={{marginTop: 10}}
                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="/32"
                key={form.key('clientAddressPrefix')}
                {...form.getInputProps('clientAddressPrefix')}
                />
                </InputWrapper>

                <InputWrapper
                id="input-port-input"
                label="VPN Port"
                description="VPN port to use. 51820 is the default WireGuard® port."
                style={{marginTop: 10}}
                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="51820"
                key={form.key('port')}
                {...form.getInputProps('port')}
                />
                </InputWrapper>

                <InputWrapper
                id="input-external-interface-input"
                label="External Interface"
                description="External Interface on the instance to route external VPN traffic over. Auto-detected by using the interface that has 0.0.0.0/0 route assigned."
                style={{marginTop: 10}}
                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="interface"
                key={form.key('externalInterface')}
                {...form.getInputProps('externalInterface')}
                />
                </InputWrapper>

                <InputWrapper
                id="input-nameservers-input"
                label="Nameservers"
                description="Nameserver IP address to use in the VPN Client. Comma separated if multiple."
                style={{marginTop: 10}}
                >
                <TextInput
                style={{ marginTop: 5 }}
                placeholder="nameserver1, nameserver2"
                key={form.key('nameservers')}
                {...form.getInputProps('nameservers')}
                />
                </InputWrapper>
                <Space h="md" />
                <UnstyledButton className={classes.button} onClick={() => form.setFieldValue("disableNAT", !form.getValues().disableNAT )}>
                    <Checkbox
                    tabIndex={-1}
                    size="md"
                    mr="xl"
                    styles={{ input: { cursor: 'pointer' } }}
                    aria-hidden
                    key={form.key('disableNAT')}
                    {...form.getInputProps('disableNAT', { type: 'checkbox' })}
                    />
                    <div>
                    <Text fw={500} mb={7} lh={1}>
                        Disable NAT
                    </Text>
                    <Text fz="sm" c="dimmed">
                        Packets will be routed to anywhere on the network, using Network Address Translation (NAT). If the VPN clients only need to access the VPN server and not other devices in the network, you can disable NAT.
                    </Text>
                    </div>
                </UnstyledButton>


                <Button type="submit" mt="md">
                Submit
                </Button>
            </form>
        </Container>
    )
}