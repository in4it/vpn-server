import { Container, Title, Text, Group, Button } from "@mantine/core";
import { useAuthContext } from "./Auth";
import classes from './CheckRole.module.css';
import { Link } from "react-router-dom";


type Props = {
  children?: React.ReactNode
  role: string
};


export const CheckRole: React.FC<Props> = ({children, role}) => {
    const {authInfo} = useAuthContext()
    if(role === authInfo.role) {
        return (
            children
          )
    } else {
        return (
            <Container className={classes.root}>
                <div className={classes.label}>403</div>
                <Title className={classes.title}>You have found a secret place.</Title>
                <Text c="dimmed" size="lg" ta="center" className={classes.description}>
                Unfortunately, you don't have access to this page.
                </Text>
                <Group justify="center">
                <Button variant="subtle" size="md">
                <Link to="/">Take me back to home page</Link>
                </Button>
                </Group>
            </Container>
        )
    }

 }