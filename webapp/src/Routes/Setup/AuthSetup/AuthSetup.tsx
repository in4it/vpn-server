import { Button, Container, Tabs, Title, rem } from "@mantine/core";
import classes from '../Setup.module.css';
import { useState } from "react";
import { NewOIDC } from "./NewOIDC";
import { ListOIDCProviders } from "./ListOIDCProviders";
import { Provisioning } from "./Provisioning";
import { TbDevices, TbIdBadge, TbUserCircle } from "react-icons/tb";
import { ListSAMLProviders } from "./ListSAMLProviders";
import { NewSAML } from "./NewSAML";

export function AuthSetup() {
    const [showNewOIDCProvider, setShowNewOIDCProvider] = useState<boolean>()
    const [showNewSAMLProvider, setShowNewSAMLProvider] = useState<boolean>()
    const iconStyle = { width: rem(12), height: rem(12) };
    return (
        <Container my={40}>

          <Title ta="center" className={classes.title} style={{marginBottom: 20}}>
            Authentication & Provisioning
          </Title>
          <Tabs defaultValue="oidc">
            <Tabs.List grow={true}>
              <Tabs.Tab value="oidc" leftSection={<TbIdBadge style={iconStyle} />}>
                OpenID Connect (OIDC) Connections
              </Tabs.Tab>
              <Tabs.Tab value="saml" leftSection={<TbUserCircle style={iconStyle} />}>
                SAML
              </Tabs.Tab>
              <Tabs.Tab value="provisioning" leftSection={<TbDevices style={iconStyle} />}>
                Provisioning
              </Tabs.Tab>
            </Tabs.List>
            <Tabs.Panel value="oidc" style={{marginTop: 25}}>
              {showNewOIDCProvider ?
                <NewOIDC setShowNewOIDCProvider={setShowNewOIDCProvider} />
                :
                <>
                  <ListOIDCProviders />
                  <Button onClick={() => setShowNewOIDCProvider(true)}>New OIDC Connection</Button>
                </>
              }
            </Tabs.Panel>
            <Tabs.Panel value="saml" style={{marginTop: 25}}>
              {showNewSAMLProvider ?
                <NewSAML setShowNewSAMLProvider={setShowNewSAMLProvider} />
                :
                <>
                  <ListSAMLProviders />
                  <Button onClick={() => setShowNewSAMLProvider(true)}>New SAML Connection</Button>
                </>
              }
            </Tabs.Panel>

            <Tabs.Panel value="provisioning" style={{marginTop: 25}}>          
              <Provisioning />
            </Tabs.Panel>
          </Tabs>
        </Container>

    )
}