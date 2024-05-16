import { Container, Title, Text } from "@mantine/core";
import classes from './NewConnection.module.css';
import { ListConnections } from "./ListConnections";
import { NewConnection } from "./NewConnection";
import { useAuthContext } from "../../Auth/Auth";
export function Connections() {
    const {authInfo} = useAuthContext();

    return (
        <Container my={40}>
            <Title ta="center" className={classes.title}>
            New connection
            </Title>

            <h2>VPN Connections</h2>
            {authInfo.login === "admin" ? 
            <p>The admin user cannot create new connections. Login with another user first.</p>
            :
            <>
                <Text fz="sm" style={{marginBottom: 5}}>
                Create a new connection per device you want to use the VPN on. Download the configuration below, and use it with a <a href="https://www.wireguard.com/install/" target="_blank">WireGuard® client</a>.
                </Text>
                <ListConnections />
                <NewConnection />
            </> 
            }
        </Container>
    )
}
