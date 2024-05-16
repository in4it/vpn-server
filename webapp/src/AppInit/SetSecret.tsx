import { Text, Title, TextInput, Button } from '@mantine/core';
import classes from './SetupBanner.module.css';
import {useState} from 'react';
import axios from 'axios';
import { AppSettings } from '../Constants/Constants';
import {
  useQueryClient,
  useMutation,
} from '@tanstack/react-query'

type Props = {
    onChangeStep: (newType: number) => void;
    onChangeSecret: (newType: string) => void;
  };

export function SetSecret({onChangeStep, onChangeSecret}: Props) {
    const queryClient = useQueryClient()
    const [secret, setSecret] = useState<string>("");
    const [secretError, setSecretError] = useState<string>("");
    const secretMutation = useMutation({
    mutationFn: (newSecret: string) => {
      setSecretError("")
      return axios.post(AppSettings.url + '/context', {secret: newSecret})
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['context'] })
      onChangeSecret(secret)
      onChangeStep(1)
    },
    onError: (error) => {
      if(error.message.includes("status code 403")) {
        setSecretError("Invalid secret")
      } else {
        setSecretError("Error: "+ error.message)
      }
    }
  })
  return (
    <div className={classes.wrapper}>
      <div className={classes.body}>
        <Title className={classes.title}>Start Setup...</Title>
        <Text fw={500} fz="lg" mb={5}>
          Enter the secret to start the setup.
        </Text>
        <Text fz="sm" c="dimmed">
          To ensure you have administrator access to the instance, enter the secret to start the setup. You can get the secret by logging in to the instance (login is ubuntu), and entering the command:
        </Text>
        <pre>sudo cat /vpn/setup-code.txt</pre>
        <Text fz="sm" c="dimmed">
          Alternatively, if you want to securely enter your admin password over SSH, you can execute the following command on the instance:
        </Text>
        <pre>sudo /vpn/reset-admin-password</pre>
        {secretMutation.isPending ? (
          <div>Checking secret...</div>
        ) : (
          <div className={classes.controls}>
            <TextInput
              placeholder="secret"
              classNames={{ input: classes.input, root: classes.inputWrapper }}
              onChange={(event) => setSecret(event.currentTarget.value)}
              value={secret}
              error={secretError}
            />
            <Button className={classes.control} onClick={() => secretMutation.mutate(secret)}>Continue</Button>
          </div>
        )}
      </div>
    </div>
  );
}