import { useEffect, useState } from "react";
import { AppSettings } from "../Constants/Constants";
import axios from "axios";
import { useMutation } from "@tanstack/react-query";
import { useMatch } from "react-router-dom";

type AuthParams = {
    loginType: string;
    loginID: string;
  };

export function OIDCAndSAMLRedirect() {
    const [redirectError, setRedirectError] = useState<string>("")
    const loginMatch = useMatch("/login/:loginType/:loginID");
    const redirect = useMutation({
        mutationFn: (authParams:AuthParams) => {
          return axios.get(AppSettings.url + '/authmethods/'+authParams.loginType+'/' + authParams.loginID)
        },
        onSuccess: (response) => {
          const data = response.data as OIDCProvider
          const redirectURI = data.redirectURI || ""
          if(redirectURI === "") {
            setRedirectError("Could not redirect at this time: redirectURI is empty")
          } else {
            window.location.href = redirectURI
          }
          
        },
        onError: (error) => {
            setRedirectError("Could not redirect at this time: "+ error.message)
        }
      })

    useEffect(() => {
        if(loginMatch && loginMatch.params.loginType !== undefined && loginMatch.params.loginID !== undefined) {
            redirect.mutate({loginType:loginMatch.params.loginType, loginID: loginMatch.params.loginID})
        }  
    }, []);
    if(redirectError !== "") return <p>RedirectError</p>
    return (<></>)
}