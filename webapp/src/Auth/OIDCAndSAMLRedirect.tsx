import { useEffect } from "react";
import { AppSettings } from "../Constants/Constants";
import { useMatch } from "react-router-dom";

export function OIDCAndSAMLRedirect() {
  const loginMatch = useMatch("/login/:loginType/:loginID");

  useEffect(() => {
    if (loginMatch && loginMatch.params.loginType !== undefined && loginMatch.params.loginID !== undefined) {
      window.location.href = AppSettings.url + '/authmethods/' + loginMatch.params.loginType + '/' + loginMatch.params.loginID + "/redirect"
    }
  }, []);
  return (<></>)
}