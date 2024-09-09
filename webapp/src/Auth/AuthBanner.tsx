import {useState} from 'react';
import {
    TextInput,
    PasswordInput,
    Paper,
    Title,
    Container,
    Button,
    Group,
    Divider,
    Text,
    Alert,
  } from '@mantine/core';
import classes from './AuthBanner.module.css';
import { useMutation, useQuery } from '@tanstack/react-query';
import axios from 'axios';
import { AppSettings } from '../Constants/Constants';
import { useAuthContext } from './Auth';
import { AuthError } from './AuthError';
import { MFAInput } from './MFAInput';
import { TbInfoCircle } from "react-icons/tb";


type LoginResponse = {
  authenticated: string;
  token: string;
  mfaRequired: boolean;
  factors: Array<string>;
};

type OIDCProvider = {
  id: string;
  name: string;
  redirectURI: string | null;
};

type AuthMethods = {
  localAuthDisabled: boolean;
  oidcProviders: [OIDCProvider];
}

type FactorResponse = {
  name: string,
  code: string,
}
type LoginPassword = {
  login: string;
  password: string;
  factorResponse: FactorResponse;
};


export function AuthBanner() {
    const [login, setLogin] = useState<string>("");
    const [password, setPassword] = useState<string>("");
    const [authError, setAuthError] = useState<string>("");
    const {authInfo, setAuthInfo} = useAuthContext();
    const [oidcRedirectError, setOidcRedirectError] = useState<string>("")
    const [showMFAFactors, setShowMFAFactors] = useState<Array<string>>([])
    const [factorResponse, setFactorResponse] = useState<FactorResponse>({name: "", code: ""})
    const { error, isPending, data } = useQuery<any,any,AuthMethods>({
      queryKey: ['authmethods'],
      queryFn: () =>
        fetch(AppSettings.url + '/authmethods')
        .then((res) =>
          res.json(),
        )
    })
    
    const authenticate = useMutation({
        mutationFn: (loginPassword:LoginPassword) => {
          setAuthError("")
          return axios.post(AppSettings.url + '/auth', loginPassword)
        },
        onSuccess: (response) => {
          const data = response.data as LoginResponse
          if(data.mfaRequired) {
            setShowMFAFactors(data.factors)
          } else {
            setAuthInfo({...authInfo, token: data.token})
          }
        },
        onError: (error) => {
          if(error.message.includes("status code 401")) {
            setAuthError("Invalid credentials")
          } else if(error.message.includes("status code 429")) {
              setAuthError("too many attempts. Try again later")
          } else {
            setAuthError("Error: "+ error.message)
          }
        }
      })
     const oidcRedirect = useMutation({
        mutationFn: (id:string) => {
          return axios.get(AppSettings.url + '/authmethods/oidc/' + id)
        },
        onSuccess: (response) => {
          const data = response.data as OIDCProvider
          const redirectURI = data.redirectURI || ""
          if(redirectURI === "") {
            setOidcRedirectError("Could not redirect at this time: redirectURI is empty")
          } else {
            window.location.href = redirectURI
          }
          
        },
        onError: (error) => {
          setOidcRedirectError("Could not redirect at this time: "+ error.message)
        }
      })
    const onClickOidcRedirect = (id:string) => {
        oidcRedirect.mutate(id)
    }

    const captureEnter = (e: React.KeyboardEvent<HTMLDivElement>) => {
      if (e.key === "Enter") {
        authenticate.mutate({login, password, factorResponse})
      }
    }
    const alertIcon = <TbInfoCircle />
    
    if (error) return 'An backend error has occurred: ' + error.message

    const authMethodsButtons = data?.oidcProviders.map((oidcProvider:OIDCProvider) => (
      <Container key={oidcProvider.id}><Button radius="xl" fullWidth={true} key={oidcProvider.id} onClick={() => onClickOidcRedirect
        (oidcProvider.id)}>Login with {oidcProvider.name}</Button></Container>
    ))

    return (
        <Container size={420} my={40}>
          <Title ta="center" className={classes.title}>
            VPN Server
          </Title>
          <AuthError />
          <Paper withBorder shadow="md" p={30} mt={30} radius="md">
            <Group grow mb="md" mt="md">
            {oidcRedirectError === "" ? "" : <p>oidcRedirectError</p>}
            {authMethodsButtons}
            </Group>
            {isPending || data?.localAuthDisabled ? null : 
              <>
                {data?.oidcProviders === undefined || data?.oidcProviders.length < 1 ? null :
                  <Divider label="Or continue with login" labelPosition="center" my="lg" />
                }
                {authError !== "" ? 
                    <Alert variant="light" color="red" title="Error" icon={alertIcon}>{authError}</Alert>
                  :
                    null
                }
                {showMFAFactors.length > 0 ? 
                  <MFAInput factors={showMFAFactors} setFactorResponse={setFactorResponse} captureEnter={captureEnter} /> 
                :
                  <>
                    <TextInput label="Login" placeholder="Your username" required onChange={(event) => setLogin(event.currentTarget.value)} value={login} onKeyDown={(e) => captureEnter(e)} />
                    <PasswordInput label="Password" placeholder="Your password" required mt="md" onChange={(event) => setPassword(event.currentTarget.value)} value={password} onKeyDown={(e) => captureEnter(e)} />
                  </>
                }
                <Button fullWidth mt="xl" onClick={() => authenticate.mutate({login, password, factorResponse})}>
                  Sign in
                </Button>
                <Text size="xs" style={{marginTop: 20}}>By clicking 'Sign in', you accept the <a href="https://in4it.com/vpn-server-terms-conditions/" target="_blank">Terms & Conditions</a></Text>
              </>
            }

          </Paper>
        </Container>
      );
    
}