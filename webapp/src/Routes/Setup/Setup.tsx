import { Container, Tabs, Title, rem } from "@mantine/core";
import classes from './Setup.module.css';
import { IconFile, IconNetwork, IconRestore, IconRewindBackward10, IconSettings } from "@tabler/icons-react";
import { GeneralSetup } from "./GeneralSetup";
import { VPNSetup } from "./VPNSetup";
import { TemplateSetup } from "./TemplateSetup";
import { Restart } from "./Restart";

export function Setup() {
  const iconStyle = { width: rem(12), height: rem(12) };
  return (
      <Container my={40}>

        <Title ta="center" className={classes.title} style={{marginBottom: 20}}>
          VPN Setup
        </Title>
        <Tabs defaultValue="general">
          <Tabs.List grow={true}>
            <Tabs.Tab value="general" leftSection={<IconSettings style={iconStyle} />}>
              General
            </Tabs.Tab>
            <Tabs.Tab value="vpn" leftSection={<IconNetwork style={iconStyle} />}>
              VPN
            </Tabs.Tab>
            <Tabs.Tab value="templates" leftSection={<IconFile style={iconStyle} />}>
              Templates
            </Tabs.Tab>
            <Tabs.Tab value="restart" leftSection={<IconRestore style={iconStyle} />}>
              Restart
            </Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="general" style={{marginTop: 25}}>
            <GeneralSetup />
          </Tabs.Panel>
          <Tabs.Panel value="vpn" style={{marginTop: 25}}>
            <VPNSetup />
          </Tabs.Panel>
          <Tabs.Panel value="templates" style={{marginTop: 25}}>          
            <TemplateSetup />
          </Tabs.Panel>
          <Tabs.Panel value="restart" style={{marginTop: 25}}>          
            <Restart />
          </Tabs.Panel>
        </Tabs>
      </Container>

  )
}