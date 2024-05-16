import { useEffect } from "react";
import { useCookies } from "react-cookie"
import { useAuthContext } from "./Auth";

export function Logout() {
    const {setAuthInfo} = useAuthContext()

    const [_, setCookie] = useCookies(['token']);
    useEffect(() => {
        setCookie("token", "", {path: "/"})
        setAuthInfo({login: "", role: "", token:"", userType: ""})
        window.history.replaceState(null, "VPN Server", "/")
        location.reload();
    }, []);
    return (<></>)
}