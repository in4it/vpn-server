import { Text, Title, Button, PasswordInput } from '@mantine/core';
import classes from './SetupBanner.module.css';
import {useState} from 'react';
import axios from 'axios';
import { AppSettings } from '../Constants/Constants';
import {
  useMutation,
  useQueryClient,
} from '@tanstack/react-query'

type Props = {
    onChangeStep: (newType: number) => void;
    secret: string
  };

export function SetAdminPassword({onChangeStep, secret}: Props) {
    const queryClient = useQueryClient()
    const [password, setPassword] = useState<string>("");
    const [password2, setPassword2] = useState<string>("");
    const [passwordError, setPasswordError] = useState<string>("");
    const [password2Error, setPassword2Error] = useState<string>("");
    const passwordMutation = useMutation({
    mutationFn: (newPassword: string) => {
      return axios.post(AppSettings.url + '/context', {secret: secret, adminPassword: newPassword, hostname: window.location.host, protocol: window.location.protocol})
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['context'] })
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
    }
    passwordMutation.mutate(password)
  }
  return (
    <div className={classes.wrapper}>
      <div className={classes.body}>
        <Title className={classes.title}>Set admin Password...</Title>
        <Text fw={500} fz="lg" mb={5}>
          Set a password for the admin user. At the next screen you'll be able to login with the username "admin" and the password you'll set now.
        </Text>
        {passwordMutation.isPending ? (
          <div>Setting Password...</div>
        ) : (
          <div>
            <Text component="label" htmlFor="your-password" size="sm" fw={500}>
            Your password
            </Text>
            <PasswordInput placeholder="Your password" id="your-password-1"
                onChange={(event) => setPassword(event.currentTarget.value)}
                value={password}
                error={passwordError}
                />
            <Text component="label" htmlFor="your-password" size="sm" fw={500}>
            Repeat password
            </Text>
            <PasswordInput
                placeholder="Repeat your password"
                id="your-password-2"
                onChange={(event) => setPassword2(event.currentTarget.value)}
                value={password2}
                error={password2Error} 
                />
            <br />
            <Button onClick={() => changePassword()}>Set Admin Password</Button>
          </div>
        )}
      </div>
    </div>
  );
}