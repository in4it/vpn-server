import "@mantine/core/styles.css";
import '@mantine/dates/styles.css';
import { AppShell, MantineProvider } from "@mantine/core";
import { theme } from "./theme";
import { NavBar } from "./NavBar/NavBar";
import { AppInit } from "./AppInit/AppInit";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Auth } from "./Auth/Auth";
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Home } from "./Routes/Home/Home";
import { AuthSetup } from "./Routes/Setup/AuthSetup/AuthSetup";
import { Setup } from "./Routes/Setup/Setup";
import { useDisclosure } from "@mantine/hooks";
import { Logout } from "./Auth/Logout";
import { Connections } from "./Routes/Connection/Connections";
import { CheckRole } from "./Auth/CheckRole";
import { Users } from "./Routes/Users/Users";
import { Profile } from "./Routes/Profile/Profile";
import { Upgrade } from "./Routes/Upgrade/Upgrade";
import { GetMoreLicenses } from "./Routes/Licenses/GetMoreLicenses";
import { PacketLogs } from "./Routes/PacketLogs/PacketLogs";
import { Logs } from "./Routes/Logs/Logs";

const queryClient = new QueryClient()

export default function App() {
  const [opened] = useDisclosure();
  return <MantineProvider theme={theme} forceColorScheme="light">
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <AppInit serverType="vpn">
                <Auth>
                  <AppShell
                    navbar={{
                      width: 300,
                      breakpoint: 'sm',
                      collapsed: { mobile: opened },
                    }}
                    padding="md"
                  >
                    <AppShell.Navbar>
                      <NavBar serverType="vpn" />
                    </AppShell.Navbar>
                    <AppShell.Main>     
                        <Routes>
                          <Route path="/" element={<Home />} />
                          <Route path="/users" element={<CheckRole role="admin"><Users /></CheckRole>} />
                          <Route path="/setup" element={<CheckRole role="admin"><Setup /></CheckRole>} />
                          <Route path="/setup/:page" element={<CheckRole role="admin"><Setup /></CheckRole>} />
                          <Route path="/auth-setup" element={<CheckRole role="admin"><AuthSetup /></CheckRole>} />
                          <Route path="/upgrade" element={<CheckRole role="admin"><Upgrade /></CheckRole>} />
                          <Route path="/licenses" element={<CheckRole role="admin"><GetMoreLicenses /></CheckRole>} />
                          <Route path="/packetlogs" element={<CheckRole role="admin"><PacketLogs /></CheckRole>} />
                          <Route path="/logout" element={<Logout />} />
                          <Route path="/login/:logintype/:id" element={<Navigate to={"/"} />} />
                          <Route path="/callback/:callbacktype/:id" element={<Navigate to={"/"} />} />
                          <Route path="/connection" element={<Connections />} />
                          <Route path="/profile" element={<Profile />} />
                        </Routes>
                    </AppShell.Main>
                  </AppShell>
                </Auth>
              </AppInit>
              <AppInit serverType="observability">
                <Auth>
                  <AppShell
                    navbar={{
                      width: 300,
                      breakpoint: 'sm',
                      collapsed: { mobile: opened },
                    }}
                    padding="md"
                  >
                    <AppShell.Navbar>
                      <NavBar serverType="observability" />
                    </AppShell.Navbar>
                    <AppShell.Main>     
                        <Routes>
                          <Route path="/" element={<Home />} />
                          <Route path="/users" element={<CheckRole role="admin"><Users /></CheckRole>} />
                          <Route path="/setup" element={<CheckRole role="admin"><Setup /></CheckRole>} />
                          <Route path="/setup/:page" element={<CheckRole role="admin"><Setup /></CheckRole>} />
                          <Route path="/auth-setup" element={<CheckRole role="admin"><AuthSetup /></CheckRole>} />
                          <Route path="/upgrade" element={<CheckRole role="admin"><Upgrade /></CheckRole>} />
                          <Route path="/licenses" element={<CheckRole role="admin"><GetMoreLicenses /></CheckRole>} />
                          <Route path="/logs" element={<Logs />} />
                          <Route path="/logout" element={<Logout />} />
                          <Route path="/login/:logintype/:id" element={<Navigate to={"/"} />} />
                          <Route path="/callback/:callbacktype/:id" element={<Navigate to={"/"} />} />
                          <Route path="/profile" element={<Profile />} />
                        </Routes>
                    </AppShell.Main>
                  </AppShell>
                </Auth>
              </AppInit>
            </BrowserRouter>
          </QueryClientProvider>
          </MantineProvider>;
}
