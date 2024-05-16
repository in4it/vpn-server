import {
    useQuery,
  } from '@tanstack/react-query'

import React, { createContext, useContext, useEffect, useState } from 'react';
import { AppSettings } from '../Constants/Constants';
import { AuthBanner } from './AuthBanner';
import { useCookies } from 'react-cookie';
import { useMatch } from 'react-router-dom';
import { OIDCAndSAMLAuth } from './OIDCAndSAMLAuth';
import { OIDCAndSAMLRedirect } from './OIDCAndSAMLRedirect';
type Props = {
  children?: React.ReactNode
};

type AuthInfo = {
  login: string,
  role: string,
  token: string,
  userType: string,
}

type UserInfoResponse = {
  login: string,
  role: string,
  userType: string,
}

interface AuthContext {
  authInfo: AuthInfo;
  setAuthInfo: (authInfo: AuthInfo) => void;
}

const AuthContext = createContext<AuthContext>({
  authInfo: {login: "", role: "", token: "", userType: ""},
  setAuthInfo: () => {},
});

export const useAuthContext = () => {
  return useContext(AuthContext);
}  

export const Auth: React.FC<Props> = ({children}) => {
    const [authenticated, setAuthenticated] = useState<boolean>(false);
    const [authInfo, setAuthInfo] = useState<AuthInfo>({login: "", role: "", token: "", userType: ""});
    const [cookie, setCookie] = useCookies(['token'])
    const callbackMatch = useMatch("/callback/:type/*");
    const loginMatch = useMatch("/login/:type/*");

    const getToken = () => {
      if(authInfo.token === "" && cookie.token?.length > 0) {
        setAuthInfo({ ...authInfo, token: cookie.token })
        return cookie.token
      }
      return authInfo.token
    }
        
    const { isPending, error, data } = useQuery({
      queryKey: ['userinfo', getToken()],
      queryFn: () => 
        fetch(AppSettings.url + '/userinfo', {
          headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + getToken()
          },
        }).then((res) => {
          if(res.status == 200) {
            if(!authenticated) {
              setCookie("token", getToken(), {path: "/"})
              setAuthenticated(true)
            }
          } else {
            if(authenticated) {
              setAuthenticated(false)
            }
          }
          return res.json()
          }
        ),
    })

    useEffect(() => {
      if(data !== undefined) {
        const userInfoResponse = data as UserInfoResponse
        if (userInfoResponse.login !== "") {
          setAuthInfo({...authInfo, login: data.login, role: data.role, userType: data.userType})
        }
      }
    }, [data]);

    if (isPending) return ''
    if (error) return 'A backend error has occurred: ' + error.message

    if(authenticated) {
      return (
        <AuthContext.Provider value={{authInfo, setAuthInfo}}>
          {children}
        </AuthContext.Provider>
      )
    }

    if(callbackMatch !== null) {
      return (
        <AuthContext.Provider value={{authInfo, setAuthInfo}}>
          <OIDCAndSAMLAuth />
        </AuthContext.Provider>
      )
    }

    if(loginMatch !== null) {
      return (
        <AuthContext.Provider value={{authInfo, setAuthInfo}}>
          <OIDCAndSAMLRedirect />
        </AuthContext.Provider>
      )
    }

    return (
      <AuthContext.Provider value={{authInfo, setAuthInfo}}>
          <AuthBanner /> 
      </AuthContext.Provider>
    )
 }