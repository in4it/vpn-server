import { Card, Center, Divider, Grid, Select, Text } from "@mantine/core";
import { DatePickerInput } from '@mantine/dates';
import { useQuery } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { AppSettings } from '../../Constants/Constants';
import { format } from "date-fns";
import { Chart } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import { Chart as ChartJS, LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale, ChartOptions, Legend } from 'chart.js';
import { useState } from "react";
  

export function UserStats() {
    ChartJS.register(LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale, Legend);

    const {authInfo} = useAuthContext()
    const [statsDate, setStatsDate] = useState<Date | null>(new Date());
    const [unit, setUnit] = useState<string>("MB")
    const { isPending, error, data } = useQuery({
        queryKey: ['userstats', statsDate, unit],
        queryFn: () =>
            fetch(AppSettings.url + '/stats/user/' + format(statsDate === null ? new Date() : statsDate, "yyyy-MM-dd") + "?unit=" +unit, {
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
            position: 'right' as const,
            display: true,
          },
        },
        scales: {
            x: {
                type: 'time',
                min: '00:00:00',
                /*time: {
                    displayFormats: {
                        quarter: 'HHHH MM'
                    }
                }*/
            },
            y: {
                min: 0
            }
        }
    }

    if (isPending) return ''
    if (error) return 'cannot retrieve licensed users'
    

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
        </>
    )
}