import { Anchor, Checkbox, Grid, InputWrapper, PasswordInput, Text, TextInput, UnstyledButton } from "@mantine/core";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AppSettings } from "../../../Constants/Constants";
import { useAuthContext } from "../../../Auth/Auth";
import classes from './Provisioning.module.css';
import axios from "axios";
import { useClipboard } from "@mantine/hooks";
import { useState } from "react";
import { fetchJSON } from "../../../Auth/fetchJSON";
import { getErrorMessage } from "../../../Auth/errorMessage";
import { ConfirmModal } from "../../../ConfirmModal";
import { ErrorAlert } from "../../../ErrorAlert";

type SCIMSetup = {
    enabled: boolean;
    regenerateToken: boolean;
  }

export function Provisioning() {
    const queryClient = useQueryClient()
    const {authInfo} = useAuthContext();
    const clipboard = useClipboard({ timeout: 500 });
    const [confirmRegenerate, setConfirmRegenerate] = useState(false)

    const { isPending, error, data } = useQuery({
        queryKey: ['scim-setup'],
        queryFn: () => fetchJSON(AppSettings.url + '/scim-setup', authInfo.token),
    })

    const scimSetup = useMutation({
        mutationFn: (payload:SCIMSetup) => {
          return axios.post(AppSettings.url + '/scim-setup', payload, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['scim-setup'] })
        },
        onSettled: () => {
            setConfirmRegenerate(false)
        },
      })

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    
    return (
        <>
        <ConfirmModal
            opened={confirmRegenerate}
            title="Regenerate SCIM token"
            confirmLabel="Regenerate"
            loading={scimSetup.isPending}
            onClose={() => setConfirmRegenerate(false)}
            onConfirm={() => scimSetup.mutate({enabled: true, regenerateToken: true})}
        >
            Are you sure you want to regenerate the SCIM token? The current token stops working immediately, and your Identity Management System needs to be updated with the new token.
        </ConfirmModal>
        <ErrorAlert message={scimSetup.isError ? "Could not update SCIM settings: " + getErrorMessage(scimSetup.error) : ""} />
        <Grid>
            <Grid.Col span={6}>
            <UnstyledButton onClick={() => scimSetup.mutate({enabled:!data.enabled, regenerateToken: false})} className={classes.button}>
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
                Enable SCIM v2 endpoint
                </Text>
                <Text fz="sm" c="dimmed">
                SCIM is a System for Cross-domain Identity Management. An external Identity Management System can provision users using SCIM. Calls can be made to create/delete/suspend users. Check your Identity Management System for SCIM support and configuration.
                </Text>
            </div>
            </UnstyledButton>
        </Grid.Col>
        <Grid.Col span={6}>
            {data.enabled ? 
                <>
                    <InputWrapper
                    id="token"
                    label="SCIM v2 Bearer Token"
                    style={{marginTop: 0, paddingTop: 0 }} >
                        <Anchor href="#" onClick={() => clipboard.copy(data.token)} pt={2} fw={500} fz="xs" style={{marginLeft: 10}}>
                            Copy Token
                        </Anchor> - 
                        <Anchor href="#" onClick={(event) => { event.preventDefault(); setConfirmRegenerate(true) }} pt={2} fw={500} fz="xs" style={{marginLeft: 10}}>
                            Regenerate Token
                        </Anchor>
                        <PasswordInput id="token" value={data.token} variant="filled" readOnly={true} />
                    </InputWrapper>
                    <InputWrapper
                    id="token"
                    label="SCIM v2 Base URL"
                    style={{marginTop: 20 }} >
                        <Anchor href="#" onClick={() => clipboard.copy(data.baseURL)} pt={2} fw={500} fz="xs" style={{marginLeft: 10}}>
                            Copy URL
                        </Anchor>
                        <TextInput id="baseURL" value={data.baseURL} variant="filled" readOnly={true} />
                    </InputWrapper>
                </>
                    : null }
        </Grid.Col>
        </Grid>
        </>
    )
}