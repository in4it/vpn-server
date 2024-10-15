import { Link } from "react-router-dom";
import { AppSettings } from '../../Constants/Constants';
import { useAuthContext } from "../../Auth/Auth";

type Props = {
    id: string;
    name: string;
  };

export function Download({id, name}:Props) {
    const {authInfo} = useAuthContext();
    const handleDownload = () => {
        fetch(AppSettings.url + '/vpn/connection/'+id, {
            headers: {
              "Authorization": "Bearer " + authInfo.token
            },
        })
        .then((response) => response.blob())
        .then((blob) => {
          const url = window.URL.createObjectURL(new Blob([blob]));
          const link = document.createElement("a");
          link.href = url;
          link.download = name+".conf";
          document.body.appendChild(link);
  
          link.click();
  
          document.body.removeChild(link);
          window.URL.revokeObjectURL(url);
        })
        .catch((error) => {
          console.error("Error fetching the config:", error);
        });
  
    }
    return (
        <Link to={"?"+id} onClick={handleDownload}>Download New Config</Link>
    )
}