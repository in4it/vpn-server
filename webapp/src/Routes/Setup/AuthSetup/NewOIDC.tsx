import { Button, Container, Paper, TextInput, Title, Text, Center, Box, InputWrapper, Group } from "@mantine/core";
import { useState } from "react";
import classes from './NewOIDC.module.css';
import { useMutation } from "@tanstack/react-query";
import axios, { AxiosError } from "axios";
import { AppSettings } from "../../../Constants/Constants";
import { TbX } from "react-icons/tb";
import { useAuthContext } from "../../../Auth/Auth";

  type OIDCProviderError = {
    error: string;
  }

  type Props = {
    setShowNewOIDCProvider: (newType: boolean) => void;
  };


export function NewOIDC({setShowNewOIDCProvider} :Props) {
    const {authInfo} = useAuthContext()
    const [oidcProviderError, setOIDCProviderError] = useState<string>("")
    const [oidcProvider, setOIDCProvider] = useState<OIDCProvider>({
        id: "",
        clientId: "",
        clientSecret: "",
        scope: "openid email profile offline_access",
        discoveryURI: "",
        redirectURI: "",
        loginURL: "",
        name: "myprovider",
    });
    const oidcProviderMutation = useMutation({
        mutationFn: (oidcProvider: OIDCProvider) => {
          return axios.post(AppSettings.url + '/oidc', oidcProvider, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            setShowNewOIDCProvider(false)
        },
        onError: (error:AxiosError) => {
            const errorMessage = error.response?.data as OIDCProviderError
            if(errorMessage?.error === undefined) {
                setOIDCProviderError("Error: "+ error.message)
            } else {
                setOIDCProviderError("Error: "+ errorMessage.error)
            }
        }
      })
    return (
        <Container size={520} my={40}>
          <Title ta="center" className={classes.title}>
            New OIDC Provider
          </Title>
          <Paper withBorder shadow="md" p={30} mt={30} radius="md">
            {oidcProviderError != "" ? 
            <>
            <Text component="div" c="red" mt={5} size="sm">
            <Center inline>
                <TbX size="0.9rem"/>
                <Box ml={7}>{oidcProviderError}</Box>
            </Center>
            </Text>
            </>
            
            : <></>}
            <TextInput style={{marginTop: 10}} classNames={classes} label="Name" placeholder="Name" required onChange={(event) => setOIDCProvider({...oidcProvider, name: event.currentTarget.value})} value={oidcProvider.name} />
            <TextInput style={{marginTop: 10}} classNames={classes} label="Client ID" placeholder="Client ID" required onChange={(event) => setOIDCProvider({...oidcProvider, clientId: event.currentTarget.value})} value={oidcProvider.clientId} />
            <TextInput style={{marginTop: 10}} classNames={classes} label="Client Secret" placeholder="Client Secret" required onChange={(event) => setOIDCProvider({...oidcProvider, clientSecret: event.currentTarget.value})} value={oidcProvider.clientSecret} />
            <InputWrapper
                style={{marginTop: 10}} 
                id="scope"
                required
                label="Scope"
                description="hint: for Onelogin, you might need to remove offline_access. It'll still return the refresh token, even without it."
            >
              <TextInput style={{marginTop: 5}} classNames={classes} placeholder="scope" required onChange={(event) => setOIDCProvider({...oidcProvider, scope: event.currentTarget.value})} value={oidcProvider.scope} />
            </InputWrapper>
            <TextInput style={{marginTop: 10}} classNames={classes} label="Discovery URI" placeholder="discoveryURI" required onChange={(event) => setOIDCProvider({...oidcProvider, discoveryURI: event.currentTarget.value})} value={oidcProvider.discoveryURI} />
            <Group justify="center">
              <Button style={{marginTop: 20}} onClick={() => setShowNewOIDCProvider(false)} variant="default">Back</Button>
              <Button style={{marginTop: 20}} onClick={() => oidcProviderMutation.mutate(oidcProvider)}>Save</Button>
            </Group>
            
          </Paper>
          
        </Container>

    )
}