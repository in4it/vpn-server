import { Alert, Button, PasswordInput, Space } from "@mantine/core";
import { TbInfoCircle } from "react-icons/tb";
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { useState } from "react";
import { useCookies } from "react-cookie";
import { AppSettings } from "../../Constants/Constants";
import { useAuthContext } from "../../Auth/Auth";
import { getErrorMessage } from "../../Auth/errorMessage";
import { TOKEN_COOKIE, tokenCookieRemoveOptions } from "../../Auth/cookies";

type UpdatePassword = {
    password: string;
    currentPassword: string;
  }

export function ChangePassword() {
    const [currentPassword, setCurrentPassword] = useState<string>("");
    const [newPassword, setNewPassword] = useState<string>("");
    const [newPasswordRepeat, setNewPasswordRepeat] = useState<string>("");
    const [passwordError, setPasswordError] = useState<string>("");
    const [passwordUpdated, setPasswordUpdated] = useState<boolean>();
    const {authInfo, setAuthInfo} = useAuthContext();
    const [, , removeCookie] = useCookies([TOKEN_COOKIE]);

    // Changing the password expires all earlier sessions (their tokens are
    // rejected by the server), so log out and send the user back to the login
    // prompt after a successful change.
    const logoutAndRedirectToLogin = () => {
        removeCookie(TOKEN_COOKIE, tokenCookieRemoveOptions)
        setAuthInfo({login: "", role: "", token: "", userType: ""})
        window.history.replaceState(null, "VPN Server", "/")
        location.reload()
    }

    const changePassword = useMutation({
        mutationFn: (userIDAndPassword:UpdatePassword) => {
          return axios.post(AppSettings.url + '/profile/password', userIDAndPassword, {
            headers: {
                "Authorization": "Bearer " + authInfo.token
            },
          })
        },
        onSuccess: () => {
            setPasswordUpdated(true)
            setCurrentPassword("")
            setNewPassword("")
            setNewPasswordRepeat("")
            setPasswordError("")
            // give the user a moment to see the confirmation, then log out
            setTimeout(logoutAndRedirectToLogin, 1500)
        },
        onError: (error) => {
            // e.g. wrong current password, or the new password is rejected
            setPasswordError(getErrorMessage(error))
        },
      })


    const onClickChangePassword = () => {
        if(currentPassword === "") {
            setPasswordError("current password is empty")
            return
        }
        if(newPassword !== newPasswordRepeat) {
            setPasswordError("passwords don't match")
            return
        }
        if(newPassword === undefined || newPassword === "") {
            setPasswordError("password is empty")
            return
        }
        if(newPassword.length < 6) {
            setPasswordError("password needs to have at least 6 characters (including 1 number and 1 special character)")
            return
        }
        if(!/[ `!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?~]/.test(newPassword)) {
            setPasswordError("password needs to have at least 1 special character (1 number and 1 special character)")
            return
        }
        if(!/\d/.test(newPassword)) {
            setPasswordError("password needs to have at least 1 number (0-9 number and 1 special character)")
            return
        }
        setPasswordError("")
        changePassword.mutate({password: newPassword, currentPassword: currentPassword})
    }
    const alertIcon = <TbInfoCircle />;

    return (
        <>
        <h2>Change Password</h2>
            {passwordUpdated ? <Alert variant="light" color="green" title="Password Updated" icon={alertIcon} style={{ marginBottom: 20 }}>Password updated. You will be logged out and redirected to the login page.</Alert> : null}
            {passwordError !== "" ? <Alert variant="light" color="red" title="Error" icon={alertIcon} style={{ marginBottom: 20 }}>{passwordError}</Alert> : null }
            <Alert variant="light" color="blue" title="You will be logged out" icon={alertIcon} style={{ marginBottom: 20 }}>Changing your password will log you out of this and any other active sessions. You'll need to log in again with your new password.</Alert>
            <PasswordInput placeholder="Current Password" aria-label="Current password" id="your-current-password" autoComplete="current-password" onChange={(event) => setCurrentPassword(event.currentTarget.value)} value={currentPassword} /><Space h="md" />
            <PasswordInput placeholder="New Password" id="your-password" onChange={(event) => setNewPassword(event.currentTarget.value)} value={newPassword} /><Space h="md" />
            <PasswordInput placeholder="Repeat Password" id="your-password-repeat" onChange={(event) => setNewPasswordRepeat(event.currentTarget.value)} value={newPasswordRepeat} /><Space h="md" />
            <Button onClick={() => onClickChangePassword()} loading={changePassword.isPending}>Change Password</Button>
        </>
)
}