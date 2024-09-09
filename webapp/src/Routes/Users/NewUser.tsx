import { Button, Container, Paper, TextInput, Title, Alert, Select, PasswordInput, InputWrapper } from "@mantine/core";
import { useState } from "react";
import classes from './NewUser.module.css';
import { useMutation } from "@tanstack/react-query";
import axios, { AxiosError } from "axios";
import { AppSettings } from "../../Constants/Constants";
import { TbInfoCircle } from "react-icons/tb";
import { useAuthContext } from "../../Auth/Auth";
import { useForm } from "@mantine/form";

  type UserError = {
    error: string;
  }

  type Props = {
    setShowNewUser: (newType: boolean) => void;
  };


export function NewUser({setShowNewUser} :Props) {
    const {authInfo} = useAuthContext()
    const [userError, setUserError] = useState<string>("")
    const userMutation = useMutation({
        mutationFn: (user: User) => {
          return axios.post(AppSettings.url + '/users', user, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            setShowNewUser(false)
        },
        onError: (error:AxiosError) => {
            const errorMessage = error.response?.data as UserError
            if(errorMessage?.error === undefined) {
                setUserError("Error: "+ error.message)
            } else {
                setUserError("Error: "+ errorMessage.error)
            }
        }
    })
    const alertIcon = <TbInfoCircle />;
    const rolesData = ["user", "admin"];
    const form = useForm({
        mode: 'uncontrolled',
        initialValues: {
            id: "",
            login: "",
            password: "",
            role: "user",
            oidcID: "",
            samlID: "",
            provisioned: false,
            suspended: false,
            lastTokenRenewal: "",
            lastLogin: "",
            connectionsDisabledOnAuthFailure: false,
        },
        validate: {
            login: (value) => (/^[a-z0-9]+$/.test(value) ? null : 'Invalid login (only alphanumeric characters allowed)'),
        },
      });
    return (
        <Container size={800} my={40}>
          <Title ta="center" className={classes.title}>
            New Local User
          </Title>
          <Paper withBorder shadow="md" p={30} mt={30} radius="md">
            <form onSubmit={form.onSubmit((values) => userMutation.mutate(values))}>
            {userError != "" ? <Alert variant="light" color="red" title="Error" icon={alertIcon}>{userError}</Alert> : null}

            <p>Don't add any OIDC users. They'll be automatically added once a user logs in for the first time.</p>
            <InputWrapper
                id="login-demo"
                required
                label="Login"
                description="Login can only contain alphanumeric characters."
                
            >
            <TextInput placeholder="Login" required key={form.key('login')} {...form.getInputProps('login')} style={{ marginTop: 5 }} />
            </InputWrapper>
            <PasswordInput mt="md" label="Password" placeholder="Password" required key={form.key('password')} {...form.getInputProps('password')} />
            <Select
               mt="md"
               label="Role"
               data={rolesData}
               defaultValue={"user"}
               allowDeselect={false}
               required
               key={form.key('role')}
               {...form.getInputProps('role')}
            />
            <Button mt="md" type="submit">Save</Button>
            </form>
          </Paper>
          
        </Container>

    )
}