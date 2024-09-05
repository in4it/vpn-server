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

export function NavBar() {
  const {authInfo} = useAuthContext();
  const location = useLocation();
  const { pathname } = location;
  const [active, setActive] = useState(pathname);

  const data = authInfo.role === "admin" ? [
    { link: '/', label: 'Status', icon: TbBellRinging },
    { link: '/connection', label: 'VPN Connections', icon: TbPlugConnected },
    { link: '/users', label: 'Users', icon: TbUser },
    { link: '/setup', label: 'VPN Setup', icon: TbSettings },
    { link: '/auth-setup', label: 'Authentication & Provisioning', icon: TbCloudDataConnection },
    { link: '/packetlogs', label: 'Logging', icon: FaStream },
    { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: TbBook },
  ] : 
  [
    { link: '/connection', label: 'VPN Connections', icon: TbPlugConnected },
    { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: TbBook },
  ];

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
          VPN Server
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