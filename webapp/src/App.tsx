import "@mantine/core/styles.css";
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

const queryClient = new QueryClient()

export default function App() {
  const [opened] = useDisclosure();
  return <MantineProvider theme={theme} forceColorScheme="light">
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <AppInit>
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
                      <NavBar />
                    </AppShell.Navbar>
                    <AppShell.Main>     
                        <Routes>
                          <Route path="/" element={<Home />} />
                          <Route path="/users" element={<CheckRole role="admin"><Users /></CheckRole>} />
                          <Route path="/setup" element={<CheckRole role="admin"><Setup /></CheckRole>} />
                          <Route path="/auth-setup" element={<CheckRole role="admin"><AuthSetup /></CheckRole>} />
                          <Route path="/logout" element={<Logout />} />
                          <Route path="/login/:logintype/:id" element={<Navigate to={"/"} />} />
                          <Route path="/callback/:callbacktype/:id" element={<Navigate to={"/"} />} />
                          <Route path="/connection" element={<Connections />} />
                          <Route path="/profile" element={<Profile />} />
                          <Route path="/upgrade" element={<Upgrade />} />
                          <Route path="/licenses" element={<GetMoreLicenses />} />
                        </Routes>
                    </AppShell.Main>
                  </AppShell>
                </Auth>
              </AppInit>
            </BrowserRouter>
          </QueryClientProvider>
          </MantineProvider>;
}
