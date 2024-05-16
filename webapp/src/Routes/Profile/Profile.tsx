import { Button, Container, Title, Text } from "@mantine/core";
import { useAuthContext } from "../../Auth/Auth";
import classes from './Profile.module.css';
import { ChangePassword } from "./ChangePassword";
import { ListFactors } from "./ListFactors";
import { useState } from "react";
import { NewFactor } from "./NewFactor";
import base32Encode from "base32-encode";

export function Profile() {
    const {authInfo} = useAuthContext();
    const [showNewFactor, setShowNewFactor] = useState<boolean>()

    const secret = base32Encode(window.crypto.getRandomValues(new Uint8Array(160 / 8)), 'RFC4648', { padding: false })
    
    if(showNewFactor) {
        return <NewFactor setShowNewFactor={setShowNewFactor} secret={secret} />
    }

    return (
        <Container my={40} size="40rem">
          <Title ta="center" className={classes.title}>
            Profile
          </Title>
    
          {authInfo.userType == "local" ?
            <>
              <ChangePassword />
              <ListFactors />
              <Button onClick={() => setShowNewFactor(true)}>New Factor (MFA)</Button>
            </>
            :
            <Text>No profile information available for OIDC users.</Text>
          }

        </Container>

    )
}