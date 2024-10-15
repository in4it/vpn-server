import { Card, Center, Grid, Select, Text } from "@mantine/core";
import { DatePickerInput } from '@mantine/dates';
import { useQuery } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { AppSettings } from '../../Constants/Constants';
import { format } from "date-fns";
import { Chart } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import { Chart as ChartJS, LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale, ChartOptions, Legend, Tooltip } from 'chart.js';
import { useState } from "react";

export function UserStats() {
    ChartJS.register(LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale, Legend, Tooltip);
    const timezoneOffset = new Date().getTimezoneOffset() * -1
    const {authInfo} = useAuthContext()
    const [statsDate, setStatsDate] = useState<Date | null>(new Date());
    const [unit, setUnit] = useState<string>("MB")
    const { isPending, error, data } = useQuery({
        queryKey: ['userstats', statsDate, unit],
        queryFn: () =>
            fetch(AppSettings.url + '/vpn/stats/user/' + format(statsDate === null ? new Date() : statsDate, "yyyy-MM-dd") + "?offset="+timezoneOffset+"&unit=" +unit, {
            headers: {
                "Content-Type": "application/json",
                "Authorization": "Bearer " + authInfo.token
            },
            }).then((res) => {
            return res.json()
            }
            
            ),
            enabled: authInfo.role === "admin",
    })
    
    const options:ChartOptions<"line"> = {
        responsive: true,
        plugins: {
          legend: {
            position: 'bottom' as const,
            display: true,
          },
          tooltip: {
            callbacks: {
              //title: (xDatapoint) => {return "this is the data: " + xDatapoint.},
              label: (yDatapoint) => {return " "+yDatapoint.formattedValue + " " + unit},
            }
          }      
        },
        scales: {
            x: {
                type: 'time',
            },
            y: {
                min: 0
            }
        },
  
         hover: {
            mode: 'index',
            intersect: false
         }      
    }

    if (isPending) return ''
    if (error) return 'cannot retrieve licensed users'

    if(data.receivedBytes.datasets === null) {
        data.receivedBytes.datasets = [{ data: [0], label: "no data"}]
    }
    if(data.transmitBytes.datasets === null) {
        data.transmitBytes.datasets = [{ data: [0], label: "no data"}]
    }
    if(data.handshakes.datasets === null) {
        data.handshakes.datasets = [{ data: [0], label: "no data"}]
    }

    return (
        <>
        <Card withBorder radius="md" bg="var(--mantine-color-body)" mt={20}>
            <Grid>
            <Grid.Col span={6}></Grid.Col>

            <Grid.Col span={4}>
                <DatePickerInput
                value={statsDate}
                onChange={setStatsDate}
                size="xs"
                />
                </Grid.Col>
            <Grid.Col span={2}>
            <Select
                data={['Bytes', 'KB', 'MB', 'GB']}
                defaultValue={"MB"}
                allowDeselect={false}
                size="xs"
                withCheckIcon={false}
                value={unit}
                onChange={(_value) => setUnit(_value === null ? "" : _value)}
                />
            </Grid.Col>
            </Grid>

            <Center mt={10}>
            <Text fw={500} size="lg">Data Received by VPN
            </Text>          
            </Center>
            <Chart type="line" data={data.receivedBytes} options={options} />
        </Card>
        <Card withBorder radius="md" bg="var(--mantine-color-body)" mt={20}>
            <Center>
            <Text fw={500} size="lg">Data Sent by VPN</Text>
            </Center>
            <Chart type="line" data={data.transmitBytes} options={options} />
        </Card>
        <Card withBorder radius="md" bg="var(--mantine-color-body)" mt={20}>
            <Center>
            <Text fw={500} size="lg">User Handshakes</Text>
            </Center>
            <Chart type="line" data={data.handshakes} options={{...options, plugins: {...options.plugins, tooltip: { ...options.plugins?.tooltip, callbacks: {label: (yDatapoint) => {return " "+yDatapoint.formattedValue }} }} }} />
        </Card>
        </>
    )
}