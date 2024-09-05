import { Button, Container, Paper, TextInput, Title, Text, Center, Box, Group, Select } from "@mantine/core";
import { useState } from "react";
import classes from './NewSAML.module.css';
import { useMutation } from "@tanstack/react-query";
import axios, { AxiosError } from "axios";
import { AppSettings } from "../../../Constants/Constants";
import { TbX } from "react-icons/tb";
import { useAuthContext } from "../../../Auth/Auth";

  type SAMLProviderError = {
    error: string;
  }

  type Props = {
    setShowNewSAMLProvider: (newType: boolean) => void;
  };


export function NewSAML({setShowNewSAMLProvider} :Props) {
    const {authInfo} = useAuthContext()
    const [samlProviderError, setSAMLProviderError] = useState<string>("")
    const [samlProvider, setSAMLProvider] = useState<SAMLProvider>({
        id: "",
        name: "samlProvider",
        metadataURL: "",
        issuer: "",
        audience: "",
        acs: "",
        allowMissingAttributes: false,
    });
    const samlProviderMutation = useMutation({
        mutationFn: (samlProvider: SAMLProvider) => {
          return axios.post(AppSettings.url + '/saml-setup', samlProvider, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            setShowNewSAMLProvider(false)
        },
        onError: (error:AxiosError) => {
            const errorMessage = error.response?.data as SAMLProviderError
            if(errorMessage?.error === undefined) {
                setSAMLProviderError("Error: "+ error.message)
            } else {
                setSAMLProviderError("Error: "+ errorMessage.error)
            }
        }
      })
    return (
        <Container size={520} my={40}>
          <Title ta="center" className={classes.title}>
            New SAML Provider
          </Title>
          <Paper withBorder shadow="md" p={30} mt={30} radius="md">
            {samlProviderError != "" ? 
            <>
            <Text component="div" c="red" mt={5} size="sm">
            <Center inline>
                <TbX size="0.9rem" />
                <Box ml={7}>{samlProviderError}</Box>
            </Center>
            </Text>
            </>
            
            : <></>}
            <TextInput style={{marginTop: 10}} classNames={classes} label="Name" placeholder="Name" required onChange={(event) => setSAMLProvider({...samlProvider, name: event.currentTarget.value})} value={samlProvider.name} />
            <TextInput style={{marginTop: 10}} classNames={classes} label="Metadata URL" placeholder="Metadata URL" required onChange={(event) => setSAMLProvider({...samlProvider, metadataURL: event.currentTarget.value})} value={samlProvider.metadataURL} />
            <Select
               mt="md"
               label="Allow Missing Attributes"
               description="If you're using Onelogin, you might have to turn this on"
               data={["False","True"]}
               defaultValue={"False"}
               allowDeselect={false}
               onChange={(_value, option) => setSAMLProvider({...samlProvider, allowMissingAttributes: option.value === "True" ? true : false})}
               required
            />
            <Group justify="center">
              <Button style={{marginTop: 20}} onClick={() => setShowNewSAMLProvider(false)} variant="default">Back</Button>
              <Button style={{marginTop: 20}} onClick={() => samlProviderMutation.mutate(samlProvider)}>Save</Button>
            </Group>
            
          </Paper>
          
        </Container>

    )
}