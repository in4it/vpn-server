import { Card } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { useAuthContext } from "../../Auth/Auth";
import { AppSettings } from '../../Constants/Constants';
import { Chart } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import { Chart as ChartJS, LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale } from 'chart.js';

export function UserStats() {
    ChartJS.register(LineController, LineElement, PointElement, LinearScale, Title, CategoryScale, TimeScale);

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
    const options = {
        responsive: true,
        plugins: {
          legend: {
            position: 'top' as const,
          },
          title: {
            display: true,
            text: 'VPN Received (in bytes)',
          },
        },
        scales: {
            x: {
                type: 'time',
            }
        }
    }

    if (isPending) return ''
    if (error) return 'cannot retrieve licensed users'
        
    return (
        <Card withBorder radius="md" padding="xl" bg="var(--mantine-color-body)" mt={20}>
        <Chart type="line" data={data.receivedBytes} options={options} />
        </Card>
    )
}