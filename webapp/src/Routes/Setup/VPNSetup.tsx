
import { Container, TextInput, Alert, InputWrapper, Button, Space, UnstyledButton, Checkbox, Text, MultiSelect } from "@mantine/core";
import { useEffect, useState } from "react";
import classes from './Setup.module.css';
import { TbInfoCircle } from "react-icons/tb";
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
    enablePacketLogs: boolean,
    packetLogsTypes: string[],
    packetLogsRetention: string,
};
export function VPNSetup() {
    const [saved, setSaved] = useState(false)
    const [saveError, setSaveError] = useState("")
    const {authInfo} = useAuthContext();
    const queryClient = useQueryClient()
    const { isPending, error, data, isSuccess } = useQuery({
      queryKey: ['vpn-setup'],
      queryFn: () =>
        fetch(AppSettings.url + '/vpn/setup/vpn', {
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
        enablePacketLogs: false,
        packetLogsTypes: [],
        packetLogsRetention: "",
      },
    });
    const setupMutation = useMutation({
      mutationFn: (setupRequest: VPNSetupRequest) => {
        return axios.post(AppSettings.url + '/vpn/setup/vpn', setupRequest, {
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

    const alertIcon = <TbInfoCircle />;

    useEffect(() => {
      if (isSuccess) {
        form.setValues({ ...data });
      }
    }, [isSuccess]); 
  

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    return (
        <Container my={40} size="40rem">
            <Alert variant="light" color="blue" title="Note!" icon={alertIcon}>Changes to Address Range, Port, External Interface, or NAT will need a wireguard reload. You can click the "Reload WireGuard" button in the Restart tab after submitting the changes. This will disconnect active VPN clients, and if the Address Range or Port is changed, all clients will need to download a new VPN Config.</Alert>
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
                        Packets will be routed to anywhere on the network, using Network Address Translation (NAT). If the VPN clients only need to access the VPN server and no other devices in the network, you can disable NAT.
                    </Text>
                    </div>
                </UnstyledButton>
                <Space h="md" />
                <UnstyledButton className={classes.button} onClick={() => form.setFieldValue("enablePacketLogs", !form.getValues().enablePacketLogs )}>
                    <Checkbox
                    tabIndex={-1}
                    size="md"
                    mr="xl"
                    styles={{ input: { cursor: 'pointer' } }}
                    aria-hidden
                    key={form.key('enablePacketLogs')}
                    {...form.getInputProps('enablePacketLogs', { type: 'checkbox' })}
                    />
                    <div>
                    <Text fw={500} mb={7} lh={1}>
                        Enable IP Packet logging
                    </Text>
                    <Text fz="sm" c="dimmed">
                        Metadata of IP packets passing the VPN can be logged and displayed. Useful if you want to see TCP connection requests, DNS requests, or http/https requests passing the VPN. Can generate a lot of logging data when all traffic is routed over the VPN (0.0.0.0/0 route), or when DNS requests are being logged.
                    </Text>
                    </div>
                </UnstyledButton>
                {form.getValues().enablePacketLogs ?
                    <>
                    <InputWrapper
                    id="input-packetlogger-type-input"
                    label="Select types of packets to log"
                    description="Select the type of packets that need to be logged. Note: Caution! By default all DNS requests are tunneled over the VPN. Enabling DNS will generate a lot of log data!"
                    style={{marginTop: 10}}
                    >
                      <MultiSelect
                      style={{marginTop: 10}}
                      searchable
                      hidePickedOptions
                      comboboxProps={{ offset: 0 }}
                      data={[
                        { value: 'dns', label: 'DNS' },
                        { value: 'http+https', label: 'HTTP/HTTPS' },
                        { value: 'tcp', label: 'New TCP Connections (SYN)' },
                      ]}
                      {...form.getInputProps('packetLogsTypes')}
                      />
                    </InputWrapper>
                    <InputWrapper
                      id="input-packetlogs-retention"
                      label="Log Retention"
                      description="How many days should packet logfiles be kept, in days. Default is 7 days."
                      style={{marginTop: 10}}
                      >
                      <TextInput
                      style={{ marginTop: 5 }}
                      placeholder="7"
                      key={form.key('packetLogsRetention')}
                      {...form.getInputProps('packetLogsRetention')}
                      />
                      </InputWrapper>
                    </>
                : null}
                <Space h="md" />
                <Button type="submit" mt="md">
                Submit
                </Button>
            </form>
        </Container>
    )
}