package wireguard

const CONFIGMANAGER_URI = "127.0.0.1:8081"
const VPN_USER = "vpn"
const VPN_INTERFACE_NAME = "vpn"
const DEFAULT_VPN_PREFIX = "10.189.184.1/21"
const VPN_CONFIG_NAME = "vpn-config.json"
const IP_LIST_PATH = "config/iplist.json"
const VPN_CLIENTS_DIR = "clients"
const VPN_SERVER_SECRETS_PATH = "secrets"
const VPN_PRIVATE_KEY_FILENAME = "priv.key"
const PRESHARED_KEY_FILENAME = "preshared.key"
const WIREGUARD_TEMPLATE_DIR = "templates"
const WIREGUARD_TEMPLATE_SERVER = "server.tmpl"
const WIREGUARD_CONFIG = "/etc/wireguard/vpn.conf"
const DEFAULT_CLIENT_TEMPLATE = `# default wireguard client template
[Interface]
Address = {{ .Address }}
PrivateKey = {{ .PrivateKey }}
DNS = {{ .DNS }}

[Peer]
PublicKey = {{ .ServerPublicKey }}
PresharedKey = {{ .PresharedKey }}
Endpoint = {{ .Endpoint }}
AllowedIPs = {{StringsJoin .AllowedIPs "," }}

PersistentKeepalive = 25
`

const DEFAULT_SERVER_TEMPLATE = `# default wireguard server template
[Interface]
Address = {{ .Address }}
PrivateKey = {{ .PrivateKey }}
ListenPort = {{ .Port }}
{{if not .DisableNAT }}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .ExternalInterface }} -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .ExternalInterface }} -j MASQUERADE
{{end}}
`

// config notify actions
const ACTION_ADD = "add"
const ACTION_DELETE = "delete"
const ACTION_CLEANUP = "cleanup"
