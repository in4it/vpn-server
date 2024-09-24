import { Card, Container, Text, Table, Title, Button, Grid, Select, Popover, Group, TextInput, rem, ActionIcon, Checkbox, Highlight} from "@mantine/core";
import { AppSettings } from "../../Constants/Constants";
import { useInfiniteQuery } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { Link, useSearchParams } from "react-router-dom";
import { TbArrowRight, TbSearch, TbSettings } from "react-icons/tb";
import { DatePickerInput } from "@mantine/dates";
import { useEffect, useState } from "react";
import React from "react";

type LogsDataResponse = {
    enabled: boolean;
    logEntries: LogEntry[];
    environments: string[];
    nextPos: number;
    keys: Keys[];
}
type LogEntry = {
    data: string;
    timestamp: string;
}
type Keys = {
  key: string;
  value: string;
  total: number;
}
type Tag = {
  key: string;
  value: string;
}

function getDate(date:Date) {
  var dd = String(date.getDate()).padStart(2, '0');
  var mm = String(date.getMonth() + 1).padStart(2, '0'); //January is 0!
  var yyyy = date.getFullYear();  
  return yyyy + "-" + mm + '-' + dd;
}

export function Logs() {
    const {authInfo} = useAuthContext();
    const timezoneOffset = new Date().getTimezoneOffset() * -1
    const [currentQueryParameters] = useSearchParams();
    const dateParam = currentQueryParameters.get("date")
    const environmentParam = currentQueryParameters.get("environment")
    const [tags, setTags] = useState<Tag[]>([])
    const [search, setSearch] = useState<string>("")
    const [searchParam, setSearchParam] = useState<string>("")
    const [logsDate, setLogsDate] = useState<Date | null>(dateParam === null ? new Date() : new Date(dateParam));
    const [environment, setEnvironment] = useState<string>(environmentParam === null ? "all" : environmentParam)
    const { isPending, fetchNextPage, hasNextPage, error, data } = useInfiniteQuery<LogsDataResponse>({
      queryKey: ['logs', environment, logsDate, tags, searchParam],
      queryFn: async ({ pageParam }) =>
        fetch(AppSettings.url + '/observability/logs?environment='+(environment === undefined || environment === "" ? "all" : environment)+'&fromDate='+(logsDate == undefined ? getDate(new Date()) : getDate(logsDate)) + '&endDate='+(logsDate == undefined ? getDate(new Date()) : getDate(logsDate)) + "&pos="+pageParam+"&offset="+timezoneOffset+"&tags="+encodeURIComponent(tags.map(t => t.key + "=" + t.value).join(","))+"&search="+encodeURIComponent(searchParam), {
          headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + authInfo.token
          },
        }).then((res) => {
          return res.json()
          }
        ),
        initialPageParam: 0,
        getNextPageParam: (lastRequest) => lastRequest.nextPos === -1 ? null : lastRequest.nextPos,
    })

    const captureEnter = (e: React.KeyboardEvent<HTMLDivElement>) => {
      if (e.key === "Enter") {
        setSearchParam(search)
      }
    }

    useEffect(() => {
      const handleScroll = () => {
        const { scrollTop, clientHeight, scrollHeight } =
          document.documentElement;
        if (scrollTop + clientHeight >= scrollHeight - 20) {
          fetchNextPage();
        }
      };
  
      window.addEventListener("scroll", handleScroll);
      return () => {
        window.removeEventListener("scroll", handleScroll);
      };
    }, [fetchNextPage])

    if(isPending) return "Loading..."
    if(error) return 'A backend error has occurred: ' + error.message

    if(data.pages.length === 0 || !data.pages[0].enabled) { // show disabled page if not enabled
      return (
        <Container my={40}>
          <Title ta="center" style={{marginBottom: 20}}>
            Logs
          </Title>
          <Card withBorder radius="md" padding="xl" bg="var(--mantine-color-body)">
            <Text fz="xs" tt="uppercase" fw={700} c="dimmed">
              { !data.pages[0].enabled ? 
                "Logs are not enabled." 
              : 
                null
              }
            </Text>
            <Card.Section inheritPadding mt="sm" pb="md">
              <Link to="/setup/vpn">
                <Button leftSection={<TbSettings size={14} />} fz="sm" mt="md" radius="md" variant="default" size="sm">
                  Logs Settings
                </Button>
              </Link>
            </Card.Section>
          </Card>
        </Container>
      )
    }

    const rows = data.pages.map((group, groupIndex) => (
      <React.Fragment key={groupIndex}>
        {group.logEntries.map((row, i) => (
          <Table.Tr key={i}>
            <Table.Td>{row.timestamp}</Table.Td>
            <Table.Td>{searchParam === "" ? row.data : <Highlight color="lime" highlight={searchParam}>{row.data}</Highlight>}</Table.Td>
          </Table.Tr>
        ))}
      </React.Fragment>
      ));
    return (
        <Container my={40} size="80rem">
          <Title ta="center" style={{marginBottom: 20}}>
          Logs
          </Title>
          <Grid>
            <Grid.Col span={4}>
            <TextInput
                placeholder="Search..."
                rightSectionWidth={30}
                size="xs"
                leftSection={<TbSearch style={{ width: rem(18), height: rem(18) }} />}
                rightSection={
                  <ActionIcon size={18} radius="xl" variant="filled" onClick={() => setSearchParam(search)}>
                    <TbArrowRight style={{ width: rem(14), height: rem(14) }} />
                  </ActionIcon>
                }
                onKeyDown={(e) => captureEnter(e)}
                onChange={(e) => setSearch(e.currentTarget.value)}
                value={search}
              />
            </Grid.Col>
            <Grid.Col span={4}>
                <DatePickerInput
                value={logsDate}
                onChange={setLogsDate}
                size="xs"
                />
                </Grid.Col>
            <Grid.Col span={2}>
            <Select
                data={data.pages[0].environments.map((key, index) => {
                  return {
                    label: key,
                    value: index.toString(),
                  }
                })}
                size="xs"
                withCheckIcon={false}
                value={environment}
                onChange={(_value) => setEnvironment(_value === null ? "" : _value)}
                placeholder="Environment"
                />
            </Grid.Col>
            <Grid.Col span={2}>
              <Popover width={300} position="bottom" withArrow shadow="md">
                <Popover.Target>
                  <Button variant="default" size="xs">Filter</Button>
                </Popover.Target>
                <Popover.Dropdown>
                {data.pages[0].keys.map((element) => {
                  return (
                    <Checkbox
                      key={element.key +"="+element.value}
                      label={element.key + " = " + element.value.substring(0, 10) + (element.value.length > 10 ? "..." : "") + " (" + element.total + ")"}
                      radius="xs"
                      size="xs"
                      style={{marginBottom: 3}}
                      onChange={(event) => event.currentTarget.checked ? setTags([...tags, {key: element.key, value: element.value }]) : setTags(tags.filter((tag) => { return tag.key !== element.key || tag.value !== element.value } ))}
                      checked={tags.some((tag) => tag.key === element.key && tag.value === element.value)}
                    />
                  )
                })}
                </Popover.Dropdown>
              </Popover>           
            </Grid.Col>
          </Grid>
          <Table>
              <Table.Thead>
                  <Table.Tr key="heading">
                  <Table.Th>Timestamp</Table.Th>
                  <Table.Th>Log</Table.Th>
                  </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                {rows}
              </Table.Tbody>
          </Table>
          <Group justify="center">
          {hasNextPage ? <Button onClick={() => fetchNextPage()} variant="default">Loading more...</Button> : null}
          </Group>

        </Container>

    )
}