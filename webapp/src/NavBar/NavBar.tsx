import { useState } from 'react';
import { Group, Code } from '@mantine/core';
import {
  IconBellRinging,
  IconSettings,
  IconLogout,
  IconUser,
  IconPlugConnected,
  IconCloudDataConnection,
  IconBook,
  IconUserCircle,
} from '@tabler/icons-react';
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
    { link: '/', label: 'Status', icon: IconBellRinging },
    { link: '/connection', label: 'VPN Connections', icon: IconPlugConnected },
    { link: '/users', label: 'Users', icon: IconUser },
    { link: '/setup', label: 'VPN Setup', icon: IconSettings },
    { link: '/auth-setup', label: 'Authentication & Provisioning', icon: IconCloudDataConnection },
    { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: IconBook },
  ] : 
  [
    { link: '/connection', label: 'VPN Connections', icon: IconPlugConnected },
    { link: 'https://vpn-documentation.in4it.com', label: 'Documentation', icon: IconBook },
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
      <item.icon className={classes.linkIcon} stroke={1.5} />
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
            <IconUserCircle className={classes.linkIcon} stroke={1.5} />
            <span>Profile</span>
          </NavLink>
          :
          null
        }
        <NavLink to="/logout" className={classes.link}>
          <IconLogout className={classes.linkIcon} stroke={1.5} />
          <span>Logout</span>
        </NavLink>
      </div>
    </nav>
  );
}