import { useState } from "react";
import { Link } from "react-router";
import { AppSettings } from '../../Constants/Constants';
import { useAuthContext } from "../../Auth/Auth";
import { httpErrorFromResponse } from "../../Auth/fetchJSON";
import { ErrorAlert } from "../../ErrorAlert";

type Props = {
    id: string;
    name: string;
  };

export function Download({id, name}:Props) {
    const {authInfo} = useAuthContext();
    const [downloadError, setDownloadError] = useState("")
    const handleDownload = async () => {
        setDownloadError("")
        try {
          const response = await fetch(AppSettings.url + '/vpn/connection/'+id, {
              headers: {
                "Authorization": "Bearer " + authInfo.token
              },
          })
          // don't save an error body (e.g. {"error": "..."}) as the config file
          if (!response.ok) {
            const error = await httpErrorFromResponse(response)
            setDownloadError("Could not download the config: " + error.message)
            return
          }
          const blob = await response.blob()
          const url = window.URL.createObjectURL(new Blob([blob]));
          const link = document.createElement("a");
          link.href = url;
          link.download = name+".conf";
          document.body.appendChild(link);

          link.click();

          document.body.removeChild(link);
          window.URL.revokeObjectURL(url);
        } catch (error) {
          setDownloadError("Could not download the config: " + (error instanceof Error ? error.message : String(error)))
        }
    }
    return (
        <>
        <Link to={"?"+id} onClick={handleDownload}>Download New Config</Link>
        <ErrorAlert message={downloadError} />
        </>
    )
}
