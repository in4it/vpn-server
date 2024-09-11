import { Text, Title, Button, PasswordInput } from '@mantine/core';
import classes from './SetupBanner.module.css';
import {useState} from 'react';
import axios from 'axios';
import { AppSettings } from '../Constants/Constants';
import {
  useMutation,
} from '@tanstack/react-query'

type Props = {
    onChangeStep: (newType: number) => void;
    secrets: SetupResponse
  };

export function SetAdminPassword({onChangeStep, secrets}: Props) {
    const [password, setPassword] = useState<string>("");
    const [password2, setPassword2] = useState<string>("");
    const [passwordError, setPasswordError] = useState<string>("");
    const [password2Error, setPassword2Error] = useState<string>("");
    const passwordMutation = useMutation({
    mutationFn: (newPassword: string) => {
      return axios.post(AppSettings.url + '/context', {...secrets, adminPassword: newPassword, hostname: window.location.host, protocol: window.location.protocol})
    },
    onSuccess: () => {
      onChangeStep(2)
    },
    onError: (error) => {
        setPasswordError("Error: "+ error.message)
    }
  })
  const changePassword = () => {
    setPasswordError("")
    setPassword2Error("")
    if(password !== password2) {
        setPassword2Error("Password doesn't match")
        return
    }
    if(password === "") {
        setPasswordError("admin password cannot be blank")
        return
    }
    passwordMutation.mutate(password)
  }
  const captureEnter = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "Enter") {
      if(password !== "" && password2 !== "") {
        changePassword()
      }
    }
  }
  return (
    <div className={classes.wrapper}>
      <div className={classes.body}>
        <Title className={classes.title}>Set admin Password...</Title>
        <Text fw={500} fz="lg" mb={5}>
          Set a password for the admin user. At the next screen you'll be able to login with the username "admin" and the password you'll set now.
        </Text>
        {passwordMutation.isPending ? (
          <div>Setting Password for user 'admin'...</div>
        ) : (
          <div>
            <Text component="label" htmlFor="your-password" size="sm" fw={500}>
            Your password
            </Text>
            <PasswordInput
                placeholder="Your password for user admin"
                id="your-password-1"
                autoComplete="new-password"
                onChange={(event) => setPassword(event.currentTarget.value)}
                value={password}
                error={passwordError}
                onKeyDown={(e) => captureEnter(e)}
                />
            <Text component="label" htmlFor="your-password" size="sm" fw={500}>
            Repeat password
            </Text>
            <PasswordInput
                placeholder="Repeat your password"
                id="your-password-2"
                autoComplete="new-password"
                onChange={(event) => setPassword2(event.currentTarget.value)}
                value={password2}
                error={password2Error}
                onKeyDown={(e) => captureEnter(e)}
                />
            <br />
            <Button onClick={() => changePassword()}>Set Admin Password</Button>
          </div>
        )}
      </div>
    </div>
  );
}