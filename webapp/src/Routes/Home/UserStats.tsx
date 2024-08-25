import { Card, Center, Text } from "@mantine/core";
//import { DatePicker } from '@mantine/dates';
import { useQuery } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { AppSettings } from '../../Constants/Constants';
import { Chart } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import { Chart as ChartJS, LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale, ChartOptions, Legend } from 'chart.js';

export function UserStats() {
    ChartJS.register(LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale, Legend);

    const {authInfo} = useAuthContext()
    const { isPending, error, data } = useQuery({
        queryKey: ['userstats'],
        queryFn: () =>
            fetch(AppSettings.url + '/stats/user', {
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
    
    if(data.receivedBytes.datasets === null) {
        data.receivedBytes.datasets = [{ data: [0], label: "no data"}]
    }
    if(data.transmitBytes.datasets === null) {
        data.transmitBytes.datasets = [{ data: [0], label: "no data"}]
    }

    return (
        <>
        <Card withBorder radius="md" bg="var(--mantine-color-body)" mt={20}>
            <Center>
            <Text fw={500} size="lg">VPN Data Received (bytes)</Text>
            <Text>
            </Text>
            
            </Center>
            <Chart type="line" data={data.receivedBytes} options={options} />
        </Card>
        <Card withBorder radius="md" bg="var(--mantine-color-body)" mt={20}>
            <Center>
            <Text fw={500} size="lg">VPN Data Sent (bytes)</Text>
            </Center>
            <Chart type="line" data={data.transmitBytes} options={options} />
        </Card>
        </>
    )
}