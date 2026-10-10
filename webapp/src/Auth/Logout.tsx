import { useEffect } from "react";
import { useCookies } from "react-cookie"
import { useAuthContext } from "./Auth";
import { TOKEN_COOKIE, tokenCookieRemoveOptions } from "./cookies";

export function Logout() {
    const {setAuthInfo} = useAuthContext()

    const [, , removeCookie] = useCookies([TOKEN_COOKIE]);
    useEffect(() => {
        removeCookie(TOKEN_COOKIE, tokenCookieRemoveOptions)
        setAuthInfo({login: "", role: "", token:"", userType: ""})
        window.history.replaceState(null, "VPN Server", "/")
        location.reload();
    }, []);
    return (<></>)
}
