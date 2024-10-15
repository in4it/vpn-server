import { Container, Button, Alert, Textarea, Space } from "@mantine/core";
import { useEffect, useState } from "react";
import { TbInfoCircle } from "react-icons/tb";
import { AppSettings } from "../../Constants/Constants";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { useForm } from '@mantine/form';
import axios, { AxiosError } from "axios";

type TemplateSetupError = {
    error: string;
}

type TemplateSetupRequest = {
   clientTemplate: string;
   serverTemplate: string;
};
export function TemplateSetup() {
    const [saved, setSaved] = useState(false)
    const [saveError, setSaveError] = useState("")
    const {authInfo} = useAuthContext();
    const queryClient = useQueryClient()
    const { isPending, error, data, isSuccess } = useQuery({
      queryKey: ['templates-setup'],
      queryFn: () =>
        fetch(AppSettings.url + '/vpn/setup/templates', {
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
        clientTemplate: "",
        serverTemplate: "",
      },
    });
    const alertIcon = <TbInfoCircle />;
    const setupMutation = useMutation({
      mutationFn: (setupRequest: TemplateSetupRequest) => {
        return axios.post(AppSettings.url + '/vpn/setup/templates', setupRequest, {
          headers: {
              "Authorization": "Bearer " + authInfo.token
          },
        })
      },
      onSuccess: () => {
          setSaved(true)
          setSaveError("")
          queryClient.invalidateQueries({ queryKey: ['templates-setup'] })
          window.scrollTo(0, 0)
      },
      onError: (error:AxiosError) => {
        const errorMessage = error.response?.data as TemplateSetupError
        if(errorMessage?.error === undefined) {
            setSaveError("Error: "+ error.message)
        } else {
            setSaveError("Error: "+ errorMessage.error)
        }
      }
    })


    useEffect(() => {
      if (isSuccess) {
        form.setValues({ ...data });
      }
    }, [isSuccess]); 
  

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    return (
        <Container my={40} size="80rem">
          <Alert variant="light" color="blue" title="Note" icon={alertIcon}>The template files use the Golang template package (see also <a href="https://pkg.go.dev/text/template" target="_blank">https://pkg.go.dev/text/template</a>).</Alert>
          <Space h="md" />
          {saved && saveError === "" ? <Alert variant="light" color="green" title="Update!" icon={alertIcon}>Settings Saved!</Alert> : null}
          {saveError !== "" ? <Alert variant="light" color="red" title="Error!" icon={alertIcon} style={{marginTop: 10}}>{saveError}</Alert> : null}

          <form onSubmit={form.onSubmit((values: TemplateSetupRequest) => setupMutation.mutate(values))}>
            <Textarea
                label="VPN Client config template"
                key={form.key('clientTemplate')}
                {...form.getInputProps('clientTemplate')}
                autosize
                minRows={2}
                maxRows={20}
                resize="both"
            />
            <Space h="md" />
            <Textarea
                label="VPN Server config template (WireGuard® Configuration Reload in restart tab required to apply)"
                key={form.key('serverTemplate')}
                {...form.getInputProps('serverTemplate')}
                autosize
                minRows={2}
                maxRows={20}
                resize="both"
            />
            <Button type="submit" mt="md">
              Save
            </Button>
            </form>
        </Container>

    )
}