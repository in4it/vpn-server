import { useEffect, useState } from "react";
import { AppSettings, UUID_REGEX } from "../Constants/Constants";
import { useMatch } from "react-router-dom";

export function OIDCAndSAMLRedirect() {
  const [redirectError, setRedirectError] = useState<string>("")
  const loginMatch = useMatch("/login/:loginType/:loginID");

  useEffect(() => {
    if (loginMatch && loginMatch.params.loginType !== undefined && loginMatch.params.loginID !== undefined) {
      if (!UUID_REGEX.test(loginMatch.params.loginID)) {
        setRedirectError("Invalid OIDC method id (expected UUID)");
        return;
      }
      window.location.href = AppSettings.url + '/authmethods/' + encodeURIComponent(loginMatch.params.loginType) + '/' + encodeURIComponent(loginMatch.params.loginID) + "/redirect"
    }
  }, []);
  if (redirectError !== "") return <p>{redirectError}</p>
  return (<></>)
}