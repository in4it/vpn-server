import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { useEffect, useState } from "react";
import { AppSettings } from "../Constants/Constants";
import { useLocation, useSearchParams } from "react-router-dom";
import { useAuthContext } from "./Auth";

type TokenRequest = {
   code: string;
   state: string;
   redirectURI: string;
};

type LoginResponse = {
    authenticated: string;
    token: string;
    suspended: boolean;
    noLicense: boolean;
 };
  
export function OIDCAndSAMLAuth() {
    const [authError, setAuthError] = useState<string>("");
    const [searchParams] = useSearchParams();
    const {authInfo, setAuthInfo} = useAuthContext();
    const location = useLocation();
    const [isPending, setIsPending] = useState<boolean>(false)

    const providerID = location.pathname.split("/").pop()
    const authType = location.pathname.split("/").length >= 2 ? location.pathname.split("/").slice(-2, -1)[0] : ""
    const payload = {code: searchParams.get("code") || "", state: searchParams.get("state") || "", redirectURI: location.pathname}

    const getToken = useMutation({
        mutationFn: (tokenRequest:TokenRequest) => {
          return axios.post(AppSettings.url + '/authmethods/' + authType + "/" + providerID, tokenRequest)
        },
        onSuccess: (response) => {
          const data = response.data as LoginResponse
          if (data.suspended) {
            setAuthError("user is suspended")
          } else           if (data.noLicense) {
            setAuthError("user can't be added, because user license limit has been reached")
          } else {
            setAuthInfo({...authInfo, token: data.token})
          }
          
        },
        onError: (error) => {
          if(error.message.includes("status code 401")) {
            setAuthError("Invalid credentials")
          } else {
            setAuthError("Error: "+ error.message)
          }
        }
      })

      useEffect(() => {
        if(!isPending) {
            setIsPending(true)
            getToken.mutate(payload)
        }
      }, []);
      
      
      if (authError !== "") return 'Could not obtain token: ' + authError
      return (
        <></>
      )
}