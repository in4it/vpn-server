import { Anchor, Button, Checkbox, Grid, InputWrapper, Text, TextInput, UnstyledButton } from "@mantine/core";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AppSettings } from "../../../Constants/Constants";
import { useAuthContext } from "../../../Auth/Auth";
import classes from './Provisioning.module.css';
import axios from "axios";
import { useState } from "react";

type SAMLSetup = {
    enabled: boolean;
    metadataURL: string;
    regenerateCert: boolean;
  }

export function SAML() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const [metadataURL, setMetadataURL] = useState("")
    //const clipboard = useClipboard({ timeout: 500 });

    const { isPending, error, data } = useQuery({
        queryKey: ['saml-setup'],
        queryFn: () =>
          fetch(AppSettings.url + '/saml-setup', {
            headers: {
              "Content-Type": "application/json",
              "Authorization": "Bearer " + authInfo.token
            },
          }).then((res) => {
            return res.json()
            }
            
          ),
    })

    const samlSetup = useMutation({
        mutationFn: (payload:SAMLSetup) => {
          return axios.post(AppSettings.url + '/saml-setup', payload, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['saml-setup'] })
        }
      })

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    
    return (
        <Grid>
            <Grid.Col span={6}>
            <UnstyledButton onClick={() => samlSetup.mutate({enabled:!data.enabled, metadataURL: "", regenerateCert: false})} className={classes.button}>
            <Checkbox
                checked={data.enabled}
                onChange={() => {}}
                tabIndex={-1}
                size="md"
                mr="xl"
                styles={{ input: { cursor: 'pointer' } }}
                aria-hidden
            />

            <div>
                <Text fw={500} mb={7} lh={1}>
                    Enable SAML
                </Text>
                <Text fz="sm" c="dimmed">
                    Security Assertion Markup Language is an open standard for exchanging authentication and authorization data between parties.
                </Text>
            </div>
            </UnstyledButton>
        </Grid.Col>
        <Grid.Col span={6}>
            {data.enabled ? 
                <>
                    <InputWrapper
                    id="saml-metadata"
                    label="SAML Metadata URL"
                    style={{marginTop: 20 }} >
                        <Anchor href="#" onClick={() => samlSetup.mutate({enabled: data.enabled, metadataURL: "", regenerateCert: true})} pt={2} fw={500} fz="xs" style={{marginLeft: 10}}>
                            Regenerate SAML Certificate
                        </Anchor>
                        <TextInput id="metadataURL" value={data.metadataURL} onChange={(event) => setMetadataURL(event.currentTarget.value)}/>
                    </InputWrapper>
                    <Button onClick={() => samlSetup.mutate({enabled:data.enabled, metadataURL: metadataURL, regenerateCert: false})} style={{marginTop:20}}>Save</Button>
                </>
                    : null }
        </Grid.Col>
        </Grid>
    )
}