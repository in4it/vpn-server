import { Button, Container, Paper, TextInput, Title, Alert, InputWrapper } from "@mantine/core";
import { useState } from "react";
import classes from './NewFactor.module.css';
import { useMutation } from "@tanstack/react-query";
import axios, { AxiosError } from "axios";
import { AppSettings } from "../../Constants/Constants";
import { TbInfoCircle } from "react-icons/tb";
import { useAuthContext } from "../../Auth/Auth";
import { useForm } from "@mantine/form";
import { QRCode } from "./QRCode";


type FactorError = {
    error: string;
}

type Factor = {
    name: string;
    secret: string;
    code: string;
    type: string;
}

type Props = {
  setShowNewFactor: (newType: boolean) => void;
  secret: string;
};


export function NewFactor({setShowNewFactor, secret} :Props) {
    const {authInfo} = useAuthContext()
    const [factorError, setFactorError] = useState<string>("")
    const factorMutation = useMutation({
        mutationFn: (factor: Factor) => {
          return axios.post(AppSettings.url + '/profile/factors', factor, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            setShowNewFactor(false)
        },
        onError: (error:AxiosError) => {
            const errorMessage = error.response?.data as FactorError
            if(errorMessage?.error === undefined) {
                setFactorError("Error: "+ error.message)
            } else {
                setFactorError("Error: "+ errorMessage.error)
            }
        }
    })
    const alertIcon = <TbInfoCircle />;
    const form = useForm({
        mode: 'uncontrolled',
        initialValues: {
            name: "",
            secret: secret,
            code: "",
            type: "totp",
        },
        validate: {
            name: (value) => (/^[a-z0-9-]+$/.test(value) ? null : 'Invalid name (only alphanumeric characters, and the dash sign (-) is allowed)'),
            code: (value) => (/^[0-9]+$/.test(value) ? null : 'Invalid code (only numeric values)'),
        },
    });
      
    return (
        <Container size={800} my={40}>
          <Title ta="center" className={classes.title}>
            New Security Factor (MFA)
          </Title>
          <Paper withBorder shadow="md" p={30} mt={30} radius="md">
            <form onSubmit={form.onSubmit((values) => factorMutation.mutate(values))}>
            {factorError != "" ? <Alert variant="light" color="red" title="Error" icon={alertIcon}>{factorError}</Alert> : null}

            <InputWrapper
                id="name"
                required
                label="Name"
                description="Give a unique name to this factor. For example: google-auth, authy, token1."
            >
                <TextInput placeholder="Name" required key={form.key('name')} {...form.getInputProps('name')} style={{ marginTop: 5 }} maxLength={16} />
            </InputWrapper>
            <QRCode value={"otpauth://totp/VPN-Server:"+authInfo.login+"?secret=" + secret + "&issuer=VPN-Server"} />
            <InputWrapper
                id="code"
                required
                label="Code"
                description="Scan the QR code with an authenticator app (Google Auth, Authy, or other), then enter the 6 digit code below."
            >
                <TextInput placeholder="Name" required key={form.key('code')} {...form.getInputProps('code')} style={{ marginTop: 5 }} />
            </InputWrapper>
            <Button mt="md" type="submit">Add</Button>
            </form>
          </Paper>
          
        </Container>

    )
}