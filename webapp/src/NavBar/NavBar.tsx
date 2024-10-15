import { useState } from 'react';
import { Group, Code } from '@mantine/core';
import {
  TbBellRinging,
  TbSettings,
  TbLogout,
  TbUser,
  TbPlugConnected,
  TbCloudDataConnection,
  TbBook,
  TbUserCircle,
  
} from 'react-icons/tb';
import { FaStream } from "react-icons/fa";

import classes from './Navbar.module.css';
import { NavLink, useLocation } from 'react-router-dom';
import { useAuthContext } from '../Auth/Auth';
import { Version } from './Version';

type Props = {
  serverType: string
};


export function NavBar({serverType}: Props) {
  const {authInfo} = useAuthContext();
  const location = useLocation();
  const { pathname } = location;
  const [active, setActive] = useState(pathname);

  const vpnLinks = {
    "admin": [
      { link: '/', label: 'Status', icon: TbBellRinging },
      { link: '/connection', label: 'VPN Connections', icon: TbPlugConnected },
      { link: '/users', label: 'Users', icon: TbUser },
      { link: '/setup', label: 'VPN Setup', icon: TbSettings },
      { link: '/auth-setup', label: 'Authentication & Provisioning', icon: TbCloudDataConnection },
      { link: '/packetlogs', label: 'Logging', icon: FaStream },
      { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: TbBook },
    ],
    "user": [
      { link: '/connection', label: 'VPN Connections', icon: TbPlugConnected },
      { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: TbBook },
    ]
  }
  const observabilityLinks = {
    "admin": [
      { link: '/', label: 'Status', icon: TbBellRinging },
      { link: '/logs', label: 'Logs', icon: FaStream },
      { link: '/users', label: 'Users', icon: TbUser },
      { link: '/setup', label: 'Setup', icon: TbSettings },
      { link: '/auth-setup', label: 'Authentication & Provisioning', icon: TbCloudDataConnection },
      { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: TbBook },
    ],
    "user": [
      { link: '/logs', label: 'Logs', icon: FaStream },
      { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: TbBook },
    ]
  }

  const getData = () => {
    if(serverType === "vpn") {
      if (authInfo.role === "admin" ) {
        return vpnLinks.admin
      } else {
        return vpnLinks.user
      }
    }
    if(serverType === "observability") {
      if (authInfo.role === "admin" ) {
        return observabilityLinks.admin
      } else {
        return observabilityLinks.user
      }
    }
    return []
  }

  const data = getData()

  const links = data.map((item) => (
    <NavLink
      className={classes.link}
      data-active={item.link === active || undefined}
      to={item.link}
      key={item.link}
      target={item.link.startsWith("http") ? "_blank" : ""}
      onClick={() => {
        setActive(item.link);
      }}
    >
      <item.icon className={classes.linkIcon} />
      <span>{item.label}</span>
    </NavLink>
  ));

  return (
    <nav className={classes.navbar}>
      <div className={classes.navbarMain}>
        <Group className={classes.header} justify="space-between">
          {serverType === "vpn" ? "VPN Server" : "Observability Server"}
          <Code fw={700}><Version /></Code>
        </Group>
        {links}
      </div>
      <div className={classes.footer}>
        {authInfo.userType == "local" ?
          <NavLink to="/profile" className={classes.link} onClick={() => { setActive("/profile"); }} data-active={"/profile" === active || undefined}>
            <TbUserCircle className={classes.linkIcon} />
            <span>Profile</span>
          </NavLink>
          :
          null
        }
        <NavLink to="/logout" className={classes.link}>
          <TbLogout className={classes.linkIcon} />
          <span>Logout</span>
        </NavLink>
      </div>
    </nav>
  );
}