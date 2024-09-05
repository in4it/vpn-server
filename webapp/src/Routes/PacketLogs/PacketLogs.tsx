import { Card, Container, Text, Table, Title, Button, Grid, Select, MultiSelect, Popover} from "@mantine/core";
import { AppSettings } from "../../Constants/Constants";
import {  useQuery } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { TbSettings } from "react-icons/tb";
import { DatePickerInput } from "@mantine/dates";
import { useState } from "react";

type LogsDataResponse = {
    enabled: boolean;
    logData: LogData;
    logTypes: string[];
    users: UserMap;
}
type LogData = {
    schema: LogDataSchema;
    rows: LogRow[];
}
type LogDataSchema = {
    columns: string[];
}
type LogRow = {
    t: string;
    d: string[];
}
type UserMap = {
  [key: string]: string;
}

export function PacketLogs() {
    const {authInfo} = useAuthContext();
    const [currentQueryParameters, setSearchParams] = useSearchParams();
    const dateParam = currentQueryParameters.get("date")
    const userParam = currentQueryParameters.get("user")
    const [logType, setLogType] = useState<string[]>([])
    const [logsDate, setLogsDate] = useState<Date | null>(dateParam === null ? new Date() : new Date(dateParam));
    const [user, setUser] = useState<string>(userParam === null ? "all" : userParam)
    const { isPending, error, data } = useQuery<LogsDataResponse>({
      queryKey: ['packetlogs', user, logsDate, logType],
      queryFn: () =>
        fetch(AppSettings.url + '/stats/packetlogs/'+(user === undefined ? "all" : user)+'/'+(logsDate == undefined ? new Date().toISOString().slice(0, 10) : logsDate.toISOString().slice(0, 10)) + "?logtype="+encodeURIComponent(logType.join(",")), {
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
    if(error) return 'A backend error has occurred: ' + error.message

    if(!data.enabled || data.logTypes.length == 0) { // show disabled page if not enabled
      return (
        <Container my={40}>
          <Title ta="center" style={{marginBottom: 20}}>
            Packet Logs
          </Title>
          <Card withBorder radius="md" padding="xl" bg="var(--mantine-color-body)">
            <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
              { !data.enabled ? 
                "Packet Logs are not activated. Activate packet logging in the VPN Settings." 
              : 
                data.logTypes.length == 0 ? "Packet logs are activated, but no packet logging types are selected. Select at least one packet log type." : null
              }
      
            </Text>
            <Card.Section inheritPadding mt="sm" pb="md">
              <Link to="/setup/vpn">
                <Button leftSection={<TbSettings size={14} />} fz="sm" mt="md" radius="md" variant="default" size="sm">
                  VPN Settings
                </Button>
              </Link>
            </Card.Section>
          </Card>
        </Container>
      )
    }

    const rows = data.logData.rows.map((row, i) => (
        <Table.Tr key={i}>
          <Table.Td>{row.t}</Table.Td>
          {row.d.map((element, y) => {
            return (
            <Table.Td key={i+"-"+y}>{element}</Table.Td>
            )
          })}
        </Table.Tr>
      ));
    return (
        <Container my={40} size="80rem">
          <Title ta="center" style={{marginBottom: 20}}>
          Packet Logs
          </Title>
          <Grid>
            <Grid.Col span={4}></Grid.Col>
            <Grid.Col span={4}>
                <DatePickerInput
                value={logsDate}
                onChange={setLogsDate}
                size="xs"
                />
                </Grid.Col>
            <Grid.Col span={2}>
            <Select
                data={Object.keys(data.users).map((key) => {
                  return {
                    label:  data.users[key],
                    value: key,
                  }
                })}
                size="xs"
                withCheckIcon={false}
                value={user}
                onChange={(_value) => setUser(_value === null ? "" : _value)}
                placeholder="User"
                />
            </Grid.Col>
            <Grid.Col span={2}>
              <Popover width={300} position="bottom" withArrow shadow="md">
                <Popover.Target>
                  <Button variant="default" size="xs">Filter</Button>
                </Popover.Target>
                <Popover.Dropdown>
                <MultiSelect
                  label="Protocol"
                  searchable
                  hidePickedOptions
                  comboboxProps={{ offset: 0, withinPortal: false}}
                  data={data.logTypes}
                  value={logType}
                  onChange={setLogType}
                  size="xs"
                  placeholder="Log Type"
                      />
                </Popover.Dropdown>
              </Popover>           
            </Grid.Col>
          </Grid>
          <Table>
              <Table.Thead>
                  <Table.Tr key="heading">
                  <Table.Th>Timestamp</Table.Th>
                  <Table.Th>Protocol</Table.Th>
                  <Table.Th>Source IP</Table.Th>
                  <Table.Th>Dest. IP</Table.Th>
                  <Table.Th>Source Port</Table.Th>
                  <Table.Th>Dest. Port</Table.Th>
                  <Table.Th>Destination</Table.Th>
                  </Table.Tr>
              </Table.Thead>
              <Table.Tbody>{rows}</Table.Tbody>
          </Table>
        </Container>

    )
}