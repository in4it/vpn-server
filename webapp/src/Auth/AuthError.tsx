import { Alert } from "@mantine/core";
import { TbInfoCircle } from "react-icons/tb";
import { useSearchParams } from "react-router-dom";

export function AuthError() {
    const alertIcon = <TbInfoCircle />;

    let [ searchParams, _ ] = useSearchParams();
    if(!searchParams.has("error")) return ''
    return (
        <Alert style={{marginTop: 20}} variant="light" color="red" title="Error" icon={alertIcon}>{searchParams.has("error_description") ? "An error occured: "+searchParams.get("error") +": " + searchParams.get("error_description"): "An error occured: "+searchParams.get("error")}</Alert>
    )
}