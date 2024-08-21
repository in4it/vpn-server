import { Container, Title } from "@mantine/core";
import classes from './Users.module.css';
import { useState } from "react";
import { ListUsers } from "./ListUsers";
import { NewUser } from "./NewUser";
import { NewUserButton } from "./NewUserButton";
import { useAuthContext } from "../../Auth/Auth";
import { AppSettings } from "../../Constants/Constants";
import { useQuery } from "@tanstack/react-query";

export function Users() {
    const [showNewUser, setShowNewUser] = useState<boolean>()
    const {authInfo} = useAuthContext();
    const { isPending, error, data } = useQuery({
        queryKey: ['setup'],
        queryFn: () =>
          fetch(AppSettings.url + '/setup/general', {
            headers: {
              "Content-Type": "application/json",
              "Authorization": "Bearer " + authInfo.token
            },
          }).then((res) => {
            return res.json()
            }
            
          ),
      })
    if(isPending) return "Loading..."
    if(error) { return "can't load page, error: " + error.message}
    if(showNewUser) {
        return <NewUser setShowNewUser={setShowNewUser} />
    }
    return (
        <Container my={40}>
          <Title ta="center" className={classes.title}>
            Users
          </Title>
    
          <h2>Users</h2>
          <ListUsers localAuthDisabled={data.disableLocalAuth} />
          <NewUserButton setShowNewUser={setShowNewUser} localAuthDisabled={data.disableLocalAuth} />
        </Container>

    )
}