import { Text, Checkbox, Container, Title, UnstyledButton, Tooltip, Center, rem, TextInput, Space, Button, Alert, Divider, InputWrapper } from "@mantine/core";
import classes from './Setup.module.css';
import { useEffect, useState } from "react";
import { IconInfoCircle } from "@tabler/icons-react";
import { AppSettings } from "../../Constants/Constants";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { useForm } from '@mantine/form';
import axios, { AxiosError } from "axios";


type SetupRequest = {
  hostname: string;
  enableTLS: boolean;
  redirectToHttps: boolean;
  disableLocalAuth: boolean;
  enableOIDCTokenRenewal: boolean;
  routes: string;
  vpnEndpoint: string;
};
export function Setup() {
    const [saved, setSaved] = useState(false)
    const [saveError, setSaveError] = useState("")
    const {authInfo} = useAuthContext();
    const queryClient = useQueryClient()
    const { isPending, error, data, isSuccess } = useQuery({
      queryKey: ['setup'],
      queryFn: () =>
        fetch(AppSettings.url + '/setup', {
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
        hostname: "",
        enableTLS: false,
        redirectToHttps: false,
        disableLocalAuth: false,
        enableOIDCTokenRenewal: false,
        routes: "",
        vpnEndpoint: "",
      },
    });
    const alertIcon = <IconInfoCircle />;
    const setupMutation = useMutation({
      mutationFn: (setupRequest: SetupRequest) => {
        return axios.post(AppSettings.url + '/setup', setupRequest, {
          headers: {
              "Authorization": "Bearer " + authInfo.token
          },
        })
      },
      onSuccess: () => {
          queryClient.invalidateQueries({ queryKey: ['users'] })
          setSaved(true)
      },
      onError: (error:AxiosError) => {
          setSaveError("Error: "+ error.message)
      }
    })


    useEffect(() => {
      if (isSuccess) {
        form.setValues({ ...data });
      }
    }, [isSuccess]); 
  

    const hostnameTooltip = (
      <Tooltip
        label="The server hostname. This hostname will be use to request Let's encrypt TLS certificates when TLS is enabled"
        position="top-end"
        withArrow
        transitionProps={{ transition: 'pop-bottom-right' }}
      >
        <Text component="div" c="dimmed" style={{ cursor: 'help' }}>
          <Center>
            <IconInfoCircle style={{ width: rem(18), height: rem(18) }} stroke={1.5} />
          </Center>
        </Text>
      </Tooltip>
    );

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    return (
        <Container my={40} size="40rem">
          <Title ta="center" className={classes.title}>
            VPN Server Setup
          </Title>
          <Space h="md" />
          {saved ? <Alert variant="light" color="green" title="Update!" icon={alertIcon}>Settings Saved!</Alert> : null}
          {saveError !== "" ? saveError : null}
          <form onSubmit={form.onSubmit((values: SetupRequest) => setupMutation.mutate(values))}>
          <TextInput
          rightSection={hostnameTooltip}
          label="VPN Server Hostname"
          placeholder="Hostname"
          key={form.key('hostname')}
          {...form.getInputProps('hostname')}
          />
          <Space h="md" />
            <UnstyledButton className={classes.button} onClick={() => form.setFieldValue("enableTLS", !form.getValues().enableTLS )}>
              <Checkbox
                tabIndex={-1}
                size="md"
                mr="xl"
                styles={{ input: { cursor: 'pointer' } }}
                aria-hidden
                key={form.key('enableTLS')}
                {...form.getInputProps('enableTLS', { type: 'checkbox' })}
              />
              <div>
                <Text fw={500} mb={7} lh={1}>
                  Enable TLS (https)
                </Text>
                <Text fz="sm" c="dimmed">
                Enable TLS (https) using Let's Encrypt (recommended)
                </Text>
              </div>
            </UnstyledButton>
            <Space h="md" />
            <UnstyledButton className={classes.button} onClick={() => form.setFieldValue("redirectToHttps", !form.getValues().redirectToHttps )}>
              <Checkbox
                tabIndex={-1}
                size="md"
                mr="xl"
                styles={{ input: { cursor: 'pointer' } }}
                aria-hidden
                key={form.key('redirectToHttps')}
                {...form.getInputProps('redirectToHttps', { type: 'checkbox' })}
              />
              <div>
                <Text fw={500} mb={7} lh={1}>
                  Redirect http to https
                </Text>
                <Text fz="sm" c="dimmed">
                  Redirect http requests to https.
                  Not needed when terminating TLS on an external LoadBalancer.
                  Recommended once TLS is activated and working.
                </Text>
              </div>
            </UnstyledButton>
            <Space h="md" />
            <UnstyledButton className={classes.button} onClick={() => form.setFieldValue("disableLocalAuth", !form.getValues().disableLocalAuth )}>
              <Checkbox
                tabIndex={-1}
                size="md"
                mr="xl"
                styles={{ input: { cursor: 'pointer' } }}
                aria-hidden
                key={form.key('disableLocalAuth')}
                {...form.getInputProps('disableLocalAuth', { type: 'checkbox' })}
              />
              <div>
                <Text fw={500} mb={7} lh={1}>
                  Disable local auth
                </Text>
                <Text fz="sm" c="dimmed">
                  Once an OIDC Connection is setup, you can disable local authentication. Make sure to have assigned a new admin role.
                </Text>
              </div>
            </UnstyledButton>
            <Space h="md" />
            <UnstyledButton className={classes.button} onClick={() => form.setFieldValue("enableOIDCTokenRenewal", !form.getValues().enableOIDCTokenRenewal )}>
              <Checkbox
                tabIndex={-1}
                size="md"
                mr="xl"
                styles={{ input: { cursor: 'pointer' } }}
                aria-hidden
                key={form.key('enableOIDCTokenRenewal')}
                {...form.getInputProps('enableOIDCTokenRenewal', { type: 'checkbox' })}
              />
              <div>
                <Text fw={500} mb={7} lh={1}>
                  Deactivate a user's VPN connection on OIDC token renewal failure
                </Text>
                <Text fz="sm" c="dimmed">
                  OIDC Tokens can be refreshed when expired.
                  The OIDC tokens will be renewed, and on renewal failure, the VPN connection of that user will be disabled until the user logs in again.
                </Text>
                <Text fz="sm" c="dimmed" style={{marginTop: 5}}>Note: Only use this when SCIM provisioning is not possible in your setup. </Text>
              </div>
            </UnstyledButton>
            <Divider my="md" label="Wireguard® Configuration" labelPosition="center" />
            <InputWrapper
              id="input-vpn-endpoint"
              label="VPN Endpoint to use"
              description="Clients will connect to this hostname. Usually the same as the VPN Server Hostname above."
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
              
            >
            <TextInput
              style={{ marginTop: 5 }}
              placeholder="list of comma separated routes"
              key={form.key('routes')}
              {...form.getInputProps('routes')}
              />
            </InputWrapper>
            <Button type="submit" mt="md">
              Submit
            </Button>
            </form>
        </Container>

    )
}