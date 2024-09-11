import { Text, Title, TextInput, Button, Card, Grid, Container, Center, Alert } from '@mantine/core';
import classes from './SetupBanner.module.css';
import {useState} from 'react';
import axios, { AxiosError } from 'axios';
import { AppSettings } from '../Constants/Constants';
import {
  useMutation,
} from '@tanstack/react-query'
import { TbInfoCircle } from 'react-icons/tb';

type Props = {
    onChangeStep: (newType: number) => void;
    onChangeSecrets: (newType: SetupResponse) => void;
    cloudType: string;
};

type SetupResponseError = {
  error: string;
}

const randomHex = (length:number) => {
  const bytes = window.crypto.getRandomValues(new Uint8Array(length))
  var hexstring='', h;
  for(var i=0; i<bytes.length; i++) {
      h=bytes[i].toString(16);
      if(h.length==1) { h='0'+h; }
      hexstring+=h;
  }   
  return "vpnsecret-"+hexstring;
}


export function SetSecret({onChangeStep, onChangeSecrets, cloudType}: Props) {
    const [setupResponse, setSetupResponse] = useState<SetupResponse>({secret: "", tagHash: "", instanceID: ""});
    const [secretError, setSecretError] = useState<string>("");
    const [randomHexValue] = useState(randomHex(16))
    const secretMutation = useMutation({
    mutationFn: (setupResponseParam: SetupResponse) => {
      setSecretError("")
      return axios.post(AppSettings.url + '/context', setupResponseParam)
    },
    onSuccess: (_, setupResponseParam) => {
      onChangeSecrets(setupResponseParam)
      onChangeStep(1)
    },
    onError: (error:AxiosError) => {
      const errorMessage = error.response?.data as SetupResponseError
      if(errorMessage?.error === undefined) {
        setSecretError("Error: "+ error.message)
      } else {
        setSecretError(errorMessage.error)
      }
    },
  })
  const captureEnter = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.key === "Enter") {
      secretMutation.mutate(setupResponse)
    }
  }
  const alertIcon = <TbInfoCircle />
  const hasMoreOptions = cloudType === "aws" || cloudType === "digitalocean" ? true : false
  const colSpanWithSSH = hasMoreOptions ? 3 : 6

  return (
    <Container fluid style={{marginTop: 50}}>
      <Center>
      <Title order={1} style={{marginBottom: 20}}>Start Setup</Title>
      </Center>
      {secretError !== "" ? 
           <Grid>
            <Grid.Col span={3}></Grid.Col>
            <Grid.Col span={6}>
              <Alert variant="light" color="red" title="Error" radius="lg" icon={alertIcon} className={classes.error} style={{marginBottom: 20, paddingLeft: 20, paddingRight:35}}>{secretError}</Alert>
            </Grid.Col>
            </Grid>
        :
          null
      }
    <Grid>
    <Grid.Col span={3}></Grid.Col>
      <Grid.Col span={colSpanWithSSH}>
        <Card withBorder radius="md" p="xl" className={classes.card}>
        <Title order={3} style={{marginBottom: 20}}>{hasMoreOptions ? "Option 1: " : ""}With SSH Access</Title>
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
              onChange={(event) => setSetupResponse({ ...setupResponse, secret: event.currentTarget.value})}
              value={setupResponse.secret}
              onKeyDown={(e) => captureEnter(e)}
            />
            <Button className={classes.control} onClick={() => secretMutation.mutate({ secret: setupResponse.secret, tagHash: "", instanceID: ""})}>Continue</Button>
          </div>
        )}
        </Card>
      </Grid.Col>
      {cloudType === "aws" ? 
        <Grid.Col span={3}>
          <Card withBorder radius="md" p="xl" className={classes.card}>
          <Title order={3} style={{marginBottom: 20}}>{hasMoreOptions ? "Option 2: " : ""}Without SSH Access</Title>

          <Text>
            Enter the EC2 Instance ID of the VPN Server
          </Text>
          {secretMutation.isPending ? (
            <div>Checking Instance ID...</div>
          ) : (
            <div className={classes.controls}>
              <TextInput
                placeholder="i-1234567890abcdef0"
                classNames={{ input: classes.input, root: classes.inputWrapper }}
                onChange={(event) => setSetupResponse({ ...setupResponse, instanceID: event.currentTarget.value})}
                value={setupResponse.instanceID}
                onKeyDown={(e) => captureEnter(e)}
              />
              <Button className={classes.control} onClick={() => secretMutation.mutate({ secret: "", tagHash: "", instanceID: setupResponse.instanceID})}>Check Instance ID</Button>
            </div>
          )}
          </Card>
          </Grid.Col>
        : null }
        {cloudType === "digitalocean" ? 
          <Grid.Col span={3}>
            <Card withBorder radius="md" p="xl" className={classes.card}>
            <Title order={3} style={{marginBottom: 20}}>{hasMoreOptions ? "Option 2: " : ""}Without SSH Access</Title>

            <Text>
              Add the following tag to the droplet by going to the <Text span fw={700}>droplet settings</Text> and opening the <Text span fw={700}>Tags</Text> page. You can remove the tag once the setup is complete.
            </Text>
            {secretMutation.isPending ? (
              <div>Checking tag...</div>
            ) : (
              <div className={classes.controls}>
                <TextInput
                  readOnly={true}
                  classNames={{ input: classes.input, root: classes.inputWrapper }}
                  value={randomHexValue}
                />
                <Button className={classes.control} onClick={() => secretMutation.mutate({ secret: "", tagHash: randomHexValue, instanceID: ""})}>Check tag</Button>
              </div>
            )}
            </Card>
            </Grid.Col>
        : null }
      </Grid>
      </Container>
  );
}