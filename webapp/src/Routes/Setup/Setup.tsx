import { Container, Tabs, Title, rem } from "@mantine/core";
import classes from './Setup.module.css';
import { TbFile, TbNetwork, TbRestore, TbSettings } from "react-icons/tb";
import { GeneralSetup } from "./GeneralSetup";
import { VPNSetup } from "./VPNSetup";
import { TemplateSetup } from "./TemplateSetup";
import { Restart } from "./Restart";
import { useParams } from "react-router-dom";

export function Setup() {
  const iconStyle = { width: rem(12), height: rem(12) };
  let { page } = useParams();
  return (
      <Container my={40}>

        <Title ta="center" className={classes.title} style={{marginBottom: 20}}>
          VPN Setup
        </Title>
        <Tabs defaultValue={page == undefined ? "general" : page}>
          <Tabs.List grow={true}>
            <Tabs.Tab value="general" leftSection={<TbSettings style={iconStyle} />}>
              General
            </Tabs.Tab>
            <Tabs.Tab value="vpn" leftSection={<TbNetwork style={iconStyle} />}>
              VPN
            </Tabs.Tab>
            <Tabs.Tab value="templates" leftSection={<TbFile style={iconStyle} />}>
              Templates
            </Tabs.Tab>
            <Tabs.Tab value="restart" leftSection={<TbRestore style={iconStyle} />}>
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